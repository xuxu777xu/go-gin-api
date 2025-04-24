package conf

// Config 应用总配置
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Modules  ModulesConfig  `yaml:"modules"` // 使用结构体替代 map 来支持更复杂的模块配置
	Logger   LoggerConfig   `yaml:"logger"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Addr         string `yaml:"addr"`
	ReadTimeout  string `yaml:"readTimeout"`  // 例如: "15s"
	WriteTimeout string `yaml:"writeTimeout"` // 例如: "15s"
}

// ModulesConfig 模块配置集合
type ModulesConfig struct {
	RateLimit RateLimitConfig `yaml:"ratelimit"`
	Auth      AuthConfig      `yaml:"auth"` // 保留 Auth 配置结构以备将来使用
	// 在此添加其他模块的配置结构
}

// RateLimitConfig 限流插件配置
type RateLimitConfig struct {
	Enable bool    `yaml:"enable"` // 是否启用插件
	Rate   float64 `yaml:"rate"`   // 每秒允许的请求数 (令牌生成速率)
	Burst  int     `yaml:"burst"`  // 令牌桶的容量 (允许的瞬时突发量)
}

// AuthConfig 认证插件配置
type AuthConfig struct {
	Enable bool   `mapstructure:"enable"`
	Secret string `mapstructure:"secret"` // JWT 密钥
	Expire int64  `mapstructure:"expire"` // 过期时间 (秒)
	Issuer string `mapstructure:"issuer"` // 签发者 (可选)
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	Level      string `yaml:"level"`
	File       string `yaml:"file"`
	MaxSize    int    `yaml:"maxSize"`    // 兆字节
	MaxBackups int    `yaml:"maxBackups"` // 备份文件数
	MaxAge     int    `yaml:"maxAge"`     // 天数
	Compress   bool   `yaml:"compress"`
}

// DatabaseConfig 数据库配置
// 注意：yaml 不直接支持 time.Duration, gorm 通常在初始化时处理 dsn 中的参数或单独设置
type DatabaseConfig struct {
	Enable          bool   `yaml:"enable"`
	Driver          string `yaml:"driver"`
	DSN             string `yaml:"dsn"`
	MaxIdleConns    int    `yaml:"maxIdleConns"`
	MaxOpenConns    int    `yaml:"maxOpenConns"`
	ConnMaxLifetime string `yaml:"connMaxLifetime"` // 保持字符串形式，初始化时解析
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Enable   bool   `yaml:"enable"`
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}