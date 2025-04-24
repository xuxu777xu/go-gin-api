package main

import (
	"context" // 导入 context 包
	"fmt"     // 用于日志记录器初始化前的错误输出
	"net/http"
	"os"      // 用于 os.Exit 和 os.Signal
	"os/signal" // 用于信号处理
	"syscall" // 用于 syscall.SIGTERM
	"time"    // 用于 context 超时

	"github.com/gin-gonic/gin"
	// "github.com/redis/go-redis/v9" // 如果需要进行 map 值类型断言，请显式导入 redis
	"go.uber.org/zap"
	// "gorm.io/gorm" // 如果需要进行 map 值类型断言，请显式导入 gorm

	"myGin/internal/bootstrap"
	// 使用匿名导入来解决 "imported and not used" 的 Linter 错误
	// 这表明我们需要包的类型定义，即使不直接引用包名。
	_ "myGin/internal/conf"
)

func main() {
	// 1. 加载配置
	cfg, err := bootstrap.LoadConfig() // LoadConfig 返回 *conf.Config, err
	if err != nil {
		// 日志记录器尚未初始化，直接输出到 stderr 并退出
		fmt.Fprintf(os.Stderr, "FATAL: Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// 2. 初始化日志记录器
	// 传递 cfg.Logger (类型为 conf.LoggerConfig) 而不是整个 cfg
	bootstrap.InitializeLogger(cfg.Logger) // 调用新的初始化函数
	// 不再需要本地 zlog 变量
	// 可选：不再需要替换全局 logger
	// zap.ReplaceGlobals(zlog)
	// 确保在程序退出时刷新缓冲的日志条目
	// 在 defer 中获取 logger 实例
	defer func() { _ = bootstrap.GetLogger().Sync() }() // 使用 defer 确保 Sync 被调用

	// 3. 初始化数据库连接 (如果启用)
	db, dbCleanup, err := bootstrap.InitDB(cfg.Database) // 传递数据库特定配置
	if err != nil {
		// 根据需要处理错误，如果数据库是可选的，可以只记录错误
		bootstrap.GetLogger().Error("Failed to initialize database", zap.Error(err)) // 使用 GetLogger()
		// 如果数据库是必需的，可以在这里 panic 或 os.Exit
		// os.Exit(1)
	}
	if dbCleanup != nil {
		defer dbCleanup() // 注册清理函数
	}

	// 4. 初始化 Redis 连接 (如果启用)
	rdb, redisCleanup, err := bootstrap.InitRedis(cfg.Redis) // 传递 Redis 特定配置
	if err != nil {
		// 根据需要处理错误，如果 Redis 是可选的，可以只记录错误
		bootstrap.GetLogger().Error("Failed to initialize Redis", zap.Error(err)) // 使用 GetLogger()
		// 如果 Redis 是必需的，可以在这里 panic 或 os.Exit
		// os.Exit(1)
	}
	if redisCleanup != nil {
		defer redisCleanup() // 注册清理函数
	}

	// 5. 创建 Gin 引擎
	engine := gin.New()

	// 6. 附加核心中间件 (Recovery, Logging 等)
	// 使用正确的函数签名，只传递 engine 和 cfg
	bootstrap.AttachCoreMiddleware(engine, cfg)

	// 7. 准备插件依赖项
	dependencies := make(map[string]interface{})
	dependencies["logger"] = bootstrap.GetLogger() // 传递 logger 实例
	if db != nil {
		dependencies["db"] = db // 只在成功初始化时传递 DB
	}
	if rdb != nil {
		dependencies["redis"] = rdb // 只在成功初始化时传递 Redis
	}
	// 可以添加其他共享依赖项...

	// 8. 附加可选插件
	// 使用新的签名传递依赖 map
	bootstrap.AttachPlugins(engine, cfg, dependencies)

	// 9. 注册路由 (步骤编号顺延)
	bootstrap.RegisterRoutes(engine, cfg)

	// 10. 创建 HTTP 服务器 (步骤编号顺延)
	srv := &http.Server{
		Addr:    cfg.Server.Addr, // 从配置中获取监听地址
		Handler: engine,
	}

	// 11. 启动 HTTP 服务器 (goroutine)
	bootstrap.GetLogger().Info("Server starting", zap.String("address", srv.Addr)) // 使用 GetLogger()
	go func() {
		// 服务连接
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			bootstrap.GetLogger().Fatal("Failed to run server", zap.Error(err)) // 使用 GetLogger()
		}
	}()

	// 12. 等待中断信号以优雅地关闭服务器（设置 10 秒超时）
	quit := make(chan os.Signal, 1) // 创建一个接收信号的通道
	/*第一种关停方式：使用 Ctrl+C
	第二种：使用 ps aux | grep main  查找main.go 进程
	然后，使用 kill <PID> 命令
	kill <PID>
	优雅关停。
	kill -2 <PID>
	Ctrl+C 发送的信号是同一个
	kill -9 <PID
	绕过所有的优雅关停逻辑，直接强制退出程序。
	*/ 
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // 监听 SIGINT 和 SIGTERM
	<-quit                                               // 阻塞，直到接收到信号
	bootstrap.GetLogger().Info("Shutting down server...") // 使用 GetLogger()

	// 创建一个 10 秒超时的 context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel() // 确保在 main 函数退出前调用 cancel，释放资源

	// 调用 Shutdown 进行优雅关停
	if err := srv.Shutdown(ctx); err != nil {
		// 如果 Shutdown 返回错误，记录 Fatal 级别的日志
		bootstrap.GetLogger().Fatal("Server forced to shutdown:", zap.Error(err)) // 使用 GetLogger()
	}

	bootstrap.GetLogger().Info("Server exiting") // 使用 GetLogger() 记录服务器成功退出的信息
}