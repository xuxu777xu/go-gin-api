package service

import (
	"context"
	// "errors" // 移除了未使用的导入 - encoding/json 已移除
	"fmt"
	"strings" // 导入 strings 包
	"time"

	// "myGin/internal/conf" // 移除了未使用的导入
	"myGin/internal/dto"
	"myGin/internal/pkg/tongchengapi" // 更新了导入路径

	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

// 用于认证令牌的 Context 键
type contextKey string

const (
	CtxKeyTcUserID   contextKey = "tcUserID"
	CtxKeyTcSecToken contextKey = "tcSecToken"
	CtxKeySecToken   contextKey = "secToken" // 对应 CreateOrder 中的 tcsectk 头部
	CtxKeyDeviceID   contextKey = "deviceID"
)

// TongchengAPIClient 定义了与同程 API 交互的接口。
// 这允许在测试期间模拟 API 调用。
type TongchengAPIClient interface {
	Get_airline_message(opts *tongchengapi.Options) (string, error)
	CreateOrder(opts *tongchengapi.Options) (string, error)
}

// FlightService 定义了机票相关操作的接口。
type FlightService interface {
	Search(ctx context.Context, opt dto.SearchOption) (*dto.SearchResult, error)
	CreateOrder(ctx context.Context, req dto.OrderRequest) (*dto.OrderResponse, error)
}

// flightService 实现了 FlightService 接口。
type flightService struct {
	logger    *zap.Logger
	apiClient TongchengAPIClient // 注入的依赖
	// config *conf.Config // 示例：如果令牌存储在配置中，则注入配置
}

// NewFlightService 创建一个新的 FlightService 实例并注入依赖。
func NewFlightService(logger *zap.Logger, apiClient TongchengAPIClient /*, cfg *conf.Config*/) FlightService {
	if logger == nil {
		logger = zap.NewNop()
	}
	if apiClient == nil {
		// 在实际应用中，你可能需要在此处 panic 或返回错误，
		// 或者提供一个默认的生产客户端。对于测试，必须提供此客户端。
		panic("apiClient cannot be nil")
	}
	return &flightService{
		logger:    logger.Named("flightService"),
		apiClient: apiClient,
		// config: cfg,
	}
}

// 从 context 获取令牌的辅助函数
func getAuthTokensFromContext(ctx context.Context) (userID, tcSecToken, secToken, deviceID string, err error) {
	userIDVal := ctx.Value(CtxKeyTcUserID)
	tcSecTokenVal := ctx.Value(CtxKeyTcSecToken)
	secTokenVal := ctx.Value(CtxKeySecToken)
	deviceIDVal := ctx.Value(CtxKeyDeviceID)

	var missing []string
	if userIDVal == nil || userIDVal.(string) == "" {
		missing = append(missing, string(CtxKeyTcUserID))
	} else {
		userID = userIDVal.(string)
	}
	if tcSecTokenVal == nil || tcSecTokenVal.(string) == "" {
		missing = append(missing, string(CtxKeyTcSecToken))
	} else {
		tcSecToken = tcSecTokenVal.(string)
	}
	// secToken 和 deviceID 主要用于 CreateOrder
	if secTokenVal == nil || secTokenVal.(string) == "" {
		// 对于 Search 是可选的，后续 CreateOrder 流程可能需要
		// missing = append(missing, string(CtxKeySecToken))
		secToken = "" // 如果未找到则默认为空，让 CreateOrder 处理需求
	} else {
		secToken = secTokenVal.(string)
	}
	if deviceIDVal == nil || deviceIDVal.(string) == "" {
		// 对于 Search 是可选的，后续 CreateOrder 流程可能需要
		// missing = append(missing, string(CtxKeyDeviceID))
		deviceID = "" // 默认为空，让 CreateOrder 处理需求
	} else {
		deviceID = deviceIDVal.(string)
	}

	if len(missing) > 0 {
		err = fmt.Errorf("missing required authentication tokens in context: %s", strings.Join(missing, ", "))
	}
	return
}

// Search 通过调用现有的 api.Get_airline_message 实现航班搜索逻辑。
func (s *flightService) Search(ctx context.Context, opt dto.SearchOption) (*dto.SearchResult, error) {
	s.logger.Info("Search request received", zap.Any("options", opt))

	// --- 获取认证令牌 ---
	userID, tcSecToken, _, _, err := getAuthTokensFromContext(ctx)
	if err != nil {
		// 仅检查 Search 所需的基本令牌
		if strings.Contains(err.Error(), string(CtxKeyTcUserID)) || strings.Contains(err.Error(), string(CtxKeyTcSecToken)) {
			s.logger.Error("Authentication tokens missing for flight search", zap.Error(err))
			return nil, fmt.Errorf("authentication required for flight search: %w", err)
		}
		// 如果需要，将其他缺失的令牌记录为警告，但继续进行 Search
		s.logger.Warn("Non-essential auth tokens missing, proceeding with search", zap.Error(err))
	}

	// --- 1. 为 api.Get_airline_message 准备 api.Options ---
	apiOpts := tongchengapi.NewOptions()
	apiOpts.Set("tcuserid", userID)
	apiOpts.Set("tcsectoken", tcSecToken)

	// --- 2. 通过注入的客户端调用 API 函数 ---
	resultJson, err := s.apiClient.Get_airline_message(&apiOpts) // 传递指针
	if err != nil {
		s.logger.Error("apiClient.Get_airline_message call failed", zap.Error(err))
		return nil, fmt.Errorf("flight search API call failed: %w", err) // 包装错误
	}
	if resultJson == "" {
		s.logger.Warn("api.Get_airline_message returned empty result")
		return &dto.SearchResult{Flights: []dto.FlightInfo{}, Total: 0}, nil // 返回空结果集
	}

	s.logger.Debug("api.Get_airline_message raw result", zap.String("json", resultJson))

	// --- 3. 解析 JSON 结果并映射到 DTO ---
	if !gjson.Valid(resultJson) {
		s.logger.Error("Invalid JSON received from flight search API", zap.String("result", resultJson))
		return nil, fmt.Errorf("invalid response format from flight search API")
	}
	jsonParse := gjson.Parse(resultJson)

	// 检查 API 级别的成功指示符
	// 已验证路径：顶层的 'success' 布尔值
	if !jsonParse.Get("success").Bool() {
		// 已验证路径假设：'message' 包含错误详情
		errMsg := jsonParse.Get("message").String()
		if errMsg == "" {
			errMsg = "Unknown API error" // 如果路径缺失则使用默认消息
		}
		s.logger.Error("Flight search API indicated failure", zap.String("message", errMsg), zap.String("rawResponse", resultJson))
		return nil, fmt.Errorf("flight search failed: %s", errMsg)
	}

	searchResult := &dto.SearchResult{
		Flights: make([]dto.FlightInfo, 0),
	}

	// 提取航班列表 - 已验证路径：'data.fl'（注意：原始注释中为 'data.fls'）
	flightsArray := jsonParse.Get("data.fl")
	if !flightsArray.Exists() || !flightsArray.IsArray() {
		s.logger.Warn("No flight array ('data.fl') found in API response or invalid format.")
		return searchResult, nil // 未找到航班不视为错误
	}

	searchResult.Total = len(flightsArray.Array())

	for i, flightJson := range flightsArray.Array() {
		// 验证：时间格式假设 - 检查这是否与实际 API 响应格式匹配。
		layout := "2006-01-02 15:04:05"
		// 已验证路径：dt (起飞时间)
		depTimeStr := flightJson.Get("dt").String()
		depTime, depErr := time.Parse(layout, depTimeStr)
		// 已验证路径：at (到达时间)
		arrTimeStr := flightJson.Get("at").String()
		arrTime, arrErr := time.Parse(layout, arrTimeStr)

		if depErr != nil || arrErr != nil {
			s.logger.Warn("Failed to parse flight times, skipping flight",
				zap.Int("index", i),
				zap.String("flightNum", flightJson.Get("fn").String()), // fn 路径已验证
				zap.String("depTimeStr", depTimeStr),
				zap.String("arrTimeStr", arrTimeStr),
				zap.Error(depErr), zap.Error(arrErr))
			continue // 跳过此航班记录
		}

		// 已验证路径：fn, asn/amn, dac, aac, lps.0.sp
		// 'id' 路径需要根据实际响应进行验证，暂时使用航班号。
		flight := dto.FlightInfo{
			// ID:            flightJson.Get("id").String(), // 验证：路径 'id' 需要确认
			ID:            flightJson.Get("fn").String(), // 使用航班号作为临时 ID
			FlightNumber:  flightJson.Get("fn").String(),
			Airline:       flightJson.Get("asn").String(), // 使用 'asn' (航空公司名称) 或 'amn' (简称)
			DepartureTime: depTime,
			ArrivalTime:   arrTime,
			Origin:        flightJson.Get("dac").String(),     // 使用 'dac' (出发机场代码)
			Destination:   flightJson.Get("aac").String(),     // 使用 'aac' (到达机场代码)
			Price:         flightJson.Get("lps.0.sp").Float(), // 使用 'lps.0.sp' (经济舱最低价)
			Currency:      "CNY",                              // 假设为 CNY 或验证路径
		}
		searchResult.Flights = append(searchResult.Flights, flight)
	}

	s.logger.Info("Flight search successful", zap.Int("flightsFound", len(searchResult.Flights)), zap.Int("totalReported", searchResult.Total))
	return searchResult, nil
}

// CreateOrder 通过调用 api.CreateOrder 实现航班订单创建逻辑
func (s *flightService) CreateOrder(ctx context.Context, req dto.OrderRequest) (*dto.OrderResponse, error) {
	s.logger.Info("CreateOrder request received", zap.Any("request", req))

	// --- 获取认证令牌 ---
	userID, tcSecToken, secToken, deviceID, err := getAuthTokensFromContext(ctx)
	if err != nil {
		// 根据 requtst.go 分析，CreateOrder 流程可能需要所有令牌
		s.logger.Error("Authentication or session tokens missing for order creation", zap.Error(err))
		// 检查具体缺少哪些令牌以提供更精确的错误
		var requiredMissing []string
		if userID == "" {
			requiredMissing = append(requiredMissing, string(CtxKeyTcUserID))
		}
		if tcSecToken == "" {
			requiredMissing = append(requiredMissing, string(CtxKeyTcSecToken))
		}
		if secToken == "" {
			requiredMissing = append(requiredMissing, string(CtxKeySecToken))
		}
		if deviceID == "" {
			requiredMissing = append(requiredMissing, string(CtxKeyDeviceID))
		}

		if len(requiredMissing) > 0 {
			return nil, fmt.Errorf("missing required tokens for order creation: %s", strings.Join(requiredMissing, ", "))
		}
		// 如果 err 不为 nil 但令牌似乎存在，则记录为警告
		s.logger.Warn("Error retrieving tokens, but proceeding", zap.Error(err))
	}

	// --- 1. 为 api.CreateOrder 准备 Options
	passengerApiOpts := tongchengapi.NewOptions()

	// --- 设置认证/会话令牌（用于请求头）---
	passengerApiOpts.Set("tcuserid", userID)
	passengerApiOpts.Set("tcsectoken", tcSecToken)
	passengerApiOpts.Set("sec_token", secToken)
	passengerApiOpts.Set("deviceId", deviceID)

	passengerApiOpts.Set("OrderSerialId", req.OrderSerialId)
	// 使用 DTO 中的 Code 和 PromotionSign
	passengerApiOpts.Set("code", req.Code)
	passengerApiOpts.Set("promotionSign", req.PromotionSign)

	passengerApiOpts.Set("mobile", req.ContactPhone) // 假设 ContactPhone 映射到 API 的 'LinkMobile'
	passengerApiOpts.Set("LinkMan", req.ContactName) // API 使用硬编码的 "许玉鹏"，在此处设置可能会被忽略。

	// --- 遍历乘客并为每个乘客调用 API ---
	// 注意：底层的 tongchengapi.CreateOrder 每次调用仅支持一位乘客。
	// 此循环遍历请求中的乘客，并为每位乘客单独调用 API。
	orderResponse := &dto.OrderResponse{
		PassengerResults: make([]dto.PassengerOrderResult, 0, len(req.Passengers)),
	}
	successfulOrders := 0
	var accumulatedErrors []string // 存储错误用于日志记录/摘要

	if len(req.Passengers) == 0 {
		s.logger.Error("No passengers provided in the request for CreateOrder")
		return nil, fmt.Errorf("at least one passenger is required to create an order")
	}

	for i, p := range req.Passengers {
		s.logger.Info("Processing passenger", zap.Int("index", i), zap.String("name", p.Name))

		// 为每个乘客创建一个 *新的* api.Options 以避免覆盖
		passengerApiOpts := tongchengapi.NewOptions()

		// --- 设置认证/会话令牌（此请求中所有乘客相同）---
		passengerApiOpts.Set("tcuserid", userID)
		passengerApiOpts.Set("tcsectoken", tcSecToken)
		passengerApiOpts.Set("sec_token", secToken)
		// deviceID 在 API 中是硬编码的

		// --- 设置订单上下文参数（所有乘客相同）---
		passengerApiOpts.Set("OrderSerialId", req.OrderSerialId)
		passengerApiOpts.Set("code", req.Code)
		passengerApiOpts.Set("promotionSign", req.PromotionSign)

		// --- 设置联系人信息（所有乘客相同）---
		passengerApiOpts.Set("mobile", req.ContactPhone)
		// LinkMan 在 API 中是硬编码的

		// --- 设置当前乘客信息 ---
		passengerApiOpts.Set("passenger_name", p.Name)
		passengerApiOpts.Set("passenger_no", p.IDNumber) // 用于 certNo
		passengerApiOpts.Set("birthday", p.Birthday)
		passengerApiOpts.Set("gender", p.Gender)
		passengerApiOpts.Set("passenger_idcard", p.IDNumber) // 用于 passId

		// --- 记录选项（屏蔽敏感令牌）---
		// logOpts := passengerApiOpts.Clone()
		// logOpts.Delete("tcsectoken")
		// logOpts.Delete("sec_token")
		// s.logger.Debug("Calling api.CreateOrder for passenger", zap.String("passengerName", p.Name), zap.Any("apiOptions", logOpts))

		// --- 通过注入的客户端为当前乘客调用 API ---
		resultJson, err := s.apiClient.CreateOrder(&passengerApiOpts) // 传递指针
		passengerResult := dto.PassengerOrderResult{
			PassengerName: p.Name,
		}

		if err != nil {
			s.logger.Error("api.CreateOrder call failed for passenger", zap.String("passengerName", p.Name), zap.Error(err))
			passengerResult.Success = false
			passengerResult.ErrorMessage = fmt.Sprintf("API call failed: %v", err)
			accumulatedErrors = append(accumulatedErrors, fmt.Sprintf("%s: %s", p.Name, passengerResult.ErrorMessage))
		} else if resultJson == "" {
			s.logger.Warn("api.CreateOrder returned empty result for passenger", zap.String("passengerName", p.Name))
			passengerResult.Success = false
			passengerResult.ErrorMessage = "API returned empty response"
			accumulatedErrors = append(accumulatedErrors, fmt.Sprintf("%s: %s", p.Name, passengerResult.ErrorMessage))
		} else {
			s.logger.Debug("api.CreateOrder raw result for passenger", zap.String("passengerName", p.Name), zap.String("json", resultJson))
			if !gjson.Valid(resultJson) {
				s.logger.Error("Invalid JSON received from create order API for passenger", zap.String("passengerName", p.Name), zap.String("result", resultJson))
				passengerResult.Success = false
				passengerResult.ErrorMessage = "Invalid API response format"
				accumulatedErrors = append(accumulatedErrors, fmt.Sprintf("%s: %s", p.Name, passengerResult.ErrorMessage))
			} else {
				jsonParse := gjson.Parse(resultJson)
				if jsonParse.Get("success").Bool() {
					passengerResult.Success = true
					// 尝试解析 OrderID - 路径需要验证
					passengerResult.OrderID = jsonParse.Get("data.OrderInfo.OrderId").String()
					successfulOrders++
					s.logger.Info("Order creation successful for passenger", zap.String("passengerName", p.Name), zap.String("orderId", passengerResult.OrderID))
				} else {
					passengerResult.Success = false
					// 尝试获取错误消息
					passengerResult.ErrorMessage = jsonParse.Get("message").String()
					if passengerResult.ErrorMessage == "" {
						passengerResult.ErrorMessage = jsonParse.Get("msg").String()
					}
					if passengerResult.ErrorMessage == "" {
						passengerResult.ErrorMessage = "API reported failure (unknown reason)"
					}
					s.logger.Error("Order creation API reported failure for passenger",
						zap.String("passengerName", p.Name),
						zap.String("message", passengerResult.ErrorMessage),
						zap.String("rawResponse", resultJson))
					accumulatedErrors = append(accumulatedErrors, fmt.Sprintf("%s: %s", p.Name, passengerResult.ErrorMessage))
				}

			}
		}
		orderResponse.PassengerResults = append(orderResponse.PassengerResults, passengerResult)
	} // 结束乘客循环

	// --- 设置总体成功状态和消息 ---
	orderResponse.Success = successfulOrders == len(req.Passengers)
	if orderResponse.Success {
		orderResponse.Message = fmt.Sprintf("Successfully created orders for all %d passengers.", len(req.Passengers))
	} else {
		orderResponse.Message = fmt.Sprintf("Created orders for %d out of %d passengers. Failures: %s",
			successfulOrders, len(req.Passengers), strings.Join(accumulatedErrors, "; "))
		// 可选：如果 *任何* 乘客失败则返回错误，具体取决于期望的行为。
		// 目前，我们在响应中返回部分成功/失败的详细信息。
		// return nil, fmt.Errorf("未能为部分乘客创建订单：%s", strings.Join(accumulatedErrors, "; "))
	}

	s.logger.Info("Finished processing CreateOrder request",
		zap.Int("totalPassengers", len(req.Passengers)),
		zap.Int("successfulOrders", successfulOrders),
		zap.Bool("overallSuccess", orderResponse.Success))

	return orderResponse, nil
}
