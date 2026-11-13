package service

import (
	"fmt"
	"strings"

	"blog/internal/model"
	"blog/internal/repository"
)

// TagService 标签业务
type TagService struct {
	repo *repository.TagRepository
}

func NewTagService(repo *repository.TagRepository) *TagService {
	return &TagService{repo: repo}
}

// TagInput 新建/编辑标签的入参
type TagInput struct {
	Name string
	Slug string
}

// List 标签列表（带文章数）
func (s *TagService) List() ([]repository.TagWithCount, error) {
	return s.repo.ListWithArticleCount()
}

// Create 新建标签
func (s *TagService) Create(in TagInput) (*model.Tag, error) {
	if err := s.validate(in); err != nil {
		return nil, err
	}

	exists, err := s.repo.ExistsNameOrSlug(in.Name, in.Slug, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("%w: 标签名或标识已存在", ErrConflict)
	}

	t := &model.Tag{Name: in.Name, Slug: in.Slug}
	if err := s.repo.Create(t); err != nil {
		return nil, err
	}
	return t, nil
}

// Update 编辑标签
func (s *TagService) Update(id uint64, in TagInput) (*model.Tag, error) {
	t, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrNotFound
	}
	if err := s.validate(in); err != nil {
		return nil, err
	}

	exists, err := s.repo.ExistsNameOrSlug(in.Name, in.Slug, id)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("%w: 标签名或标识已被其他标签占用", ErrConflict)
	}

	t.Name = in.Name
	t.Slug = in.Slug
	if err := s.repo.Save(t); err != nil {
		return nil, err
	}
	return t, nil
}

// Delete 删除标签：连同事务清理 article_tags 关联。
// 与分类不同，标签允许直接删——文章失去一个标签不影响内容完整性。
func (s *TagService) Delete(id uint64) error {
	if _, err := s.repo.FindByID(id); err != nil {
		return ErrNotFound
	}
	return s.repo.DeleteWithLinks(id)
}

func (s *TagService) validate(in TagInput) error {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return fmt.Errorf("%w: 标签名不能为空", ErrInvalidInput)
	}
	if len([]rune(in.Name)) > 50 {
		return fmt.Errorf("%w: 标签名不能超过 50 个字", ErrInvalidInput)
	}
	if in.Slug == "" {
		return fmt.Errorf("%w: 标识 slug 不能为空", ErrInvalidInput)
	}
	if len(in.Slug) > 50 {
		return fmt.Errorf("%w: slug 不能超过 50 个字符", ErrInvalidInput)
	}
	return nil
}
