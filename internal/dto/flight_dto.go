package dto

import "time"

// SearchOption 定义机票搜索的请求参数
// 使用 binding 和 validate 标签进行参数校验 (validate 标签需配合 validator 库使用)
type SearchOption struct {
	From     string    `json:"from" binding:"required,len=3"`           // 出发地 (例如："SHA") - binding 用于 Gin 参数绑定
	To       string    `json:"to" binding:"required,len=3"`               // 目的地 (例如："PEK")
	// 使用 string 接收日期，方便 Gin 绑定和初始校验格式
	Date     string    `json:"date" binding:"required,datetime=2006-01-02"` // 日期 (格式：YYYY-MM-DD) - datetime 标签用于 Gin v1.9+ 的 validator v10
	// --- 可选参数示例 ---
	CabinClass string `json:"cabinClass,omitempty"` // 舱位等级 (例如："Economy", "Business")
	// Passengers int    `json:"passengers,omitempty" binding:"omitempty,gte=1"` // 乘客人数, omitempty 表示可选, gte=1 表示大于等于1
}

// --- 以下为其他 DTO 骨架 ---

// SearchResult 代表机票搜索结果 (示例结构)
type SearchResult struct {
	Flights []FlightInfo `json:"flights"`
	Total   int          `json:"total"` // 搜索到的总航班数（如果需要分页）
	// Query   SearchOption `json:"query"` // 返回搜索条件（可选）
}

// FlightInfo 代表单个航班信息 (示例结构)
type FlightInfo struct {
	ID            string    `json:"id"`           // 航班唯一标识符（可能用于后续订票）
	FlightNumber  string    `json:"flightNumber"` // 航班号
	Airline       string    `json:"airline"`      // 航空公司
	DepartureTime time.Time `json:"departureTime"`// 出发时间 (time.Time 类型方便处理)
	ArrivalTime   time.Time `json:"arrivalTime"`  // 到达时间
	Origin        string    `json:"origin"`       // 出发机场代码
	Destination   string    `json:"destination"`  // 到达机场代码
	Price         float64   `json:"price"`        // 价格
	Currency      string    `json:"currency"`     // 货币 (例如："CNY")
	// ... 其他航班详情, 例如：经停信息, 剩余座位数等
}

// OrderRequest 代表创建订单的请求 (示例结构)
type OrderRequest struct {
	FlightID      string       `json:"flightId" binding:"required"`      // 要预订的航班ID (来自搜索结果)
	Passengers    []Passenger `json:"passengers" binding:"required,min=1"` // 乘客信息列表 (至少一个)
	ContactName   string       `json:"contactName" binding:"required"`   // 联系人姓名
	ContactPhone  string       `json:"contactPhone" binding:"required"`  // 联系人电话
	OrderSerialId string       `json:"orderSerialId,omitempty"`         // 订单序列号 (从 Buildtemporder 获取)
	Code          string       `json:"code,omitempty"`                  // 优惠码/活动代码
	PromotionSign string       `json:"promotionSign,omitempty"`         // 促销标识 (与 Code 关联)
	// ... 其他必要信息, 如支付方式等
}

// Passenger 代表乘客信息 (示例结构)
type Passenger struct {
	Name       string `json:"name" binding:"required"`        // 姓名
	IDType     string `json:"idType" binding:"required"`      // 证件类型 (例如："IDCard", "Passport")
	IDNumber   string `json:"idNumber" binding:"required"`    // 证件号码
	Birthday   string `json:"birthday" binding:"omitempty,datetime=2006-01-02"` // 生日 (格式：YYYY-MM-DD)
	Gender     int    `json:"gender" binding:"omitempty,oneof=0 1 2"` // 性别 (0:未知, 1:男, 2:女) - 根据API调整，API似乎用1表示
}


// PassengerOrderResult 代表单个乘客的订单创建结果
type PassengerOrderResult struct {
	Success       bool   `json:"success"`       // 该乘客是否成功下单
	PassengerName string `json:"passengerName"` // 乘客姓名
	OrderID       string `json:"orderId,omitempty"` // 成功时的订单号
	ErrorMessage  string `json:"errorMessage,omitempty"` // 失败时的错误信息
}

// OrderResponse 代表创建订单的响应 (聚合了所有乘客的结果)
type OrderResponse struct {
	Success          bool                   `json:"success"`          // 整体操作是否完全成功 (所有乘客都成功)
	Message          string                 `json:"message"`          // 整体操作的消息总结
	PassengerResults []PassengerOrderResult `json:"passengerResults"` // 每个乘客的下单结果列表
}

// --- 可根据 FlightService 接口添加更多 DTO ---