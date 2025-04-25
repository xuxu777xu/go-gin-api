package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"time"

	"myGin/internal/conf" // 模块路径

	"go.uber.org/zap"
	"gorm.io/driver/mysql" // 初始假设使用 MySQL
	// "gorm.io/driver/postgres" // 如果需要，添加其他驱动
	// "gorm.io/driver/sqlite"
	// "gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// InitDB 根据配置初始化数据库连接。
// 返回 GORM DB 实例、一个清理函数以及可能发生的任何错误。
func InitDB(cfg conf.DatabaseConfig) (*gorm.DB, func(), error) {
	// 如果数据库未启用，则安全返回 nil。
	if !cfg.Enable {
		GetLogger().Info("数据库在配置中被禁用") // 使用 GetLogger()
		cleanup := func() {} // 空操作清理
		return nil, cleanup, nil
	}

	GetLogger().Info("正在初始化数据库连接", zap.String("driver", cfg.Driver)) // 使用 GetLogger()

	var dialector gorm.Dialector
	switch cfg.Driver {
	case "mysql":
		// dsn := "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
		dialector = mysql.Open(cfg.DSN)
	// 如果需要，为其他驱动添加 case：
	// case "postgres":
	// 	dialector = postgres.Open(cfg.DSN)
	// case "sqlite":
	// 	dialector = sqlite.Open(cfg.DSN)
	// case "sqlserver":
	// 	dialector = sqlserver.Open(cfg.DSN)
	default:
		err := fmt.Errorf("不支持的数据库驱动: %s", cfg.Driver)
		GetLogger().Error("初始化数据库失败", zap.Error(err)) // 使用 GetLogger()
		return nil, func() {}, err
	}

	// 使用我们的 zap 包装器配置 GORM 日志记录器
	// 根据环境调整 LogLevel（例如，开发环境为 Info，生产环境为 Warn）
	gormLogCfg := gormlogger.Config{
		SlowThreshold:             200 * time.Millisecond, // 慢 SQL 阈值
		LogLevel:                  gormlogger.Warn,        // 日志级别 (Warn, Error, Info) - 根据需要调整
		IgnoreRecordNotFoundError: true,                   // 不将 'record not found' 错误记录为 Error 级别
		Colorful:                  false,                  // 禁用 JSON/结构化日志的彩色输出
	}
	gormZapLogger := newZapGormLogger(GetLogger(), gormLogCfg) // 使用 GetLogger()

	// 打开数据库连接
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: gormZapLogger,
		// 如果需要，添加其他 GORM 配置（例如，NamingStrategy, PrepareStmt）
	})
	if err != nil {
		GetLogger().Error("打开数据库连接失败", zap.String("driver", cfg.Driver), zap.Error(err)) // 使用 GetLogger()
		return nil, func() {}, fmt.Errorf("gorm.Open 失败: %w", err)
	}

	// 获取底层的 sql.DB 以进行连接池设置
	sqlDB, err := db.DB()
	if err != nil {
		GetLogger().Error("获取底层 sql.DB 失败", zap.Error(err)) // 使用 GetLogger()
		// 如果 db.DB() 失败，底层连接可能有问题。
		// GORM v2 没有在 *gorm.DB 上公开直接的 Close 方法。
		// 返回错误应该就足够了。
		return nil, func() {}, fmt.Errorf("db.DB() 失败: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)

	// 解析连接最大生命周期持续时间字符串
	if cfg.ConnMaxLifetime != "" {
		lifetime, err := time.ParseDuration(cfg.ConnMaxLifetime)
		if err != nil {
			GetLogger().Warn("解析 ConnMaxLifetime 失败，使用 GORM 默认值", // 使用 GetLogger()
				zap.String("value", cfg.ConnMaxLifetime),
				zap.Error(err))
		} else {
			sqlDB.SetConnMaxLifetime(lifetime)
		}
	} else {
        // 可选：如果配置中未提供，则设置默认生命周期
        // sqlDB.SetConnMaxLifetime(time.Hour)
    }

	// Ping 数据库以验证连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // 为 ping 添加超时
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		GetLogger().Error("Ping 数据库失败", zap.Error(err)) // 使用 GetLogger()
		_ = sqlDB.Close() // 尝试在 ping 失败时关闭
		return nil, func() {}, fmt.Errorf("sqlDB.Ping 失败: %w", err)
	}

	GetLogger().Info("数据库连接建立成功", zap.String("driver", cfg.Driver)) // 使用 GetLogger()

	// 定义用于关闭连接的清理函数
	cleanup := func() {
		GetLogger().Info("正在关闭数据库连接", zap.String("driver", cfg.Driver)) // 使用 GetLogger()
		if err := sqlDB.Close(); err != nil {
			GetLogger().Error("关闭数据库连接失败", zap.String("driver", cfg.Driver), zap.Error(err)) // 使用 GetLogger()
		}
	}

	return db, cleanup, nil
}

// zapGormLogger 使用 zap 实现了 gormlogger.Interface。
type zapGormLogger struct {
	zapLogger    *zap.Logger
	gormLogLevel gormlogger.LogLevel
	slowThreshold time.Duration
}

// newZapGormLogger 创建一个新的 GORM 日志记录器实例。
func newZapGormLogger(logger *zap.Logger, config gormlogger.Config) gormlogger.Interface { // 更新了签名以接受配置
	return &zapGormLogger{
		zapLogger:     logger.WithOptions(zap.AddCallerSkip(4)), // 调整跳过的调用层级以获取准确的调用者信息
		gormLogLevel:  config.LogLevel,
		slowThreshold: config.SlowThreshold, // 使用配置中的阈值
		// IgnoreRecordNotFoundError 在 Trace 方法中处理
	}
}

// LogMode 设置日志级别。
func (l *zapGormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newLogger := *l
	newLogger.gormLogLevel = level
	return &newLogger
}

// Info 记录信息性消息。
func (l *zapGormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.gormLogLevel >= gormlogger.Info {
		l.zapLogger.Sugar().Infow(msg, data...)
	}
}

// Warn 记录警告消息。
func (l *zapGormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.gormLogLevel >= gormlogger.Warn {
		l.zapLogger.Sugar().Warnw(msg, data...)
	}
}

// Error 记录错误消息。
func (l *zapGormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.gormLogLevel >= gormlogger.Error {
		l.zapLogger.Sugar().Errorw(msg, data...)
	}
}

// Trace 记录 SQL 查询和执行详情。
func (l *zapGormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.gormLogLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()
	fields := []zap.Field{
		zap.Duration("elapsed", elapsed),
		zap.Int64("rows", rows),
		zap.String("sql", sql),
	}

	// 记录错误（除非级别为 Info，否则排除 ErrRecordNotFound）或慢查询
	// 检查初始化时传递的 GORM 配置中的 IgnoreRecordNotFoundError（尽管目前在 InitDB 中是硬编码的）
	// 我们将假设此日志记录器逻辑中其为 true。更好的方法是将 gormLogCfg.IgnoreRecordNotFoundError 传递给 newZapGormLogger。
	isErr := err != nil && !(errors.Is(err, gorm.ErrRecordNotFound) /* && ignoreRecordNotFoundError */ && l.gormLogLevel < gormlogger.Info)
	isSlow := l.slowThreshold > 0 && elapsed > l.slowThreshold

	switch {
	case isErr && l.gormLogLevel >= gormlogger.Error:
		fields = append(fields, zap.Error(err))
		l.zapLogger.Error("[GORM] Trace", fields...)
	case isSlow && l.gormLogLevel >= gormlogger.Warn:
		fields = append(fields, zap.Duration("threshold", l.slowThreshold))
		l.zapLogger.Warn("[GORM] Slow Query", fields...)
	case l.gormLogLevel >= gormlogger.Info:
        if isErr { // 如果级别为 Info，则将 ErrRecordNotFound 记录为 Info
            fields = append(fields, zap.Error(err))
        }
		l.zapLogger.Info("[GORM] Trace", fields...)
	}
}