# My Gin Skeleton

## 1. 项目简介

本项目是一个基于 [Gin](https://github.com/gin-gonic/gin) 框架构建的 Web 服务骨架。旨在提供一个快速开发、结构清晰、易于扩展的 Go Web 应用基础模板。

**主要技术栈:**

*   Web 框架: Gin
*   配置管理: [Viper](https://github.com/spf13/viper)
*   日志库: [Zap](https://github.com/uber-go/zap)
*   数据库 ORM: [Gorm](https://gorm.io/)
*   缓存: Redis ([go-redis/redis](https://github.com/go-redis/redis))
*   插件系统: 自定义接口实现

## 2. 项目结构

```
.
├── api             # API 定义文件 (例如 Swagger)
├── cmd             # 程序入口
│   └── server      # Web 服务启动入口 (main.go)
├── configs         # 配置文件目录
│   └── config.yml  # 主配置文件
├── internal        # 内部业务逻辑，不对外暴露
│   ├── bootstrap   # 初始化程序 (配置, 日志, DB, Redis, 路由, 中间件, 插件)
│   ├── conf        # 配置映射结构体
│   ├── dto         # 数据传输对象 (Data Transfer Objects)
│   ├── handler     # HTTP 请求处理器 (Controller)
│   ├── middleware  # 中间件
│   ├── pkg         # 项目内部公共包 (错误处理, 加密, 第三方 API 封装)
│   │   ├── errs
│   │   ├── tcrypt
│   │   └── tongchengapi
│   ├── plugin      # 插件实现 (鉴权, 限流等)
│   ├── repo        # 数据仓库层 (数据库操作)
│   └── service     # 业务逻辑层
├── logs            # 日志文件目录
├── scripts         # 脚本工具
│   ├── newbiz.sh   # 快速创建业务模块脚本
│   └── templates   # newbiz.sh 使用的模板
└── test            # 测试文件
    └── flight_api_test.go # flight 模块 API 集成测试示例
```

## 3. 快速开始

### 3.1 环境要求

*   Go: 1.18 或更高版本 (请根据 `go.mod` 文件确认)
*   数据库: MySQL (或其他 Gorm 支持的数据库)
*   缓存: Redis

### 3.2 配置说明

项目配置通过 `configs/config.yml` 文件管理。请根据实际环境修改以下配置项：

```yaml
app:
  env: development # 环境 (development, production, test)
  name: my-gin-skeleton
  version: v1.0.0

http:
  port: 8080       # HTTP 服务端口
  read_timeout: 60
  write_timeout: 60

log:
  level: debug     # 日志级别 (debug, info, warn, error)
  path: ./logs     # 日志文件路径
  max_size: 100    # 日志文件最大大小 (MB)
  max_backups: 5   # 最大备份数量
  max_age: 7       # 最大保留天数
  compress: false  # 是否压缩

db:
  mysql:
    dsn: "user:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local" # MySQL DSN
    max_idle_conns: 10
    max_open_conns: 100
    conn_max_lifetime: 3600

redis:
  addr: "127.0.0.1:6379" # Redis 地址
  password: ""          # Redis 密码
  db: 0                 # Redis DB 编号

jwt:
  secret: "your-secret-key" # JWT 密钥
  expire: 7200            # JWT 过期时间 (秒)
  issuer: "my-gin-skeleton" # JWT 发行者

ratelimit:
  enable: true            # 是否启用限流插件
  rate: 100               # 每秒允许的请求数
  burst: 200              # 令牌桶容量

tongchengapi:             # 同程 API 配置 (示例)
  base_url: "https://example.tongcheng.api"
  app_key: "your_app_key"
  app_secret: "your_app_secret"
  sign_secret: "your_sign_secret" # 可能用于签名的密钥
```

**注意:** 生产环境中请务必修改 `jwt.secret` 和 `tongchengapi` 相关密钥。

### 3.3 编译与运行

```bash
# 下载依赖
go mod tidy

# 编译
go build -o my-gin-skeleton ./cmd/server/main.go

# 运行
./my-gin-skeleton
```

或者直接运行：

```bash
go run ./cmd/server/main.go
```

服务将在 `http://localhost:8080` (或配置文件中指定的端口) 启动。

## 4. 核心特性

### 4.1 配置加载 (Viper)

*   使用 Viper 加载 `configs/config.yml` 文件。
*   配置信息映射到 `internal/conf/config.go` 中的结构体。
*   通过 `internal/bootstrap/config.go` 初始化并全局访问。

### 4.2 日志系统 (Zap)

*   使用 Zap 提供高性能的结构化日志。
*   支持日志级别、文件轮转、输出到控制台和文件。
*   通过 `internal/bootstrap/logger.go` 初始化。

### 4.3 数据库 (Gorm) 和 Redis

*   **Gorm:** 在 `internal/bootstrap/database.go` 中初始化数据库连接池。
*   **Redis:** 在 `internal/bootstrap/redis.go` 中初始化 Redis 客户端。

### 4.4 中间件

*   在 `internal/bootstrap/middleware.go` 中注册全局中间件，例如日志记录、错误恢复、跨域处理等。
*   可以在 `internal/bootstrap/router.go` 中为特定路由组或路由注册局部中间件。

### 4.5 插件系统

*   定义了统一的插件接口 `internal/plugin/plugin.go` (`Plugin` interface)。
*   通过 `internal/bootstrap/plugin.go` 根据配置加载并初始化启用的插件。
*   插件可以在启动时执行初始化操作，并可注册为 Gin 中间件。
*   **现有插件:**
    *   `ratelimit` (`internal/plugin/ratelimit.go`): 基于令牌桶算法的速率限制。
    *   `auth` (`internal/plugin/auth.go`): 基于 JWT 的用户认证。

## 5. API 端点

API 路由在 `internal/bootstrap/router.go` 中定义。以下是 `flight` 模块的主要示例端点：

*   **机票搜索**
    *   `GET /v1/flight/search`
    *   **功能:** 根据查询参数搜索机票信息。
    *   **请求 Query:** `departureCity`, `arrivalCity`, `date` 等 (参考 `internal/dto/flight.go` 中的 `FlightSearchReq`)。
    *   **响应 Body:** 机票列表 (参考 `internal/dto/flight.go` 中的 `FlightSearchResp`)。
*   **机票下单**
    *   `POST /v1/flight/order`
    *   **功能:** 创建机票订单。
    *   **请求 Body:** 订单信息 (参考 `internal/dto/flight.go` 中的 `FlightOrderReq`)。
    *   **响应 Body:** 订单结果 (参考 `internal/dto/flight.go` 中的 `FlightOrderResp`)。
    *   **认证:** 需要有效的 JWT Token (通过 `auth` 插件)。

**注意:** 上述仅为示例，具体参数和响应格式请参考 `internal/handler/flight.go` 和 `internal/dto/flight.go` 中的定义。

## 6. 添加新业务模块

项目提供了 `scripts/newbiz.sh` 脚本，可以快速生成新业务模块所需的 `handler`, `service`, `repo` 文件骨架。

**使用方法:**

```bash
cd scripts
bash newbiz.sh YourModuleName
```

例如，要创建一个名为 `hotel` 的新模块：

```bash
cd scripts
bash newbiz.sh hotel
```

脚本将在 `internal/handler`, `internal/service`, `internal/repo` 目录下分别创建 `hotel.go` 文件，并填充基础代码模板。之后，你需要在 `internal/bootstrap/router.go` 中注册新的路由，并在 `internal/bootstrap/database.go` 或其他初始化文件中注入依赖。

## 7. 测试

项目包含 API 集成测试示例。

**运行测试:**

在项目根目录 (`gogin/my-gin-skeleton`) 下执行：

```bash
go test ./test/... -v
```

这将运行 `test` 目录下的所有测试用例 (例如 `flight_api_test.go`)。`-v` 参数会显示详细的测试过程和结果。

确保在运行测试前，相关的依赖服务（如数据库、Redis）已正确配置并运行。测试可能需要独立的测试数据库或配置。