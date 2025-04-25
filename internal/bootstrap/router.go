package bootstrap

import (
	"net/http"

	"myGin/internal/conf"    // 模块路径
	"myGin/internal/handler" // 导入 handler 包
	"myGin/internal/service" // 导入 service 包

	"github.com/gin-gonic/gin"
	"go.uber.org/zap" // 使用 zap 进行日志记录
)

// RegisterRoutes 初始化并注册所有应用程序路由。
func RegisterRoutes(engine *gin.Engine, cfg *conf.Config) {
	// 使用 bootstrap.GetLogger() 获取 logger 实例
	// 不再需要检查或从 zap.L() 获取
	logger := GetLogger() // 直接调用包内函数

	logger.Info("Registering application routes...")

	// // 1. 健康检查端点（对监控至关重要）
	// engine.GET("/healthz", func(c *gin.Context) {
	// 	// 在实际应用中，这里可能会检查数据库/Redis 连接。
	// 	c.String(http.StatusOK, "OK")
	// })
	logger.Debug("Registered health check route: GET /healthz")

	// 2. API 版本分组（良好实践）
	apiV1 := engine.Group("/api/v1")
	{ // 大括号提高了分组路由的可读性
		logger.Info("Setting up API v1 route group", zap.String("prefix", "/api/v1"))

		// --- 业务逻辑路由 ---






		// 初始化并注册航班路由
		///api/v1/flights/ 路由组
		apiClient := service.NewTongchengAPIClientImpl() // 创建生产环境 API 客户端
		flightSvc := service.NewFlightService(logger, apiClient) // 使用客户端初始化航班服务
		flightHdl := handler.NewFlightHandler(flightSvc, logger) // 注入 logger
		flightHdl.RegisterRoutes(apiV1)                          // 注册航班路由（例如 /api/v1/flights/tickets/search）
		logger.Info("Registered flight routes")




		// 添加一个简单的 ping 路由用于测试 v1 分组
		apiV1.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "pong v1"})
		})
		logger.Debug("Registered ping route: GET /api/v1/ping")

		// --- 业务逻辑路由结束 ---
	} // apiV1 分组结束

	// TODO: 根据需要添加其他 API 版本或路由组（例如 /admin）

	logger.Info("Finished registering application routes.")
}