# Go Gin API 项目重构计划

本文档概述了将现有 `my-gin-skeleton` 项目重构为更模块化、可维护性更高的结构，并集成 `uber/fx` 进行依赖注入的计划。

## 1. 目标项目结构 (在 `my-gin-skeleton` 内部)

```
my-gin-skeleton/
├── cmd/
│   └── server/
│       └── main.go         # 应用启动入口，集成 fx
├── internal/
│   ├── core/               # 框架核心
│   │   ├── config/         # 配置加载 (viper) + 类型定义 (types.go)
│   │   ├── logger/         # 日志 (zap)
│   │   ├── di/             # 依赖注入 (fx) 核心注册 (可选，或分散在各模块)
│   │   ├── router/         # 路由 (gin)，自动扫描 biz handler
│   │   ├── middleware/     # 中间件 (Recovery, CORS, Auth, RateLimit etc.)
│   │   ├── plugin/         # 插件系统 (Auth, RateLimit 可作为插件实现)
│   │   ├── shutdown/       # 优雅关机
│   │   ├── pkg/            # 项目内部通用包
│   │   │   ├── tcrypt/       # 加密工具
│   │   │   └── tongchengapi/ # 重构后的同程 API 客户端
│   │   ├── db/             # 数据库连接 (GORM)
│   │   └── redis/          # Redis 连接 (go-redis)
│   └── biz/                # 业务模块
│       └── flight/         # flight 业务模块
│           ├── handler.go    # Gin Handler
│           ├── service.go    # 业务逻辑
│           ├── repo.go       # 数据访问层 (Repository Pattern) - 调用 tongchengapi
│           ├── dto.go        # Data Transfer Objects
│           └── flight_test.go # 单元/集成测试
├── scripts/
│   └── newbiz.sh           # 更新后的业务模块生成脚本
├── configs/
│   └── config.yml          # 配置文件 (需要添加 tongchengApi 配置)
├── go.mod
├── go.sum
└── README.md               # 更新后的文档
```

## 2. 重构步骤

1.  **准备阶段:**
    *   确保项目已通过 Git 进行版本控制，并创建新的重构分支。
    *   运行 `go get go.uber.org/fx` 添加依赖。

2.  **创建新目录结构:**
    *   按照上述目标结构创建必要的目录。

3.  **重构 `tongchengapi` 和 `tcrypt` (关键步骤):**
    *   将 `internal/pkg/tongchengapi` 和 `internal/pkg/tcrypt` 移动到 `internal/core/pkg/` 下。
    *   **目标:** 将 `tongchengapi` 改造为一个相对**无状态**的 API 客户端库。
    *   创建 `tongchengapi.Client` 结构体，封装 HTTP 客户端 (`req.Client`) 和必要的配置（通过 `fx` 注入）。
    *   将 `Get_airline_message`, `CreateOrder` 等 API 调用重构为 `Client` 的方法，接受明确参数，**移除 `Options` 依赖**。
    *   **移除内部调用链**（如 `Get_airline_message` 不再调用 `sift`）。
    *   **提取硬编码值**（URL, ID, 密钥, 盐值等）到配置中。
    *   加密密钥应通过配置管理。
    *   通过 `fx.Provide` 提供重构后的 `tongchengapi.Client` 实例。

4.  **迁移核心组件 (`internal/bootstrap` & `internal/plugin` -> `internal/core`):**
    *   **Config:**
        *   迁移配置结构体定义到 `core/config/types.go`。
        *   **新增 `TongchengAPIConfig` 结构体**，包含从 `tongchengapi` 提取的配置项。
        *   更新顶层 `Config` 结构体。
        *   迁移加载逻辑到 `core/config/config.go`，适配 `fx`，使其能读取新配置。
        *   **更新 `configs/config.yml`**，添加 `tongchengApi` 部分（敏感信息推荐使用环境变量）。
        *   提供 `fx.Provide`。
    *   **Logger, Database, Redis:**
        *   迁移初始化逻辑到 `core/logger`, `core/db`, `core/redis`。
        *   修改构造函数，接收 `fx.Lifecycle`, `*conf.AppConfig` (或具体子配置), `*zap.Logger` 作为参数。
        *   使用 `fx.Lifecycle` 的 `Append` 方法注册 `OnStop` 清理钩子（关闭连接）。
        *   提供 `fx.Provide`。
    *   **Middleware, Plugin, Router, Shutdown:**
        *   迁移到相应的 `core` 目录。
        *   适配 `fx` 进行依赖注入和生命周期管理。
        *   `Router` 需要改造为依赖 `fx` 注入 Gin Engine、中间件和业务 Handlers，并实现自动扫描注册。

5.  **迁移 `cmd`:**
    *   将启动逻辑移至 `cmd/server/main.go`。
    *   使用 `fx.New()` 和 `fx.Run()` 组装所有 `core` 和 `biz` 模块提供的 `fx.Provide`。
    *   使用 `fx.Invoke` 注册需要启动时执行的函数（如启动 Gin 服务器）。

6.  **重构 `flight` 业务模块 (`internal/biz/flight`):**
    *   创建 `handler.go`, `service.go`, `repo.go`, `dto.go`。
    *   **DTO:** 将 `internal/dto/flight_dto.go` 移动到 `internal/biz/flight/dto.go`。
    *   **Repo:** 创建 `repo.go`。定义 `FlightRepo` 接口。实现该接口的结构体依赖注入**重构后的 `tongchengapi.Client`**，并调用其方法。提供 `fx.Provide`。
    *   **Service:** 创建 `service.go`。注入 `FlightRepo`。**负责编排业务流程**（调用 Repo 方法），处理认证上下文，解析数据，映射 DTO。提供 `fx.Provide`。
    *   **Handler:** 创建 `handler.go`。注入 `FlightService`。提供 `fx.Provide`。
    *   **DI:** 可选，在 `internal/biz/flight/di.go` 中统一提供该模块的 `fx.Provide`。

7.  **迁移测试:**
    *   将测试文件移至相应的 `biz` 模块下（如 `internal/biz/flight/`）。
    *   更新测试代码以适应新结构和依赖注入（可能需要 Mock）。

8.  **更新脚本:**
    *   修改 `scripts/newbiz.sh` 以生成包含 `handler`, `service`, `repo`, `dto` 的新业务模块骨架。

9.  **清理和文档:**
    *   删除所有旧的、不再使用的文件和目录（`internal/bootstrap`, `internal/conf`, `internal/dto`, `internal/handler`, `internal/pkg`, `internal/plugin`, `internal/service`, `test`）。
    *   运行 `go mod tidy`。
    *   更新 `README.md`，说明新结构、运行方式、添加模块方法等。

## 3. Mermaid 图示 (简化依赖关系)

```mermaid
graph TD
    subgraph cmd/server
        A[main.go (fx.New)]
    end

    subgraph internal/core
        B(Config) --> A
        C(Logger) --> A & D & E & F & G & H & K & L & M  # Logger is widely used
        D(DB/Redis) --> A & M # DB/Redis provided to App and potentially Repo
        E(Router) --> A
        F(Middleware) --> E
        G(Plugins) --> A & E # Plugins might be middleware or invoked by App
        H(Shutdown) --> A
        I(core/pkg/tongchengapi.Client) --> A & M # API Client provided to App and Repo
        J(core/pkg/tcrypt) --> I # Crypt used by API Client
    end

    subgraph internal/biz/flight
        K(Handler) --> A & E # Handler registered with Router
        L(Service) --> K # Service injected into Handler
        M(Repo) --> L # Repo injected into Service
        N(DTO) --> K & L & M # DTOs used across layers
        O(Test)
    end

    A --> E & G & K & H  # fx Invokes: Start Router, Plugins, Handlers, Shutdowner

    L --> M & C # Service depends on Repo and Logger
    K --> L & C # Handler depends on Service and Logger
    M --> I & D & C # Repo depends on API Client, potentially DB/Redis, and Logger
    I --> J & B & C # API Client depends on Crypt, Config, and Logger

    subgraph scripts
        P(newbiz.sh)
    end

    subgraph configs
        Q(config.yml) --> B
    end

    subgraph docs
        R(README.md)
    end

    A -- Reads --> Q
```

## 4. 注意事项

*   **`tongchengapi` 重构是关键:** 这是本次重构最复杂的部分，需要仔细处理硬编码、耦合和状态问题。
*   **配置管理:** 敏感信息（API 密钥、数据库密码、加密密钥等）应使用环境变量管理，避免硬编码或直接写入配置文件。
*   **错误处理:** 确保在各层之间进行适当的错误包装和处理。
*   **测试:** 重构后需要更新或编写新的单元测试和集成测试，确保功能正确性。
*   **逐步进行:** 可以考虑分阶段进行重构，例如先重构 `core` 和 `tongchengapi`，再重构 `biz` 模块。