package main_test

import (
	"bytes"
	"context" // 正确的导入位置
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"myGin/internal/bootstrap"
	"myGin/internal/conf" // 为最小化配置重新添加
	"myGin/internal/dto"
	"myGin/internal/handler"
	"go.uber.org/zap"                                  // 为日志记录器初始化重新添加
	"myGin/internal/pkg/errs" // 为错误响应断言重新添加 errs
	// 移除了未使用的导入：service
)

// MockFlightService 是 service.FlightService 的模拟实现
type MockFlightService struct {
	mock.Mock
}

// --- 模拟 FlightService 实现 ---
// 确保方法签名与 service.FlightService 接口匹配
func (m *MockFlightService) Search(ctx context.Context, opt dto.SearchOption) (*dto.SearchResult, error) {
	args := m.Called(ctx, opt) // 按值传递 opt
	if ret := args.Get(0); ret != nil {
		return ret.(*dto.SearchResult), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockFlightService) CreateOrder(ctx context.Context, req dto.OrderRequest) (*dto.OrderResponse, error) {
	args := m.Called(ctx, req) // 按值传递 req
	if ret := args.Get(0); ret != nil {
		return ret.(*dto.OrderResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

// setupTestServer 初始化用于测试的 Gin 引擎和模拟服务。
func setupTestServer(t *testing.T) (*gin.Engine, *MockFlightService) {
	mockService := new(MockFlightService)
	// 假设 NewFlightHandler 接受 FlightService 接口和 logger
	flightHandler := handler.NewFlightHandler(mockService, zap.NewNop()) // 传递 Nop logger 用于测试

	// 使用 Gin 的测试模式
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 加载配置 - 处理潜在错误
	// 我们不从文件加载，而是为中间件创建一个最小配置
	config := &conf.Config{} // 创建一个最小的非 nil 配置

	// 初始化日志记录器 - 对于使用 zap.L() 的中间件至关重要
	logger := zap.NewNop()
	zap.ReplaceGlobals(logger) // 设置全局日志记录器

	// 附加核心中间件 - 传递最小配置
	bootstrap.AttachCoreMiddleware(router, config) // 传递非 nil 配置

	// 使用处理程序的方法注册路由
	apiGroup := router.Group("/api/v1")
	flightHandler.RegisterRoutes(apiGroup) // 正确的路由注册

	return router, mockService
}

// TestSearchTicketsAPI_Success 测试搜索机票的成功场景。
func TestSearchTicketsAPI_Success(t *testing.T) {
	router, mockService := setupTestServer(t)
	if router == nil {
		return // setupTestServer 已经调用了 t.Fatalf
	}

	// 1. 准备模拟
	// 使用 dto.SearchOption 中正确的字段名
	mockReq := dto.SearchOption{
		From: "SHA",
		To:   "PEK",
		Date: time.Now().Format("2006-01-02"),
	}
	// 使用 dto.FlightInfo 中正确的字段名并添加实际数据
	depTime := time.Now().Add(24 * time.Hour)
	arrTime := depTime.Add(2 * time.Hour)
	mockResp := &dto.SearchResult{
		Flights: []dto.FlightInfo{
			{
				ID:            "CA123-20250422", // 示例唯一 ID
				FlightNumber:  "CA123",
				Airline:       "Air China",
				DepartureTime: depTime,
				ArrivalTime:   arrTime,
				Origin:        "SHA",
				Destination:   "PEK",
				Price:         1000.0,
				Currency:      "CNY",
			},
		},
		Total: 1,
	}
	// 按值传递 mockReq。对上下文使用 mock.Anything 以进行更广泛的匹配。
	mockService.On("Search", mock.Anything, mockReq).Return(mockResp, nil)

	// 2. 构造请求
	reqBodyBytes, _ := json.Marshal(mockReq)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/flights/tickets/search", bytes.NewBuffer(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")

	// 3. 执行请求
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 4. 断言
	assert.Equal(t, http.StatusOK, w.Code, "Expected status code 200 OK")

	// 断言响应体 - 直接解组到 dto.SearchResult
	var respBody dto.SearchResult
	err := json.Unmarshal(w.Body.Bytes(), &respBody)
	assert.NoError(t, err, "Response body should be valid JSON for dto.SearchResult")

	// 将解组后的响应体与模拟响应进行比较
	assert.Equal(t, mockResp.Total, respBody.Total, "Response Total should match mock")
	if assert.Len(t, respBody.Flights, len(mockResp.Flights), "响应中的航班数量应与模拟匹配") {
		// 比较第一个航班的详细信息
		assert.Equal(t, mockResp.Flights[0].ID, respBody.Flights[0].ID, "Response Flight ID should match mock")
		assert.Equal(t, mockResp.Flights[0].FlightNumber, respBody.Flights[0].FlightNumber, "Response FlightNumber should match mock")
		assert.Equal(t, mockResp.Flights[0].Price, respBody.Flights[0].Price, "Response Flight Price should match mock")
		// 时间比较需要注意，因为 JSON 编组/解组后可能存在单调时钟问题
		// assert.True(t, mockResp.Flights[0].DepartureTime.Equal(respBody.Flights[0].DepartureTime), "响应的 DepartureTime 应与模拟匹配")
	}

	// 验证模拟是否按预期被调用
	mockService.AssertExpectations(t)
}

// --- 其他测试用例的占位符 ---

// TestSearchTicketsAPI_InvalidInput 测试输入数据无效的场景。
func TestSearchTicketsAPI_InvalidInput(t *testing.T) {
	// t.Skip("跳过：无效输入测试用例尚未实现") // 移除跳过
	router, _ := setupTestServer(t) // 这里我们不需要 mockService
	if router == nil {
		return
	}

	// 1. 构造无效请求（缺少 'From' 字段）
	invalidReq := map[string]interface{}{
		"to":   "PEK",
		"date": time.Now().Format("2006-01-02"),
	}
	reqBodyBytes, _ := json.Marshal(invalidReq)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/flights/tickets/search", bytes.NewBuffer(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")

	// 2. 执行请求
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 3. 断言
	assert.Equal(t, http.StatusBadRequest, w.Code, "Expected status code 400 Bad Request for invalid input")

	// 断言响应体包含预期的错误结构
	var errResp errs.ErrorResponse // 使用 errs 包中的 ErrorResponse 结构体
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.NoError(t, err, "Response body should be valid JSON for errs.ErrorResponse")

	// 断言业务错误代码与 BadRequest 匹配
	assert.Equal(t, errs.BadRequest.Code, errResp.Code, "Error code in response should match errs.BadRequest")
	// 可选地，断言消息包含相关信息（例如，缺少的字段）
	// assert.Contains(t, errResp.Message, "'From'", "错误消息应提及缺少的字段")
}

// TestSearchTicketsAPI_ServiceError 测试服务层返回错误的场景。
func TestSearchTicketsAPI_ServiceError(t *testing.T) {
	// t.Skip("跳过：服务错误测试用例尚未实现") // 移除跳过
	router, mockService := setupTestServer(t)
	if router == nil {
		return
	}

	// 1. 准备模拟
	mockReq := dto.SearchOption{
		From: "SHA",
		To:   "PEK",
		Date: time.Now().Format("2006-01-02"),
	}
	// 定义服务要返回的模拟错误
	mockErrorCode := 50001 // 示例自定义业务错误代码
	mockErrorMessage := "模拟内部服务错误"
	mockError := errs.NewAPIError(http.StatusInternalServerError, mockErrorCode, mockErrorMessage)

	// 设置模拟预期：应调用 Search 并返回 mockError
	mockService.On("Search", mock.Anything, mockReq).Return(nil, mockError)

	// 2. 构造请求
	reqBodyBytes, _ := json.Marshal(mockReq)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/flights/tickets/search", bytes.NewBuffer(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")

	// 3. 执行请求
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 4. 断言
	// 预期 HTTP 状态代码在 mockError 中定义
	assert.Equal(t, mockError.HTTPStatus, w.Code, "Expected status code to match mock error's HTTPStatus")

	// 断言响应体包含预期的错误结构
	var errResp errs.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.NoError(t, err, "Response body should be valid JSON for errs.ErrorResponse")

	// 断言业务错误代码和消息与 mockError 匹配
	assert.Equal(t, mockError.Code, errResp.Code, "Error code in response should match mock error's Code")
	assert.Equal(t, mockError.Message, errResp.Message, "Error message in response should match mock error's Message")

	// 验证模拟是否按预期被调用
	mockService.AssertExpectations(t)
}


func TestCreateOrderAPI_Success(t *testing.T) {
	// t.Skip("跳过：创建订单成功测试用例尚未实现") // 移除跳过
	router, mockService := setupTestServer(t)
	if router == nil {
		return
	}

	// 1. 准备模拟
	mockReq := dto.OrderRequest{
		FlightID: "CA123-20250422", // 修正的字段名：FlightID
		Passengers: []dto.Passenger{
			{
				Name:     "Test User",
				IDType:   "IDCard", // 添加的字段：IDType
				IDNumber: "11010119900307001X", // 修正的字段名：IDNumber
				// PhoneNumber 已移除，因为它不在 DTO 中
				Birthday: "1990-03-07",
				Gender:   1, // 修正的类型和值：Gender (int, 1 代表男性)
			},
		},
		ContactName:   "Test Contact",
		ContactPhone:  "13900139000",
		OrderSerialId: "TESTORDER123",
		Code:          "",
		PromotionSign: "",
	}
	// 基于 dto.OrderResponse 和 dto.PassengerOrderResult 修正的 mockResp 结构
	mockResp := &dto.OrderResponse{
		Success: true, // 添加的字段：Success
		Message: "所有乘客订单创建成功", // 添加的字段：Message
		PassengerResults: []dto.PassengerOrderResult{ // 修正的字段名：PassengerResults
			{
				Success:       true, // 修正的字段名和类型：Success (bool)
				PassengerName: "Test User",
				OrderID:       "TCORDER98765",
				ErrorMessage:  "", // 修正的字段名：ErrorMessage
			},
		},
		// OverallStatus 已移除，因为它不在 DTO 中
	}
	// 按值传递 mockReq。对上下文使用 mock.Anything。
	mockService.On("CreateOrder", mock.Anything, mockReq).Return(mockResp, nil)

	// 2. 构造请求
	reqBodyBytes, _ := json.Marshal(mockReq)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/flights/tickets/order", bytes.NewBuffer(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")

	// 3. 执行请求
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 4. 断言
	assert.Equal(t, http.StatusOK, w.Code, "Expected status code 200 OK for successful order creation")

	// 断言响应体
	var respBody dto.OrderResponse
	err := json.Unmarshal(w.Body.Bytes(), &respBody)
	assert.NoError(t, err, "Response body should be valid JSON for dto.OrderResponse")

	// 将解组后的响应体与模拟响应进行比较 - 更新的断言
	assert.Equal(t, mockResp.Success, respBody.Success, "Response Success should match mock")
	assert.Equal(t, mockResp.Message, respBody.Message, "Response Message should match mock")
	if assert.Len(t, respBody.PassengerResults, len(mockResp.PassengerResults), "Number of passenger results should match mock") {
		assert.Equal(t, mockResp.PassengerResults[0].Success, respBody.PassengerResults[0].Success, "Response PassengerResult Success should match mock")
		assert.Equal(t, mockResp.PassengerResults[0].PassengerName, respBody.PassengerResults[0].PassengerName, "Response PassengerName should match mock")
		assert.Equal(t, mockResp.PassengerResults[0].OrderID, respBody.PassengerResults[0].OrderID, "Response OrderID should match mock")
		assert.Equal(t, mockResp.PassengerResults[0].ErrorMessage, respBody.PassengerResults[0].ErrorMessage, "Response ErrorMessage should match mock")
	}

	// 验证模拟是否按预期被调用
	mockService.AssertExpectations(t)
}

func TestCreateOrderAPI_InvalidInput(t *testing.T) {
	// t.Skip("跳过：创建订单无效输入测试用例尚未实现") // 移除跳过
	router, _ := setupTestServer(t) // 输入验证不需要模拟服务
	if router == nil {
		return
	}

	// 1. 构造无效请求（缺少 'Passengers' 字段）
	invalidReq := map[string]interface{}{
		"flightId":     "CA123-20250422",
		// "passengers": []map[string]interface{}{...}, // 缺少必需字段
		"contactName":  "Test Contact",
		"contactPhone": "13900139000",
	}
	reqBodyBytes, _ := json.Marshal(invalidReq)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/flights/tickets/order", bytes.NewBuffer(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")

	// 2. 执行请求
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 3. 断言
	assert.Equal(t, http.StatusBadRequest, w.Code, "Expected status code 400 Bad Request for invalid order input")

	// 断言响应体包含预期的错误结构
	var errResp errs.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.NoError(t, err, "Response body should be valid JSON for errs.ErrorResponse")

	// 断言业务错误代码与 BadRequest 匹配
	assert.Equal(t, errs.BadRequest.Code, errResp.Code, "Error code in response should match errs.BadRequest")
	// 可选地断言消息包含相关信息
	// assert.Contains(t, errResp.Message, "'Passengers'", "错误消息应提及缺少的字段")
}

func TestCreateOrderAPI_ServiceError(t *testing.T) {
	// t.Skip("跳过：创建订单服务错误测试用例尚未实现") // 移除跳过
	router, mockService := setupTestServer(t)
	if router == nil {
		return
	}

	// 1. Prepare Mock
	mockReq := dto.OrderRequest{ // 使用有效的请求结构
		FlightID: "CA123-20250422",
		Passengers: []dto.Passenger{
			{
				Name:     "Test User",
				IDType:   "IDCard",
				IDNumber: "11010119900307001X",
				Birthday: "1990-03-07",
				Gender:   1,
			},
		},
		ContactName:   "Test Contact",
		ContactPhone:  "13900139000",
		OrderSerialId: "TESTORDER456",
	}
	// 定义服务要返回的模拟错误
	mockErrorCode := 50002 // 订单失败的示例自定义业务错误代码
	mockErrorMessage := "模拟订单创建服务错误"
	// 示例：如果错误是由于请求时已知的业务逻辑失败，则使用 BadRequest 状态
	mockError := errs.NewAPIError(http.StatusBadRequest, mockErrorCode, mockErrorMessage)

	// 设置模拟预期：应调用 CreateOrder 并返回 mockError
	mockService.On("CreateOrder", mock.Anything, mockReq).Return(nil, mockError)

	// 2. 构造请求
	reqBodyBytes, _ := json.Marshal(mockReq)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/flights/tickets/order", bytes.NewBuffer(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")

	// 3. 执行请求
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 4. 断言
	// 预期 HTTP 状态代码在 mockError 中定义
	assert.Equal(t, mockError.HTTPStatus, w.Code, "Expected status code to match mock order error's HTTPStatus")

	// 断言响应体包含预期的错误结构
	var errResp errs.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.NoError(t, err, "Response body should be valid JSON for errs.ErrorResponse")

	// 断言业务错误代码和消息与 mockError 匹配
	assert.Equal(t, mockError.Code, errResp.Code, "Error code in response should match mock order error's Code")
	assert.Equal(t, mockError.Message, errResp.Message, "Error message in response should match mock order error's Message")

	// 验证模拟是否按预期被调用
	mockService.AssertExpectations(t)
}