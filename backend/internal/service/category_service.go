package service

import (
	"fmt"
	"strings"

	"blog/internal/model"
	"blog/internal/repository"
)

// CategoryService 分类业务
type CategoryService struct {
	repo *repository.CategoryRepository
}

func NewCategoryService(repo *repository.CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

// CategoryInput 新建/编辑分类的入参
type CategoryInput struct {
	Name        string
	Slug        string
	Description string
	Sort        int
}

// List 分类列表（带文章数）
func (s *CategoryService) List() ([]repository.CategoryWithCount, error) {
	return s.repo.ListWithArticleCount()
}

// Create 新建分类
func (s *CategoryService) Create(in CategoryInput) (*model.Category, error) {
	if err := s.validate(in); err != nil {
		return nil, err
	}

	exists, err := s.repo.ExistsNameOrSlug(in.Name, in.Slug, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("%w: 分类名或标识已存在", ErrConflict)
	}

	c := &model.Category{
		Name:        in.Name,
		Slug:        in.Slug,
		Description: in.Description,
		Sort:        in.Sort,
	}
	if err := s.repo.Create(c); err != nil {
		return nil, err
	}
	return c, nil
}

// Update 编辑分类
func (s *CategoryService) Update(id uint64, in CategoryInput) (*model.Category, error) {
	c, err := s.repo.FindByID(id)
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
		return nil, fmt.Errorf("%w: 分类名或标识已被其他分类占用", ErrConflict)
	}

	c.Name = in.Name
	c.Slug = in.Slug
	c.Description = in.Description
	c.Sort = in.Sort
	if err := s.repo.Save(c); err != nil {
		return nil, err
	}
	return c, nil
}

// Delete 删除分类。
// 业务规则：分类下还有文章时不允许删除——否则那些文章会指向一个不存在的分类。
// 想删就先让用户把文章移走，这是「保护数据完整性」的典型做法。
func (s *CategoryService) Delete(id uint64) error {
	if _, err := s.repo.FindByID(id); err != nil {
		return ErrNotFound
	}

	n, err := s.repo.CountArticles(id)
	if err != nil {
		return err
	}
	if n > 0 {
		return fmt.Errorf("%w: 该分类下还有 %d 篇文章，请先移走或删除文章", ErrConflict, n)
	}

	return s.repo.Delete(id)
}

func (s *CategoryService) validate(in CategoryInput) error {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return fmt.Errorf("%w: 分类名不能为空", ErrInvalidInput)
	}
	if len([]rune(in.Name)) > 50 {
		return fmt.Errorf("%w: 分类名不能超过 50 个字", ErrInvalidInput)
	}
	if in.Slug == "" {
		return fmt.Errorf("%w: 标识 slug 不能为空", ErrInvalidInput)
	}
	if len(in.Slug) > 50 {
		return fmt.Errorf("%w: slug 不能超过 50 个字符", ErrInvalidInput)
	}
	return nil
}
