package bootstrap

import (
	"os"
	"sync"
	"time" // 确保导入 time 包，如果 encoderConfig 中使用了时间相关的设置

	"myGin/internal/conf" // 使用 go.mod 模块路径
	// 移除对自身的导入 "myGin/internal/bootstrap"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	logInstance *zap.Logger
	once        sync.Once
)

// InitializeLogger 使用提供的配置初始化全局日志记录器单例
// 这个函数应该在应用程序启动时，加载配置之后被调用一次
func InitializeLogger(cfg conf.LoggerConfig) {
	once.Do(func() {
		logInstance = initLoggerInternal(cfg)
		logInstance.Info("Logger initialized successfully via InitializeLogger",
			zap.String("level", cfg.Level),
			zap.String("file", cfg.File))
	})
}

// GetLogger 返回已初始化的日志记录器单例实例
// 如果在调用此函数前没有调用 InitializeLogger，将会 panic
func GetLogger() *zap.Logger {
	if logInstance == nil {
		// 强制要求在使用前必须初始化
		panic("Logger has not been initialized. Call bootstrap.InitializeLogger() first.")
	}
	return logInstance
}

// initLoggerInternal 根据提供的配置初始化 zap logger (内部使用)
func initLoggerInternal(cfg conf.LoggerConfig) *zap.Logger {
	// 配置 lumberjack 进行日志切割
	lumberjackLogger := &lumberjack.Logger{
		Filename:   cfg.File,       // 日志文件路径
		MaxSize:    cfg.MaxSize,    // 每个日志文件的最大大小，单位：MB
		MaxBackups: cfg.MaxBackups, // 保留的旧日志文件的最大数量
		MaxAge:     cfg.MaxAge,     // 旧日志文件保留的最大天数
		Compress:   cfg.Compress,   // 是否压缩旧日志文件
	}

	// 配置 zap encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	// 自定义时间编码器，使其更易读 (可选)
	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
	}
	// encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder // 或者使用标准的 ISO8601
	encoderConfig.TimeKey = "time"                       // 时间字段名
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder // 大写日志级别 (例如, INFO, ERROR)
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder // 短路径调用者信息

	// 创建 JSON encoder
	jsonEncoder := zapcore.NewJSONEncoder(encoderConfig)

	// 解析日志级别字符串
	logLevel := zapcore.InfoLevel // 默认级别
	if err := logLevel.Set(cfg.Level); err != nil {
		// 如果级别设置无效，使用默认 Info 级别并打印警告
		println("Warning: invalid log level '"+cfg.Level+"' configured, using default 'info'. Error: ", err.Error())
		logLevel = zapcore.InfoLevel
	}

	// 创建 core：同时写入文件和控制台
	// 文件写入 core
	fileSyncer := zapcore.AddSync(lumberjackLogger)
	fileCore := zapcore.NewCore(jsonEncoder, fileSyncer, logLevel)

	// 控制台写入 core
	consoleSyncer := zapcore.AddSync(os.Stdout)
	// 如果希望控制台输出更易读，可以使用 ConsoleEncoder
	// consoleEncoderConfig := encoderConfig
	// consoleEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // 带颜色的级别
	// consoleEncoder := zapcore.NewConsoleEncoder(consoleEncoderConfig)
	// consoleCore := zapcore.NewCore(consoleEncoder, consoleSyncer, logLevel)
	// --- 或者保持 JSON 输出到控制台 ---
	consoleCore := zapcore.NewCore(jsonEncoder, consoleSyncer, logLevel)

	// 合并 core
	teeCore := zapcore.NewTee(fileCore, consoleCore)

	// 创建 logger
	// zap.AddCaller() 添加调用者信息 (文件名:行号)
	// zap.AddStacktrace(zapcore.ErrorLevel) 在 Error 及以上级别添加堆栈信息
	// zap.Development() 在开发模式下，DPanicLevel 会 panic，方便调试
	// 可根据环境判断是否启用 Development 模式
	// if isDevelopment {
	// 	logger = zap.New(teeCore, zap.AddCaller(), zap.Development(), zap.AddStacktrace(zapcore.ErrorLevel))
	// } else {
	var logger = zap.New(teeCore, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	// }

	// 不再替换全局 logger
	// zap.ReplaceGlobals(logger)
	// 不再在这里记录初始化，移到 GetLogger 的 once.Do 中
	// logger.Info("Logger initialized successfully", zap.String("level", cfg.Level), zap.String("file", cfg.File))

	return logger
}