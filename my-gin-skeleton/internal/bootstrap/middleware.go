package bootstrap

import (
	"net" // 用于 *net.OpError 检查
	"net/http"
	"net/http/httputil"
	"os" // 用于 *os.SyscallError 检查
	"runtime/debug"
	"strings"
	"time"
	"myGin/internal/conf" // 模块路径
	"myGin/internal/pkg/errs"
	"fmt" // 引入 fmt 包用于格式化错误信息

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	// "github.com/gin-contrib/cors" // 如果后续需要 CORS，请取消注释
)

// AttachCoreMiddleware 将核心中间件附加到 Gin 引擎。
// 包括 Gin 的默认 Logger 和一个基于 Zap 的自定义 Recovery 中间件。
func AttachCoreMiddleware(engine *gin.Engine, cfg *conf.Config) {
	// 使用 Gin 的默认 Logger 中间件
	// 将请求详细信息记录到标准输出。如果需要，可以考虑替换为 ZapLogger。
	engine.Use(gin.Logger())
	GetLogger().Debug("Attached gin.Logger middleware") // 使用 GetLogger()

	// 使用带有 Zap 日志记录的自定义 Recovery 中间件
	engine.Use(RecoveryWithZap(GetLogger())) // 使用 GetLogger()
	GetLogger().Debug("Attached custom RecoveryWithZap middleware") // 使用 GetLogger()

	// 添加新的错误日志中间件，放在 Recovery 之后处理非 panic 错误
	engine.Use(ErrorLoggerMiddleware(GetLogger())) // 使用 GetLogger()
	GetLogger().Debug("Attached custom ErrorLoggerMiddleware") // 使用 GetLogger()


	// --- 可选：CORS 中间件示例 ---
	// 如果需要跨域资源共享 (CORS)，请取消注释并进行配置。
	/*
	   import "github.com/gin-contrib/cors"
	   import "time"

	   corsConfig := cors.Config{
	       // AllowOrigins: []string{"http://localhost:3000", "https://your-frontend.com"}, // 指定允许的源
	       AllowAllOrigins:  true, // 允许所有源 (安全性较低，用于开发)
	       AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
	       AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-Requested-With"},
	       ExposeHeaders:    []string{"Content-Length"},
	       AllowCredentials: true,
	       MaxAge:           12 * time.Hour,
	   }
	   engine.Use(cors.New(corsConfig))
	   GetLogger().Debug("Attached CORS middleware") // 使用 GetLogger()
	*/

	GetLogger().Info("Attached core middleware (Logger, Recovery)") // 使用 GetLogger()
}

// RecoveryWithZap 返回一个中间件，该中间件从任何 panic 中恢复并使用 Zap 记录它们。
// 它还会写入一个 500 JSON 错误响应。
func RecoveryWithZap(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 检查连接是否断开，因为这并非真正的服务器错误。
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				if brokenPipe {
					logger.Warn("Recovered from broken pipe",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
					// 如果连接已断开，我们无法向其写入状态。
					c.Error(err.(error)) // nolint: errcheck
					c.Abort()
					return
				}

				// 记录 panic 错误以及堆栈跟踪
				logger.Error("Recovered from panic",
					zap.Time("time", time.Now()),
					zap.Any("error", err),
					zap.String("request", string(httpRequest)),
					zap.String("stack", string(debug.Stack())),
				)

				// 返回内部服务器错误响应
				errs.InternalServerError.JSON(c)
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}


// ErrorLoggerMiddleware 创建一个中间件，用于记录请求处理过程中可能出现的非 panic 错误。
// 它检查 c.Errors 和最终的 HTTP 状态码。
func ErrorLoggerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先执行后续的处理函数和中间件
		c.Next()

		// c.Next() 执行完毕后，检查是否有错误或失败的状态码
		logErrorsIfExist(c, logger)
	}
}

// logErrorsIfExist 辅助函数，检查并记录错误
func logErrorsIfExist(c *gin.Context, logger *zap.Logger) {
	status := c.Writer.Status()
	hasErrors := len(c.Errors) > 0

	// 只记录客户端错误 (4xx) 或服务器错误 (5xx)
	// 注意：5xx panic 错误通常已被 RecoveryWithZap 处理，但这里可以捕获未 panic 的 5xx
	if status >= 400 || hasErrors {
		// 准备日志字段
		fields := []zap.Field{
			zap.Int("status", status),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("query", c.Request.URL.RawQuery),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			// zap.Duration("latency", latency), // 如果需要延迟，可以在中间件开始时记录时间
		}

		// 构造错误信息字符串
		var errorMessages strings.Builder
		if hasErrors {
			for i, ginErr := range c.Errors {
				if i > 0 {
					errorMessages.WriteString("; ")
				}
				// 尝试获取更具体的错误信息
				metaMsg := ""
				if ginErr.Meta != nil {
					metaMsg = fmt.Sprintf(" (Meta: %v)", ginErr.Meta) // 使用 fmt 将 Meta 格式化
				}
				errorMessages.WriteString(fmt.Sprintf("Error #%d: %s%s", i+1, ginErr.Error(), metaMsg)) // 使用 fmt
			}
		} else {
			// 如果没有 c.Errors，但状态码是 4xx/5xx，记录一个通用消息
			errorMessages.WriteString(fmt.Sprintf("Request failed with status %d", status))
		}
		fields = append(fields, zap.String("errors", errorMessages.String()))


		// 根据状态码选择日志级别
		if status >= 500 {
			logger.Error("Request completed with server error", fields...)
		} else { // status >= 400
			// 如果是绑定错误，我们可能已经在 handler 记过 Warn 了，但这里提供统一入口
			// 可以考虑增加逻辑避免重复，但多记一次通常问题不大
			logger.Warn("Request completed with client error", fields...)
		}
	}
}


// 可选：ZapLogger 返回一个使用 Zap 记录请求的中间件。
// func ZapLogger(logger *zap.Logger) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		start := time.Now()
// 		path := c.Request.URL.Path
// 		query := c.Request.URL.RawQuery

// 		// 处理请求
// 		c.Next()

// 		// 记录请求详细信息
// 		end := time.Now()
// 		latency := end.Sub(start)
// 		if len(c.Errors) > 0 {
// 			// 如果这是一个错误的请求，则附加错误字段。
// 			for _, e := range c.Errors.Errors() {
// 				logger.Error(e)
// 			}
// 		} else {
// 			logger.Info(path,
// 				zap.Int("status", c.Writer.Status()),
// 				zap.String("method", c.Request.Method),
// 				zap.String("path", path),
// 				zap.String("query", query),
// 				zap.String("ip", c.ClientIP()),
// 				zap.String("user-agent", c.Request.UserAgent()),
// 				zap.Duration("latency", latency),
// 			)
// 		}
// 	}
// }