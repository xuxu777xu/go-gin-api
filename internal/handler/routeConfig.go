package handler

import (

	// "myGin/internal/bootstrap" // 移除 bootstrap 导入，解决循环依赖
	"myGin/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap" // 引入 zap 包
)

// FlightHandler 处理与机票相关的 HTTP 请求
type FlightHandler struct {
	svc    service.FlightService  // FlightService 定义了机票相关操作的接口。  机票服务接口
	logger *zap.Logger // 添加 logger 字段
}

// NewFlightHandler 创建 FlightHandler 实例，注入 logger
func NewFlightHandler(svc service.FlightService, logger *zap.Logger) *FlightHandler {
	return &FlightHandler{
		svc:    svc,
		logger: logger, // 注入 logger
	}
}

// RegisterRoutes 在 Gin 路由组中注册 FlightHandler 的路由
func (h *FlightHandler) RegisterRoutes(rg *gin.RouterGroup) {
	flightGroup := rg.Group("/flights") // 创建 /flights 子分组
	{
		// 注册机票搜索路由
		flightGroup.POST("/tickets/search", h.SearchTickets)
		// 注册创建订单路由
		flightGroup.POST("/tickets/order", h.CreateOrder)
	}
}