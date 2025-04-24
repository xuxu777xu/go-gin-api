package bootstrap

import (
	"fmt"
	"myGin/internal/conf"
	"myGin/internal/plugin"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// PluginFactory 定义了创建插件实例的函数签名。
type PluginFactory func() plugin.Plugin

// registry 保存插件名称到工厂函数的映射。
var registry = map[string]PluginFactory{
	"ratelimit": plugin.NewRateLimitPlugin,
	"auth":      plugin.NewAuthPlugin,
	// "swagger":   plugin.NewSwagger,
}

// AttachPlugins 遍历配置中启用的模块，创建、初始化并注册插件。
func AttachPlugins(engine *gin.Engine, cfg *conf.Config, dependencies map[string]interface{}) {
	GetLogger().Info("Initializing and registering enabled plugins...") // 使用 GetLogger()
	enabledCount := 0
	registeredCount := len(registry)

	// 确保 logger 在依赖项中
	if _, ok := dependencies["logger"]; !ok {
		GetLogger().Warn("Logger not found in dependencies map, adding default Nop logger.") // 使用 GetLogger()
		dependencies["logger"] = zap.NewNop() // 保留 zap.NewNop() 作为后备
	}
	// 注意：config 不再需要作为通用依赖传递给 Init，因为每个插件会接收其特定的配置部分。

	// --- 速率限制插件 ---
	if cfg.Modules.RateLimit.Enable {
		// 传递 RateLimitConfig 部分
		handlePluginLifecycle(engine, "ratelimit", &cfg.Modules.RateLimit, dependencies, &enabledCount)
	} else {
		GetLogger().Debug("插件在配置中被禁用", zap.String("pluginName", "ratelimit")) // 使用 GetLogger()
	}

	// --- 认证插件 ---
	if cfg.Modules.Auth.Enable {
		// 传递 AuthConfig 部分
		handlePluginLifecycle(engine, "auth", &cfg.Modules.Auth, dependencies, &enabledCount)
	} else {
		GetLogger().Debug("插件在配置中被禁用", zap.String("pluginName", "auth")) // 使用 GetLogger()
	}

	// --- 类似地添加其他插件 ---
	// if cfg.Modules.Swagger.Enable {
	//     handlePluginLifecycle(engine, "swagger", &cfg.Modules.Swagger, dependencies, &enabledCount)
	// } else {
	//     GetLogger().Debug("插件在配置中被禁用", zap.String("pluginName", "swagger")) // 使用 GetLogger()
	// }

	if enabledCount == 0 {
		GetLogger().Info("No plugins were enabled or registered.") // 使用 GetLogger()
	} else {
		GetLogger().Info("Finished initializing and registering plugins", zap.Int("registeredCount", enabledCount), zap.Int("availableCount", registeredCount)) // 使用 GetLogger()
	}
}

// handlePluginLifecycle 处理单个插件的创建、初始化和注册。
// moduleCfg 是该插件特定的配置结构指针 (e.g., *conf.AuthConfig)。
func handlePluginLifecycle(engine *gin.Engine, name string, moduleCfg interface{}, dependencies map[string]interface{}, enabledCount *int) {
	factory, exists := registry[name]
	if !exists {
		GetLogger().Warn("插件在配置中启用，但在静态注册表中未找到工厂函数，跳过。", // 使用 GetLogger()
			zap.String("pluginName", name))
		return
	}

	GetLogger().Info("Processing plugin...", zap.String("pluginName", name)) // 使用 GetLogger()

	var pluginInstance plugin.Plugin
	var err error

	// 1. 创建实例
	func() {
		defer func() {
			if r := recover(); r != nil {
				GetLogger().Error("Panic recovered during plugin creation", // 使用 GetLogger()
					zap.String("pluginName", name),
					zap.Any("panicValue", r),
					zap.Stack("stacktrace"),
				)
				pluginInstance = nil // 确保实例为 nil
			}
		}()
		pluginInstance = factory()
	}()

	if pluginInstance == nil {
		GetLogger().Error("创建尝试后插件实例为 nil，跳过。", zap.String("pluginName", name)) // 使用 GetLogger()
		return
	}

	// 2. 初始化插件
	func() {
		defer func() {
			if r := recover(); r != nil {
				GetLogger().Error("Panic recovered during plugin initialization", // 使用 GetLogger()
					zap.String("pluginName", name),
					zap.Any("panicValue", r),
					zap.Stack("stacktrace"),
				)
				err = fmt.Errorf("panic during initialization: %v", r)
			}
		}()
		// 将特定模块配置和共享依赖传递给 Init
		err = pluginInstance.Init(moduleCfg, dependencies)
	}()

	if err != nil {
		GetLogger().Error("初始化插件失败，跳过注册。", // 使用 GetLogger()
			zap.String("pluginName", name),
			zap.Error(err),
		)
		return
	}
	GetLogger().Debug("插件初始化成功", zap.String("pluginName", name)) // 使用 GetLogger()

	// 3. 注册插件 (例如，注册中间件或路由)
	func() {
		defer func() {
			if r := recover(); r != nil {
				GetLogger().Error("Panic recovered during plugin registration", // 使用 GetLogger()
					zap.String("pluginName", name),
					zap.Any("panicValue", r),
					zap.Stack("stacktrace"),
				)
				err = fmt.Errorf("panic during registration: %v", r)
			}
		}()
		err = pluginInstance.Register(engine)
	}()

	if err != nil {
		GetLogger().Error("注册插件失败。", // 使用 GetLogger()
			zap.String("pluginName", name),
			zap.Error(err),
		)
		return // 注册失败，不增加计数
	}

	*enabledCount++
	GetLogger().Info("成功初始化并注册插件", zap.String("pluginName", name)) // 使用 GetLogger()
}