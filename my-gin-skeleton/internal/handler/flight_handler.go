package handler

import (
	"net/http"

	// "myGin/internal/bootstrap" // 移除 bootstrap 导入，解决循环依赖
	"myGin/internal/dto"
	"myGin/internal/service"
	"myGin/internal/pkg/errs"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap" // 引入 zap 包
)

// FlightHandler 处理与机票相关的 HTTP 请求
type FlightHandler struct {
	svc    service.FlightService
	logger *zap.Logger // 添加 logger 字段
}

// NewFlightHandler 创建 FlightHandler 实例，注入 logger
func NewFlightHandler(svc service.FlightService, logger *zap.Logger) *FlightHandler {
	return &FlightHandler{
		svc:    svc,
		logger: logger, // 注入 logger
	}
}

// SearchTickets 处理搜索机票的 POST 请求
// @Summary 搜索机票
// @Description 根据条件搜索机票信息
// @Tags Flights
// @Accept json
// @Produce json
// @Param searchOption body dto.SearchOption true "搜索条件"
// @Success 200 {object} dto.SearchResult "搜索结果"
// @Failure 400 {object} gin.H "请求参数错误"
// @Failure 500 {object} gin.H "服务器内部错误"
// @Router /v1/flights/tickets/search [post]
func (h *FlightHandler) SearchTickets(c *gin.Context) {
	var opt dto.SearchOption
	if err := c.ShouldBindJSON(&opt); err != nil {
		errs.BadRequest.Wrap(err).JSON(c)
		return
	}

	res, err := h.svc.Search(c.Request.Context(), opt)
	if err != nil {
		// 检查错误是否已经是 APIError 类型
		if apiErr, ok := err.(*errs.APIError); ok {
			apiErr.JSON(c) // 返回具体的 APIError
		} else {
			// 否则，将其包装为内部服务器错误
			errs.InternalServerError.Wrap(err).JSON(c)
		}
		return
	}

	c.JSON(http.StatusOK, res)
}

// CreateOrder 处理创建订单的 POST 请求
// @Summary 创建机票订单
// @Description 根据请求信息创建机票订单
// @Tags Flights
// @Accept json
// @Produce json
// @Param orderRequest body dto.OrderRequest true "订单请求信息"
// @Success 200 {object} dto.OrderResponse "订单创建结果"
// @Failure 400 {object} errs.APIError "请求参数错误或业务逻辑失败"
// @Failure 500 {object} errs.APIError "服务器内部错误"
// @Router /v1/flights/tickets/order [post]
func (h *FlightHandler) CreateOrder(c *gin.Context) {
	var req dto.OrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 添加日志记录，记录请求体绑定失败
		h.logger.Warn("Request body binding failed", // 使用注入的 logger
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
			zap.String("error", err.Error()),
			zap.String("ip", c.ClientIP()),
		)
		errs.BadRequest.Wrap(err).JSON(c)
		return
	}

	// 注意：此 Handler 假设以下字段的值由客户端在请求体中提供，并通过 ShouldBindJSON 绑定到 req 对象:
	// - OrderSerialId: 假设客户端提供，其确切来源（如是否来自先前的 buildtemporder 调用）需根据业务流程确认。
	// - Code: 优惠码/活动代码。
	// - PromotionSign: 促销标识。
	// - Passengers 数组中的 Birthday 和 Gender 字段。
	// 如果这些字段需要从其他来源（如 Header、Context、Config 或在 Handler 中生成），则需要在此处添加相应逻辑。

	// 将包含绑定值的 req 对象传递给 Service 层
	resp, err := h.svc.CreateOrder(c.Request.Context(), req)
	if err != nil {
		// 检查错误是否已经是 APIError 类型
		if apiErr, ok := err.(*errs.APIError); ok {
			apiErr.JSON(c) // 返回具体的 APIError
		} else {
			// 否则，将其包装为内部服务器错误
			errs.InternalServerError.Wrap(err).JSON(c)
		}
		return
	}

	// 检查业务逻辑是否成功
	if !resp.Success {
		// 将业务失败转换为 APIError 返回，保持错误格式统一
		// 使用 40000 作为通用的业务逻辑失败代码
		errs.NewAPIError(http.StatusBadRequest, 40000, resp.Message).JSON(c)
		return
	}

	// 业务成功，返回订单信息
	c.JSON(http.StatusOK, resp)
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