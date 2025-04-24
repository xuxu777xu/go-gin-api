package errs

import (
	"fmt"
	"net/http"
)

// APIError 定义了 API 响应的标准错误结构。
type APIError struct {
	HTTPStatus    int    // HTTP 状态码 (例如, 400, 404, 500)
	Code          int    // 自定义业务错误码
	Message       string // 用户友好的错误信息
	originalError error  // 原始底层错误，可选
}

// Error 实现了标准的 error 接口。
func (e *APIError) Error() string {
	if e.originalError != nil {
		return fmt.Sprintf("APIError: HTTPStatus=%d, Code=%d, Message=%s, OriginalError=%v", e.HTTPStatus, e.Code, e.Message, e.originalError)
	}
	return fmt.Sprintf("APIError: HTTPStatus=%d, Code=%d, Message=%s", e.HTTPStatus, e.Code, e.Message)
}

// Wrap 使用 APIError 包装现有错误，创建一个新实例。
// 保留原始 APIError 的 HTTPStatus、Code 和 Message。
func (e *APIError) Wrap(err error) *APIError {
	if err == nil {
		return e // 如果没有错误需要包装，则返回原始错误
	}
	// 创建一个新的 APIError 实例以避免修改原始预设错误。
	return &APIError{
		HTTPStatus:    e.HTTPStatus,
		Code:          e.Code,
		Message:       e.Message, // 保留预定义的消息
		originalError: err,
	}
}

// WrapWithMessage 使用 APIError 和自定义消息包装现有错误。
func (e *APIError) WrapWithMessage(err error, message string, args ...interface{}) *APIError {
	// 创建一个新的 APIError 实例以避免修改原始预设错误。
	newErr := &APIError{
		HTTPStatus:    e.HTTPStatus,
		Code:          e.Code,
		Message:       fmt.Sprintf(message, args...), // 使用自定义消息
		originalError: err,                           // 保留原始错误（如果提供）
	}
	// 如果原始错误也是 APIError，则链接原始错误
	if apiErr, ok := err.(*APIError); ok {
		newErr.originalError = apiErr
	}
	return newErr
}

// === 预定义错误 ===

// 通用错误 (根据需要自定义 Code)
var (
	BadRequest = &APIError{HTTPStatus: http.StatusBadRequest, Code: 40000, Message: "错误的请求"}         // 400 错误请求
	Unauthorized = &APIError{HTTPStatus: http.StatusUnauthorized, Code: 40100, Message: "未授权"}       // 401 未授权
	Forbidden    = &APIError{HTTPStatus: http.StatusForbidden, Code: 40300, Message: "禁止访问"}          // 403 禁止访问
	NotFound     = &APIError{HTTPStatus: http.StatusNotFound, Code: 40400, Message: "资源未找到"}        // 404 未找到
	Conflict     = &APIError{HTTPStatus: http.StatusConflict, Code: 40900, Message: "资源冲突"}          // 409 冲突
	TooManyRequests = &APIError{HTTPStatus: http.StatusTooManyRequests, Code: 42900, Message: "请求过于频繁"} // 429 请求过多

	InternalServerError = &APIError{HTTPStatus: http.StatusInternalServerError, Code: 50000, Message: "服务器内部错误"} // 500 服务器内部错误
	ServiceUnavailable  = &APIError{HTTPStatus: http.StatusServiceUnavailable, Code: 50300, Message: "服务不可用"}    // 503 服务不可用
)

// NewAPIError 创建一个新的 APIError。
// 通常建议对预定义错误使用 Wrap 或 WrapWithMessage。
func NewAPIError(httpStatus, code int, message string) *APIError {
	return &APIError{
		HTTPStatus: httpStatus,
		Code:       code,
		Message:    message,
	}
}