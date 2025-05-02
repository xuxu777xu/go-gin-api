package service

import (
	"fmt"
	"myGin/internal/pkg/tongchengapi" // 更新了导入路径
)

// --- TongchengAPIClient 的生产环境实现 ---

type tongchengAPIClientImpl struct{}

// NewTongchengAPIClientImpl 创建一个新的生产环境客户端。
func NewTongchengAPIClientImpl() TongchengAPIClient {
	return &tongchengAPIClientImpl{}
}

// Get_airline_message 调用实际的同程 API 函数。
func (c *tongchengAPIClientImpl) Get_airline_message(opts *tongchengapi.Options) (string, error) {
	// 注意：原始函数接受 tongchengapi.Options，而不是 *tongchengapi.Options。
	// 我们需要在此处解引用指针。
	if opts == nil {
		// 服务层在调用此函数之前确保 opts 不为 nil。
		// 如果它为 nil，底层的 API 调用可能会 panic。
		// 如果需要，考虑添加显式的 nil 检查和错误返回。
		return "", fmt.Errorf("internal error: opts cannot be nil for Get_airline_message")
	}
	return tongchengapi.Get_airline_message(*opts) // 解引用指针
}

// CreateOrder 调用实际的同程 API 函数。
func (c *tongchengAPIClientImpl) CreateOrder(opts *tongchengapi.Options) (string, error) {
	// 需要与上面类似的解引用。
	if opts == nil {
		// 请参阅上面 Get_airline_message 中的注释。
		return "", fmt.Errorf("internal error: opts cannot be nil for CreateOrder")
	}
	return tongchengapi.CreateOrder(*opts) // 解引用指针
}
