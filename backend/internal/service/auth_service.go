package service

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"blog/internal/model"
	"blog/internal/pkg/hash"
	"blog/internal/pkg/jwt"
	"blog/internal/repository"

	"gorm.io/gorm"
)

// 用户名规则：3~50 位，字母数字下划线
var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]{3,50}$`)

// 密码最短长度。够用即可，别设得让用户想贴便签在显示器上。
const minPasswordLen = 6

// AuthService 认证相关业务
type AuthService struct {
	db    *gorm.DB
	users *repository.UserRepository
	jwt   *jwt.Manager
}

func NewAuthService(db *gorm.DB, users *repository.UserRepository, jwtMgr *jwt.Manager) *AuthService {
	return &AuthService{db: db, users: users, jwt: jwtMgr}
}

// LoginResult 登录成功返回给前端的东西
type LoginResult struct {
	Token     string     `json:"token"`
	ExpiresAt time.Time  `json:"expires_at"`
	User      model.User `json:"user"`
}

// Register 注册新用户（默认角色 author）
func (s *AuthService) Register(username, password, nickname, email string) (*model.User, error) {
	if !usernamePattern.MatchString(username) {
		return nil, fmt.Errorf("%w: 用户名需为 3~50 位字母、数字或下划线", ErrInvalidInput)
	}
	if len(password) < minPasswordLen {
		return nil, fmt.Errorf("%w: 密码至少 %d 位", ErrInvalidInput, minPasswordLen)
	}
	if len(password) > hash.MaxPasswordLen {
		return nil, fmt.Errorf("%w: 密码不能超过 %d 字节", ErrInvalidInput, hash.MaxPasswordLen)
	}

	exists, err := s.users.ExistsByUsername(username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("%w: 用户名 %s 已被占用", ErrConflict, username)
	}

	hashed, err := hash.Password(password)
	if err != nil {
		return nil, err
	}

	if nickname == "" {
		nickname = username
	}

	u := &model.User{
		Username:     username,
		PasswordHash: hashed,
		Nickname:     nickname,
		Email:        email,
		Role:         "author",
		Status:       1,
	}
	if err := s.users.Create(u); err != nil {
		return nil, err
	}
	return u, nil
}

// Login 登录：校验账号密码，签发 token
func (s *AuthService) Login(username, password string) (*LoginResult, error) {
	u, err := s.users.FindByUsername(username)

	// 关键细节：用户不存在和密码错误，返回同一个错误。
	// 如果区分开，攻击者就能用「用户名不存在」这个提示批量探测哪些账号真实存在（用户名枚举）。
	if err != nil || u == nil {
		return nil, fmt.Errorf("%w: 用户名或密码错误", ErrUnauthorized)
	}
	if !hash.Check(u.PasswordHash, password) {
		return nil, fmt.Errorf("%w: 用户名或密码错误", ErrUnauthorized)
	}
	if u.Status != 1 {
		return nil, fmt.Errorf("%w: 账号已被禁用", ErrUnauthorized)
	}

	token, expireAt, err := s.jwt.Generate(u.ID, u.Username, u.Role)
	if err != nil {
		return nil, err
	}

	return &LoginResult{Token: token, ExpiresAt: expireAt, User: *u}, nil
}

// Profile 查询当前登录用户资料
func (s *AuthService) Profile(userID uint64) (*model.User, error) {
	u, err := s.users.FindByID(userID)
	if err != nil {
		return nil, ErrNotFound
	}
	return u, nil
}

// ChangePassword 修改密码：先验旧密码，再写新哈希
func (s *AuthService) ChangePassword(userID uint64, oldPassword, newPassword string) error {
	u, err := s.users.FindByID(userID)
	if err != nil {
		return ErrNotFound
	}

	if !hash.Check(u.PasswordHash, oldPassword) {
		return fmt.Errorf("%w: 原密码错误", ErrInvalidInput)
	}
	if len(newPassword) < minPasswordLen {
		return fmt.Errorf("%w: 新密码至少 %d 位", ErrInvalidInput, minPasswordLen)
	}
	if len(newPassword) > hash.MaxPasswordLen {
		return fmt.Errorf("%w: 新密码不能超过 %d 字节", ErrInvalidInput, hash.MaxPasswordLen)
	}
	if oldPassword == newPassword {
		return fmt.Errorf("%w: 新密码不能与原密码相同", ErrInvalidInput)
	}

	hashed, err := hash.Password(newPassword)
	if err != nil {
		return err
	}
	if err := s.users.UpdatePasswordHash(userID, hashed); err != nil {
		return err
	}
	return nil
}

// IsNotFoundErr 便于 handler 判断
func IsNotFoundErr(err error) bool { return errors.Is(err, ErrNotFound) }
