package bootstrap

import (
	"fmt"
	"os"
	"path/filepath" // 确保导入

	"github.com/spf13/viper"
	"myGin/internal/conf" // 使用 go.mod 中定义的模块路径 tongcheng
)

// LoadConfig 从指定路径加载配置
// 默认尝试从运行目录下的 ./configs/config.yml 加载
func LoadConfig() (*conf.Config, error) {
	v := viper.New()
	v.SetConfigName("config") // 配置文件名 (不带扩展名)
	v.SetConfigType("yaml")   // 文件类型

	// 添加搜索路径，viper 会按顺序查找
	v.AddConfigPath("./configs")      // 运行目录下的 configs
	v.AddConfigPath("../configs")     // 上一级目录的 configs (例如从 cmd/server 运行)
	v.AddConfigPath("../../configs") // 再上一级目录的 configs

	// 尝试读取配置文件
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// 配置文件未找到
			return nil, fmt.Errorf("config file 'config.yaml' not found in search paths: %w", err)
		}
		// 其他读取错误
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 设置默认值 (viper 会在 Unmarshal 前应用它们，如果文件中未指定)
	// v.SetDefault("server.addr", ":8080")
	// ...

	var cfg conf.Config
	// 将配置 unmarshal 到结构体
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 可选: 验证配置项 (例如，确保某些必需字段不为空)

	// 创建日志文件所在的目录（如果需要且不存在）
	if cfg.Logger.File != "" {
		logDir := filepath.Dir(cfg.Logger.File)
        // 检查目录是否存在，不存在则创建
		if _, statErr := os.Stat(logDir); os.IsNotExist(statErr) {
            // 0750 权限: user=rwx, group=rx, other=---
			if mkErr := os.MkdirAll(logDir, 0750); mkErr != nil {
				// 这里暂时打印警告，后续应使用 logger
				fmt.Fprintf(os.Stderr, "Warning: could not create log directory '%s': %v\n", logDir, mkErr)
			}
		}
	}


	return &cfg, nil
}