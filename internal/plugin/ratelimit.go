package plugin

import (
	"fmt"
	"sync"
	"myGin/internal/conf"
	"myGin/internal/pkg/errs"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// RateLimitPlugin 实现了基于 IP 的请求速率限制插件。
// 注意：当前实现使用内存存储 (sync.Map) 来跟踪每个 IP 的限流器。
// 这适用于单实例部署或测试环境。在生产环境或分布式部署中，
// 应考虑使用 Redis 或其他分布式存储来维护限流状态，以确保跨实例的一致性。
type RateLimitPlugin struct {
	rateLimitCfg *conf.RateLimitConfig // 存储加载的限流配置
	logger       *zap.Logger
	ipLimiters   sync.Map // key: IP 地址 (string), value: *rate.Limiter
}

// NewRateLimitPlugin 创建一个新的 RateLimitPlugin 实例。
func NewRateLimitPlugin() Plugin {
	return &RateLimitPlugin{}
}

// Init 初始化 RateLimitPlugin。
func (p *RateLimitPlugin) Init(cfg interface{}, deps map[string]interface{}) error {
	// 1. 类型断言获取具体配置
	rateLimitCfg, err := GetRateLimitConfig(cfg)
	if err != nil {
		return fmt.Errorf("ratelimit plugin init failed: %w", err)
	}
	p.rateLimitCfg = rateLimitCfg

	// 2. 获取 logger 依赖
	loggerDep, ok := deps["logger"]
	if !ok {
		return fmt.Errorf("ratelimit plugin init failed: logger dependency is missing")
	}
	p.logger, ok = loggerDep.(*zap.Logger)
	if !ok {
		return fmt.Errorf("ratelimit plugin init failed: logger dependency is not of type *zap.Logger")
	}

	// 3. 检查是否启用
	if !p.rateLimitCfg.Enable {
		p.logger.Info("RateLimit Plugin is disabled by config.")
		return nil // 如果未启用，则无需执行任何操作
	}

	// 4. 检查配置有效性
	if p.rateLimitCfg.Rate <= 0 || p.rateLimitCfg.Burst <= 0 {
		p.logger.Error("RateLimit Plugin init failed: invalid rate or burst value in config",
			zap.Float64("rate", p.rateLimitCfg.Rate),
			zap.Int("burst", p.rateLimitCfg.Burst))
		return fmt.Errorf("ratelimit plugin init failed: rate and burst must be positive")
	}

	// 5. 初始化 IP 限流器存储 (确保只在初始化时执行一次)
	p.ipLimiters = sync.Map{}

	p.logger.Info("RateLimit Plugin initialized successfully.")
	return nil
}

// Register 将限流中间件注册到 Gin 引擎。
func (p *RateLimitPlugin) Register(r *gin.Engine) error {
	if !p.rateLimitCfg.Enable {
		return nil // 如果未启用，则不注册中间件
	}

	p.logger.Info("Registering RateLimit Plugin middleware...")

	// 将中间件应用到全局
	r.Use(p.rateLimitMiddleware())

	p.logger.Info("RateLimit Plugin middleware registered globally.")
	return nil
}

// rateLimitMiddleware 创建并返回限流中间件函数。
func (p *RateLimitPlugin) rateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取客户端 IP
		ip := c.ClientIP()
		if ip == "" {
			// 如果无法获取 IP，可以选择放行或记录警告
			p.logger.Warn("RateLimit middleware: Could not get client IP, allowing request")
			c.Next()
			return
		}

		// 获取或创建该 IP 对应的限流器
		// 使用 LoadOrStore 保证并发安全地获取或创建 Limiter
		limiterUntyped, _ := p.ipLimiters.LoadOrStore(ip, rate.NewLimiter(rate.Limit(p.rateLimitCfg.Rate), p.rateLimitCfg.Burst))
		limiter := limiterUntyped.(*rate.Limiter) // 类型断言

		// 检查是否允许请求
		if !limiter.Allow() {
			// 如果不允许，则拒绝请求
			p.logger.Warn("RateLimit middleware: Too many requests", zap.String("ip", ip))
			errs.TooManyRequests.JSON(c) // 使用 errs 包返回标准错误
			c.Abort()                    // 中断请求链
			return
		}

		// 如果允许，则继续处理请求
		c.Next()
	}
}