package plugin

import (
	"errors"
	"fmt"
	"strings"
	"time" // 需要 time 包来处理过期时间

	"myGin/internal/conf"
	"myGin/internal/pkg/errs"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5" // 使用别名
	"go.uber.org/zap"
)

// AuthPlugin 实现了 Plugin 接口，用于提供 JWT 认证功能。
type AuthPlugin struct {
	authCfg *conf.AuthConfig // 存储加载的认证配置
	logger  *zap.Logger
	secret  []byte // 预编译的 secret，提高性能
}

// MyCustomClaims 定义了 JWT 的自定义 Claims，嵌入了 RegisteredClaims 并添加了 UserID。
// 你可以根据需要添加更多字段。
type MyCustomClaims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"` // 添加 Username 字段示例
	jwt.RegisteredClaims
}

// NewAuthPlugin 创建一个新的 AuthPlugin 实例。
func NewAuthPlugin() Plugin {
	// 确保返回类型是 Plugin 接口
	var _ Plugin = (*AuthPlugin)(nil) // 编译时检查 *AuthPlugin 是否实现了 Plugin 接口
	return &AuthPlugin{}
}

// Init 初始化 AuthPlugin。符合 Plugin 接口，接收 interface{} 并断言为 *conf.AuthConfig。
func (p *AuthPlugin) Init(cfg interface{}, deps map[string]interface{}) error {
	// 1. 类型断言获取 Auth 配置
	authCfg, ok := cfg.(*conf.AuthConfig)
	if !ok {
		// 注意：这里返回错误，因为 bootstrap/plugin.go 明确传递了 *conf.AuthConfig
		// 如果类型不匹配，说明调用逻辑有问题。
		return fmt.Errorf("auth plugin init failed: expected config type *conf.AuthConfig, but got %T", cfg)
	}
	p.authCfg = authCfg // 存储断言后的配置

	// 2. 获取 logger 依赖
	loggerDep, ok := deps["logger"]
	if !ok {
		// 如果没有 logger，可能需要返回错误或使用默认 logger
		return fmt.Errorf("auth plugin init failed: logger dependency is missing")
	}
	p.logger, ok = loggerDep.(*zap.Logger)
	if !ok {
		return fmt.Errorf("auth plugin init failed: logger dependency is not of type *zap.Logger")
	}

	// 3. 检查是否启用 (现在 p.authCfg 肯定不是 nil)
	if !p.authCfg.Enable {
		p.logger.Info("Auth Plugin is disabled by config (Enable=false or auth section missing).")
		return nil // 如果未启用，则无需执行任何操作
	}

	// 4. 检查并存储密钥 (仅在启用时检查)
	if p.authCfg.Secret == "" {
		p.logger.Error("Auth Plugin init failed: JWT secret is empty in config")
		return fmt.Errorf("auth plugin init failed: JWT secret cannot be empty when enabled")
	}
	p.secret = []byte(p.authCfg.Secret) // 存储字节切片

	// 5. 检查过期时间 (秒) (仅在启用时检查)
	if p.authCfg.Expire <= 0 {
		p.logger.Warn("Auth Plugin: JWT expire time (seconds) is not set or invalid, using default 1 hour (3600s)", zap.Int64("expire_seconds", p.authCfg.Expire))
		p.authCfg.Expire = 3600 // 默认1小时 (3600秒)
	}

	// 6. 检查签发者 (可选) (仅在启用时检查)
	if p.authCfg.Issuer == "" {
		p.logger.Warn("Auth Plugin: JWT issuer is not set in config.")
		// 根据策略决定是否返回错误，这里仅警告
	}

	p.logger.Info("Auth Plugin initialized successfully.",
		zap.Int64("expire_seconds", p.authCfg.Expire),
		zap.String("issuer", p.authCfg.Issuer),
	)
	return nil
}

// Register 将 JWT 认证中间件注册到 Gin 引擎。
func (p *AuthPlugin) Register(r *gin.Engine) error {
	// 再次检查是否启用，因为 Init 可能因为配置缺失而将 Enable 设为 false
	if p.authCfg == nil || !p.authCfg.Enable {
		p.logger.Info("Auth Plugin is disabled, skipping middleware registration.")
		return nil // 如果未启用，则不注册任何路由或中间件
	}

	p.logger.Info("Registering Auth Plugin middleware globally...")
	r.Use(p.jwtAuthMiddleware())

	p.logger.Info("Auth Plugin middleware registered globally.")
	return nil
}


// jwtAuthMiddleware 创建并返回 JWT 认证中间件。
func (p *AuthPlugin) jwtAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			p.logger.Debug("Auth middleware: Authorization header is missing")
			errs.Unauthorized.WrapWithMessage(nil, "请求未携带Token").JSON(c)
			c.Abort()
			return
		}

		// 检查是否是 Bearer Token
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			p.logger.Debug("Auth middleware: Authorization header format is invalid", zap.String("header", authHeader))
			errs.Unauthorized.WrapWithMessage(nil, "Token格式错误").JSON(c)
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 解析和验证 Token
		claims := &MyCustomClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// 确保签名方法是预期的 HMAC 方法
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			// 返回初始化时存储的密钥
			return p.secret, nil
		}, jwt.WithIssuer(p.authCfg.Issuer), // 添加 Issuer 验证
			jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name})) // 明确指定签名算法

		if err != nil {
			p.logger.Warn("Auth middleware: Token parsing error", zap.Error(err))
			// 根据具体错误类型返回不同消息
			if errors.Is(err, jwt.ErrTokenExpired) {
				errs.Unauthorized.WrapWithMessage(err, "Token已过期").JSON(c)
			} else if errors.Is(err, jwt.ErrSignatureInvalid) {
				errs.Unauthorized.WrapWithMessage(err, "Token签名无效").JSON(c)
			} else if errors.Is(err, jwt.ErrTokenNotValidYet) {
				errs.Unauthorized.WrapWithMessage(err, "Token尚未生效").JSON(c)
			} else if errors.Is(err, jwt.ErrTokenMalformed) {
				errs.Unauthorized.WrapWithMessage(err, "Token格式错误").JSON(c)
			} else if errors.Is(err, jwt.ErrTokenUsedBeforeIssued) {
				errs.Unauthorized.WrapWithMessage(err, "Token在签发前被使用").JSON(c)
			} else {
				// 其他解析错误，例如 Issuer 不匹配等
				errs.Unauthorized.WrapWithMessage(err, "Token无效").JSON(c)
			}
			c.Abort()
			return
		}

		// 双重检查 token.Valid，虽然 ParseWithClaims 内部会检查
		if !token.Valid {
			p.logger.Warn("Auth middleware: Token is invalid (post-parsing check)")
			errs.Unauthorized.WrapWithMessage(nil, "Token无效").JSON(c)
			c.Abort()
			return
		}

		// 验证通过，将解析出的 claims 存储到 Gin 的上下文中
		c.Set("claims", claims) // 存储整个 claims 对象
		p.logger.Debug("Auth middleware: Token validated successfully", zap.Int64("userID", claims.UserID), zap.String("username", claims.Username))

		c.Next()
	}
}

// GenerateToken 生成一个新的 JWT Token。
// userID 和 username 是示例，你可以根据需要传递更多信息或一个包含所有信息的结构体。
func (p *AuthPlugin) GenerateToken(userID int64, username string) (string, error) {
	if p.authCfg == nil || !p.authCfg.Enable {
		return "", errors.New("auth plugin is disabled or not configured")
	}
	if len(p.secret) == 0 {
		// 理论上 Init 阶段已检查，这里是双重保险
		return "", errors.New("jwt secret is not configured")
	}

	// 计算过期时间 (使用秒)
	expireDuration := time.Duration(p.authCfg.Expire) * time.Second
	expireTime := time.Now().Add(expireDuration)

	// 创建 Claims
	claims := MyCustomClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),           // 过期时间
			IssuedAt:  jwt.NewNumericDate(time.Now()),           // 签发时间
			NotBefore: jwt.NewNumericDate(time.Now()),           // 生效时间
			Issuer:    p.authCfg.Issuer,                         // 签发者
			Subject:   fmt.Sprintf("%d", userID),                // 主题，通常是用户ID的字符串形式
			// ID:        "unique_id", // JWT ID, 可选
			// Audience:  []string{"some_audience"}, // 受众, 可选
		},
	}

	// 创建 Token 对象
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名 Token
	signedToken, err := token.SignedString(p.secret)
	if err != nil {
		p.logger.Error("Failed to sign token", zap.Error(err))
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}