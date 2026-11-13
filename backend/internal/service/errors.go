package service

import (
	"errors"

	"gorm.io/gorm"
)

// 业务错误：service 用这几个哨兵错误表达「业务上出了什么事」，
// handler 负责翻译成 HTTP 状态码 + 业务码。
var (
	ErrNotFound     = errors.New("资源不存在")
	ErrInvalidInput = errors.New("参数不合法")
	ErrUnauthorized = errors.New("认证失败")
	ErrConflict     = errors.New("资源冲突")
	ErrTooManyReqs  = errors.New("请求过于频繁")
)

// IsNotFound 判断是否是「找不到」类错误
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound) || errors.Is(err, gorm.ErrRecordNotFound)
}
