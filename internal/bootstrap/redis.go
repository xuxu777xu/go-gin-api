package bootstrap

import (
	"context"
	"time"

	"myGin/internal/conf" // 模块路径

	"github.com/redis/go-redis/v9" // 使用 go-redis/v9
	"go.uber.org/zap"
)

// InitRedis 根据配置初始化 Redis 客户端连接。
// 返回 Redis 客户端实例、一个清理函数以及可能发生的任何错误。
func InitRedis(cfg conf.RedisConfig) (*redis.Client, func(), error) {
	// 如果 Redis 未启用，则安全地返回 nil。
	if !cfg.Enable {
		GetLogger().Info("Redis is disabled in config") // 使用 GetLogger()
		cleanup := func() {} // 空操作清理函数
		return nil, cleanup, nil
	}

	GetLogger().Info("Initializing Redis connection", zap.String("addr", cfg.Addr), zap.Int("db", cfg.DB)) // 使用 GetLogger()

	// 创建 Redis 客户端选项
	opts := &redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password, // "" 表示没有密码
		DB:       cfg.DB,       // 使用默认数据库
		// 如果需要，添加其他选项（例如，PoolSize、ReadTimeout、DialTimeout）
	       DialTimeout: 5 * time.Second, // 示例：添加拨号超时
	}

	// 创建 Redis 客户端
	client := redis.NewClient(opts)

	// Ping Redis 服务器以验证连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // 带超时的上下文
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		GetLogger().Error("Failed to ping Redis server", zap.String("addr", cfg.Addr), zap.Error(err)) // 使用 GetLogger()
		// 即使 ping 失败也尝试关闭客户端？可能没有必要或可能阻塞。
		// 为简单起见，直接返回错误。调用者 (main.go) 应处理此问题。
		return nil, func() {}, err // 在这种情况下返回错误以及一个空操作清理函数
	}

	GetLogger().Info("Redis connection established successfully", zap.String("addr", cfg.Addr)) // 使用 GetLogger()

	// 定义清理函数以关闭 Redis 客户端连接
	cleanup := func() {
		GetLogger().Info("Closing Redis connection", zap.String("addr", cfg.Addr)) // 使用 GetLogger()
		if err := client.Close(); err != nil {
			GetLogger().Error("Failed to close Redis connection", zap.String("addr", cfg.Addr), zap.Error(err)) // 使用 GetLogger()
		}
	}

	return client, cleanup, nil
}