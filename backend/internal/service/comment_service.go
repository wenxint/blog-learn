package service

import (
	"fmt"
	"net/mail"
	"strings"

	"blog/internal/model"
	"blog/internal/pkg/ratelimit"
	"blog/internal/repository"
)

// 评论状态
const (
	CommentPending  int8 = 0 // 待审核
	CommentApproved int8 = 1 // 已通过
)

const (
	maxNicknameLen = 50
	maxContentLen  = 1000
)

// CommentService 评论业务
type CommentService struct {
	comments *repository.CommentRepository
	articles *repository.ArticleRepository
	limiter  *ratelimit.Limiter
}

func NewCommentService(
	comments *repository.CommentRepository,
	articles *repository.ArticleRepository,
	limiter *ratelimit.Limiter,
) *CommentService {
	return &CommentService{comments: comments, articles: articles, limiter: limiter}
}

// CommentNode 一条评论 + 它的回复（只有两级：评论、回复，不做无限嵌套）
type CommentNode struct {
	model.Comment
	Replies []model.Comment `json:"replies,omitempty"`
}

// CreateCommentInput 发表评论的入参
type CreateCommentInput struct {
	Nickname string
	Email    string
	Content  string
	ParentID uint64
	IP       string
}

// ListTree 组装评论树。
//
// 两种做法的对比：
//  1. 递归查库：拿到一级评论，再对每条查它的回复 —— N+1，评论多了会很慢
//  2. 一次查全，内存里组装 —— 只需要 1 次查询，本实现用这种
//
// 因为限定只有两级，内存组装非常简单：先按 parent_id 分桶，再挂到父节点上。
func (s *CommentService) ListTree(articleID uint64) ([]CommentNode, error) {
	// 先确认文章存在且已发布，否则草稿的评论不该被看到
	a, err := s.articles.FindByID(articleID)
	if err != nil {
		return nil, ErrNotFound
	}
	if a.Status != StatusPublished {
		return nil, ErrNotFound
	}

	all, err := s.comments.ListByArticle(articleID, CommentApproved)
	if err != nil {
		return nil, err
	}

	// 按 parent_id 分桶
	childMap := make(map[uint64][]model.Comment)
	for _, c := range all {
		if c.ParentID > 0 {
			childMap[c.ParentID] = append(childMap[c.ParentID], c)
		}
	}

	nodes := make([]CommentNode, 0)
	for _, c := range all {
		if c.ParentID != 0 {
			continue // 回复不单独出现，挂在父评论下面
		}
		nodes = append(nodes, CommentNode{Comment: c, Replies: childMap[c.ID]})
	}
	return nodes, nil
}

// Create 发表评论。默认进入「待审核」状态，管理员通过后才在前台展示。
func (s *CommentService) Create(articleID uint64, in CreateCommentInput) (*model.Comment, error) {
	// 限流放在最前面：连参数都不用校验，直接挡住刷接口的脚本
	if s.limiter != nil {
		if ok, retryAfter := s.limiter.Allow(in.IP); !ok {
			return nil, fmt.Errorf("%w: 发表太频繁，请 %d 秒后再试",
				ErrTooManyReqs, int(retryAfter.Seconds())+1)
		}
	}

	in.Nickname = strings.TrimSpace(in.Nickname)
	in.Content = strings.TrimSpace(in.Content)
	in.Email = strings.TrimSpace(in.Email)

	if in.Nickname == "" {
		return nil, fmt.Errorf("%w: 昵称不能为空", ErrInvalidInput)
	}
	if len([]rune(in.Nickname)) > maxNicknameLen {
		return nil, fmt.Errorf("%w: 昵称不能超过 %d 个字", ErrInvalidInput, maxNicknameLen)
	}
	if in.Content == "" {
		return nil, fmt.Errorf("%w: 评论内容不能为空", ErrInvalidInput)
	}
	if len([]rune(in.Content)) > maxContentLen {
		return nil, fmt.Errorf("%w: 评论内容不能超过 %d 个字", ErrInvalidInput, maxContentLen)
	}
	if in.Email != "" {
		if _, err := mail.ParseAddress(in.Email); err != nil {
			return nil, fmt.Errorf("%w: 邮箱格式不正确", ErrInvalidInput)
		}
	}

	// 文章必须存在且已发布
	a, err := s.articles.FindByID(articleID)
	if err != nil {
		return nil, ErrNotFound
	}
	if a.Status != StatusPublished {
		return nil, ErrNotFound
	}

	// 回复的校验：父评论必须存在、属于同一篇文章、且本身是一级评论（最多两级）
	if in.ParentID > 0 {
		parent, err := s.comments.FindByID(in.ParentID)
		if err != nil {
			return nil, fmt.Errorf("%w: 被回复的评论不存在", ErrInvalidInput)
		}
		if parent.ArticleID != articleID {
			return nil, fmt.Errorf("%w: 被回复的评论不属于这篇文章", ErrInvalidInput)
		}
		if parent.ParentID != 0 {
			return nil, fmt.Errorf("%w: 只支持两级评论，不能回复回复", ErrInvalidInput)
		}
		if parent.Status != CommentApproved {
			return nil, fmt.Errorf("%w: 被回复的评论尚未通过审核", ErrInvalidInput)
		}
	}

	c := &model.Comment{
		ArticleID: articleID,
		ParentID:  in.ParentID,
		Nickname:  in.Nickname,
		Email:     in.Email,
		Content:   in.Content,
		Status:    CommentPending,
		IP:        in.IP,
	}
	if err := s.comments.Create(c); err != nil {
		return nil, err
	}
	return c, nil
}

// ListPending 后台待审核列表
func (s *CommentService) ListPending(page, pageSize int) ([]model.Comment, int64, error) {
	return s.comments.ListByStatus(CommentPending, page, pageSize)
}

// ListByStatus 后台按状态查询
func (s *CommentService) ListByStatus(status int8, page, pageSize int) ([]model.Comment, int64, error) {
	if status != CommentPending && status != CommentApproved {
		return nil, 0, fmt.Errorf("%w: status 只能是 0 或 1", ErrInvalidInput)
	}
	return s.comments.ListByStatus(status, page, pageSize)
}

// SetStatus 审核：通过(1) 或 打回待审(0)
func (s *CommentService) SetStatus(id uint64, status int8) error {
	if status != CommentPending && status != CommentApproved {
		return fmt.Errorf("%w: status 只能是 0 或 1", ErrInvalidInput)
	}
	if _, err := s.comments.FindByID(id); err != nil {
		return ErrNotFound
	}
	return s.comments.UpdateStatus(id, status)
}

// Delete 删除评论（连同回复）
func (s *CommentService) Delete(id uint64) error {
	if _, err := s.comments.FindByID(id); err != nil {
		return ErrNotFound
	}
	return s.comments.DeleteWithReplies(id)
}

// CountPending 待审核数量
func (s *CommentService) CountPending() (int64, error) {
	return s.comments.CountPending()
}
