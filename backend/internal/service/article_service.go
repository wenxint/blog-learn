package service

import (
	"fmt"
	"time"

	"blog/internal/model"
	"blog/internal/repository"

	"gorm.io/gorm"
)

// 文章状态
const (
	StatusDraft     int8 = 0
	StatusPublished int8 = 1
)

// ArticleService 业务层：做校验、组合多次数据操作、控制事务。
// 它不关心 HTTP（不 import gin），也不关心 SQL（不 import gorm 的查询细节）。
type ArticleService struct {
	db   *gorm.DB
	repo *repository.ArticleRepository
}

func NewArticleService(db *gorm.DB, repo *repository.ArticleRepository) *ArticleService {
	return &ArticleService{db: db, repo: repo}
}

// ArticleInput 新建和更新共用的入参
type ArticleInput struct {
	Title      string
	Summary    string
	Content    string
	Cover      string
	Status     int8
	IsTop      int8
	AuthorID   uint64
	CategoryID uint64
	TagIDs     []uint64
}

// List 文章列表
func (s *ArticleService) List(f repository.ArticleFilter) ([]model.Article, int64, error) {
	return s.repo.List(f)
}

// Detail 文章详情：校验可见性 + 阅读数自增
func (s *ArticleService) Detail(id uint64, onlyPublished bool) (*model.Article, error) {
	a, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrNotFound
	}
	if onlyPublished && a.Status != StatusPublished {
		return nil, ErrNotFound
	}

	// 阅读数自增用 SQL 表达式，并发安全；这里只同步内存里的值用于返回
	if err := s.repo.IncrementViewCount(id); err == nil {
		a.ViewCount++
	}
	return a, nil
}

// Create 新建文章：写文章 + 写标签关联，必须在同一个事务里。
// 否则中间失败会出现「文章建好了但标签没了」的脏数据。
func (s *ArticleService) Create(in ArticleInput) (*model.Article, error) {
	if err := s.validate(in); err != nil {
		return nil, err
	}
	tags, err := s.loadTags(in.TagIDs)
	if err != nil {
		return nil, err
	}

	a := &model.Article{
		Title:      in.Title,
		Summary:    in.Summary,
		Content:    in.Content,
		Cover:      in.Cover,
		Status:     in.Status,
		IsTop:      in.IsTop,
		AuthorID:   in.AuthorID,
		CategoryID: in.CategoryID,
	}
	if in.Status == StatusPublished {
		now := time.Now()
		a.PublishedAt = &now
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		repo := s.repo.WithTx(tx)
		if err := repo.Create(a); err != nil {
			return err
		}
		if len(tags) > 0 {
			if err := repo.ReplaceTags(a.ID, tagIDs(tags)); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 返回带关联的完整数据
	return s.repo.FindByID(a.ID)
}

// Update 编辑文章：全量替换字段 + 全量替换标签，事务保证一致
func (s *ArticleService) Update(id uint64, in ArticleInput) (*model.Article, error) {
	a, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrNotFound
	}
	if err := s.validate(in); err != nil {
		return nil, err
	}
	tags, err := s.loadTags(in.TagIDs)
	if err != nil {
		return nil, err
	}

	a.Title = in.Title
	a.Summary = in.Summary
	a.Content = in.Content
	a.Cover = in.Cover
	a.Status = in.Status
	a.IsTop = in.IsTop
	a.CategoryID = in.CategoryID

	// 状态流转：草稿 -> 发布，补发布时间；发布 -> 草稿，清空发布时间
	switch in.Status {
	case StatusPublished:
		if a.PublishedAt == nil {
			now := time.Now()
			a.PublishedAt = &now
		}
	case StatusDraft:
		a.PublishedAt = nil
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		repo := s.repo.WithTx(tx)
		if err := repo.Save(a); err != nil {
			return err
		}
		// 传空切片表示清空该文章的所有标签
		if err := repo.ReplaceTags(a.ID, tagIDs(tags)); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.repo.FindByID(id)
}

// Delete 删除文章（软删除）
func (s *ArticleService) Delete(id uint64) error {
	if _, err := s.repo.FindByID(id); err != nil {
		return ErrNotFound
	}
	return s.repo.Delete(id)
}

// validate 业务校验
func (s *ArticleService) validate(in ArticleInput) error {
	if in.Title == "" {
		return fmt.Errorf("%w: 标题不能为空", ErrInvalidInput)
	}
	if in.AuthorID == 0 {
		return fmt.Errorf("%w: 作者不能为空", ErrInvalidInput)
	}
	if in.Status != StatusDraft && in.Status != StatusPublished {
		return fmt.Errorf("%w: 状态只能是 0(草稿) 或 1(发布)", ErrInvalidInput)
	}
	if in.CategoryID > 0 {
		var n int64
		if err := s.db.Model(&model.Category{}).Where("id = ?", in.CategoryID).Count(&n).Error; err != nil {
			return err
		}
		if n == 0 {
			return fmt.Errorf("%w: 分类 %d 不存在", ErrInvalidInput, in.CategoryID)
		}
	}
	return nil
}

// loadTags 校验标签是否都存在，并去重。
// 返回真实的 Tag 对象给 Association 用。
func (s *ArticleService) loadTags(ids []uint64) ([]model.Tag, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	seen := make(map[uint64]struct{}, len(ids))
	unique := make([]uint64, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}

	var tags []model.Tag
	if err := s.db.Where("id IN ?", unique).Find(&tags).Error; err != nil {
		return nil, err
	}
	if len(tags) != len(unique) {
		return nil, fmt.Errorf("%w: 存在无效的标签 id", ErrInvalidInput)
	}
	return tags, nil
}

// tagIDs 从 Tag 对象里取出 id 列表
func tagIDs(tags []model.Tag) []uint64 {
	ids := make([]uint64, 0, len(tags))
	for _, t := range tags {
		ids = append(ids, t.ID)
	}
	return ids
}

// Archives 归档统计
func (s *ArticleService) Archives() ([]repository.ArchiveItem, error) {
	return s.repo.Archives()
}
