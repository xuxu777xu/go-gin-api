package plugin

import (
	"fmt"
	"myGin/internal/conf"

	"github.com/gin-gonic/gin"
)

// Plugin 是所有插件必须实现的接口。
// 插件的生命周期：New -> Init -> Register
type Plugin interface {
	// Init 用于初始化插件，通常在此处处理配置和依赖注入。
	// cfg 参数是特定于该插件的配置结构 (例如 *conf.AuthConfig)。
	// deps 包含共享的依赖项，如 logger, db 等。
	Init(cfg interface{}, deps map[string]interface{}) error

	// Register 用于将插件的功能（例如中间件、路由）注册到 Gin 引擎。
	Register(r *gin.Engine) error
}

// --- 类型断言辅助函数 (可选，但推荐) ---

// GetAuthConfig 从 interface{} 安全地获取 AuthConfig。
func GetAuthConfig(cfg interface{}) (*conf.AuthConfig, error) {
	authCfg, ok := cfg.(*conf.AuthConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type for AuthPlugin, expected *conf.AuthConfig, got %T", cfg)
	}
	return authCfg, nil
}

// GetRateLimitConfig 从 interface{} 安全地获取 RateLimitConfig。
func GetRateLimitConfig(cfg interface{}) (*conf.RateLimitConfig, error) {
	rateLimitCfg, ok := cfg.(*conf.RateLimitConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type for RateLimitPlugin, expected *conf.RateLimitConfig, got %T", cfg)
	}
	return rateLimitCfg, nil
}