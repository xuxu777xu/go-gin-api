package errs

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap" // 引入 zap 日志库
)

// ErrorResponse 定义了错误 JSON 响应的结构。
type ErrorResponse struct {
	Code    int    `json:"code"`    // 自定义业务错误码
	Message string `json:"message"` // 用户友好的错误信息
	// 可选地添加 RequestID、Details 等。
	// RequestID string `json:"requestId,omitempty"`
}

// JSON 使用提供的 Gin 上下文将 APIError 作为 JSON 响应发送。
// 它设置适当的 HTTP 状态码并记录错误（如果 originalError 存在）。
func (e *APIError) JSON(c *gin.Context) {
	// 如果原始错误存在且不为 nil，则记录该错误
	if e.originalError != nil {
		// 假设上下文中或全局范围内有可用的 logger
		// 如果不同，请替换为您实际的 logger 实现
		logger, _ := c.Get("logger") // 示例：从上下文中获取 logger
		if zl, ok := logger.(*zap.Logger); ok {
			zl.Error("API Error occurred",
				zap.Int("httpStatus", e.HTTPStatus),
				zap.Int("code", e.Code),
				zap.String("message", e.Message),
				zap.Error(e.originalError), // 记录包装后的原始错误
			)
		} else {
			// 如果 logger 不是 *zap.Logger 或未找到，则使用回退或默认日志记录
			// log.Printf("[Error] HTTP %d - Code %d: %s | Original: %v\n", e.HTTPStatus, e.Code, e.Message, e.originalError)
		}
	} else {
		// 记录没有原始错误的错误，但也许是不太严重的错误？可选。
		// logger, _ := c.Get("logger")
		// if zl, ok := logger.(*zap.Logger); ok {
		// 	zl.Warn("API Error Response", // 对潜在的不太严重的错误使用 Warn 级别
		// 		zap.Int("httpStatus", e.HTTPStatus),
		// 		zap.Int("code", e.Code),
		// 		zap.String("message", e.Message),
		// 	)
		// }
	}

	c.AbortWithStatusJSON(e.HTTPStatus, ErrorResponse{
		Code:    e.Code,
		Message: e.Message,
		// RequestID: c.GetString("request_id"), // Example: include request ID if available
	})
}