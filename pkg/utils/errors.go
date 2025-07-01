package utils

import "errors"

// 错误定义
var (
	ErrTooManyRedirects = errors.New("太多重定向")
)
