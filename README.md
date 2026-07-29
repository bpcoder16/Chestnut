# 🌰 Chestnut -- Go 语言业务开发通用功能集合库

<div align="center">

![Version](https://img.shields.io/badge/version-v4.0.0-blue)
![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)
![License](https://img.shields.io/badge/License-MIT-green)

</div>

## 项目简介

Chestnut 是一个功能丰富的 Go 语言业务开发通用功能集合库，旨在提供一套完整的工具和组件，帮助开发者快速构建高性能、可扩展的应用程序。该库集成了常用的数据库操作、缓存管理、日志系统、HTTP 服务、gRPC 服务、WebSocket 支持等功能，大幅提高开发效率。

## 核心特性

- **模块化设计**：各功能模块相互独立，可按需引入
- **丰富的组件**：支持多种数据库、缓存、消息队列等
- **高性能**：核心组件经过性能优化，支持高并发场景
- **易扩展**：提供统一的接口和抽象，方便扩展和自定义
- **完善的日志**：集成结构化日志系统，便于问题排查
- **配置灵活**：支持多种配置方式，适应不同环境需求

## 主要模块

### 核心模块 (core)

- **asynctask**：异步任务处理
- **cdefer**：延迟执行管理
- **file**：文件操作工具（日志轮转、标准文件操作）
- **gtask**：协程任务管理（基于 `errgroup`）
- **log**：结构化日志系统
- **signauth**：签名认证
- **utils**：通用工具函数（环境、网络、字符串、时间、随机数、脱敏等）

### 数据存储 (contrib/orm)

- **MySQL**：关系型数据库支持
- **SQLite**：轻量级数据库支持
- **ClickHouse**：列式数据库支持
- **MongoDB**：文档数据库支持（contrib/gomongodb）
- **Redis**：内存数据库和缓存支持（contrib/goredis）
- **Elasticsearch**：搜索引擎支持（contrib/esclientv7）
- **LRU**：本地 LRU 缓存（contrib/lru）

### 网络服务 (modules)

- **httpserver**：HTTP 服务器（基于 Gin）
- **grpcserver**：gRPC 服务器（支持 Keepalive、性能参数配置、优雅关闭）
- **cronserver**：定时任务服务（基于 gocron v2，支持 CronJob / DurationJob / DurationRandomJob）
- **ginwebsocket**：WebSocket 支持（基于 Gin）

### HTTP 中间件 (contrib/httphandler/gin)

- **DefaultLogger**：请求日志中间件（含请求/响应体记录、敏感字段脱敏）
- **RecoveryWithWriter**：Panic 恢复中间件
- **CORS / CORSPreCheckRequest**：跨域资源共享中间件
- **SignAuthMiddleware**：签名认证中间件
- **PProfRegister**：pprof 性能分析路由注册
- **SwaggerRegister**：Swagger UI 路由注册
- 内置 `/health` 健康检查端点

### 并发与锁

- **concurrency**：并发任务管理器，支持多协程并行执行与结果收集
- **lock/localpool**：本地内存锁池
- **lock/nonblock**：非阻塞式 Redis 分布式锁（支持批量加锁/解锁）

### 其他功能

- **配置管理**：灵活的配置加载和环境变量支持（appconfig）
- **加密工具**：AES-GCM 加密（modules/crypto）
- **验证器**：数据验证和多语言支持（contrib/validator）
- **数据同步**：DB 到 ES 数据同步（modules/dbtoes）
- **图片存储**：阿里云 OSS 图片上传（modules/image/aliyunoss）
- **HTTP 客户端**：基于 Resty 的 HTTP 客户端封装（default/resty）

## 快速开始

### 安装

```bash
go get github.com/bpcoder16/Chestnut/v4
```

### 基本使用

```go
package main

import (
	"context"
	"path"

	"github.com/bpcoder16/Chestnut/v4/appconfig"
	"github.com/bpcoder16/Chestnut/v4/appconfig/env"
	"github.com/bpcoder16/Chestnut/v4/bootstrap"
	"github.com/bpcoder16/Chestnut/v4/core/cdefer"
	"github.com/bpcoder16/Chestnut/v4/core/gtask"
	ginhandler "github.com/bpcoder16/Chestnut/v4/contrib/httphandler/gin"
	"github.com/bpcoder16/Chestnut/v4/modules/httpserver"
	"github.com/gin-gonic/gin"
)

func main() {
	config := appconfig.MustLoadAppConfig("/conf/app-server.yaml")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bootstrap.MustInit(ctx, config)
	defer cdefer.Defer()

	var g *gtask.Group
	g, ctx = gtask.WithContext(ctx)

	bootstrap.Start(ctx, config, g.Go)

	g.Go(func() error {
		return httpserver.NewManager(
			path.Join(env.ConfigDirPath(), "http.yaml"),
			ginhandler.HTTPHandlerWithConfig(
				config,
				nil,
				// 在此注册路由
				func(rg *gin.RouterGroup) {
					// rg.POST("/api/...", handler)
				},
				// 可选：启用 Swagger UI
				// ginhandler.SwaggerRegister,
				// 可选：启用 pprof 性能分析
				// ginhandler.PProfRegister,
			),
		).Run(ctx)
	})

	g.Wait()
}
```

### Prometheus API 监控

使用 `HTTPHandlerWithConfig` 时，可通过应用主配置启用 Engine 级 Prometheus 指标。`HTTPHandler` 和 `HTTPHandlerWithMiddlewares` 保持兼容且默认不注册指标；`prometheus.enabled=false` 时，配置感知入口同样不会创建 collectors 或 `/metrics`。

```yaml
prometheus:
  enabled: true
  serviceName: "replace-with-service-name"
  # 仅支持精确 URL Path；不支持正则、glob 或 query 条件。
  excludedPaths:
    - "/metrics"
    - "/health"
    - "/ready"
  # 根据最终 HTTP 状态排除已完成请求；404 不会污染 QPS 和延迟。
  excludedStatusCodes:
    - 404
```

启用时，`serviceName` 必须非空；非法 path 或 `[100, 599]` 之外的 HTTP 状态会在 `AppConfig.Check()` 阶段阻止启动。

指标契约：

| 指标 | 类型 | 标签 | 用途 |
|---|---|---|---|
| `chestnut_http_server_requests_total` | Counter | `service`, `method`, `route`, `status_class` | 已完成且未排除的 HTTP 请求数，按 `2xx` 等状态类别聚合 |
| `chestnut_http_server_request_duration_seconds` | Histogram | `service`, `method`, `route`, `status_class` | 请求耗时，按 `2xx` 等状态类别聚合并使用 `prometheus.DefBuckets` |
| `chestnut_http_server_requests_in_flight` | Gauge | `service` | 当前正在处理的匹配业务请求数 |
| `chestnut_http_server_recovered_panics_total` | Counter | `service`, `method`, `route` | 被 Chestnut Recovery 实际捕获的 panic 数 |

每个 Gin Engine 使用独立 registry，创建多个 Engine 不会重复注册或共享请求计数。`/metrics` 同时合并 Prometheus 默认 gatherer，因此仍可采集 `go_*`、`process_*` 和应用已注册到默认 registry 的自定义指标。

采集只使用 Gin `FullPath()` 路由模板、HTTP method 和最终 HTTP status，不读取请求体、响应体或业务 JSON `Response.Code`。请求 Counter 和 Histogram 都使用 `status_class`（如 `2xx`、`5xx`），不保留精确状态码标签，以控制时序基数。未匹配路由不会产生 `unknown` 或原始 URI 时序；`/metrics` 由框架强制排除，配置中的探活路径在开始计时和增加 in-flight 前排除，配置的 HTTP 404 在请求完成后排除。最终返回排除状态的请求在处理期间可能短暂计入 in-flight，但结束后会归零。

启用后的中间件顺序是 `Prometheus -> Recovery -> 调用方中间件 -> 路由`。响应尚未提交时发生 panic，Recovery 返回 HTTP 500；响应已经提交后发生 panic，Recovery 保持客户端实际收到的状态、Header 和 body，不缓存或重放响应。两种情况都会增加 Recovery Counter，请求 Counter 和 Histogram 使用客户端实际状态；即使实际状态位于 `excludedStatusCodes`，请求 Counter/Histogram 仍按规则排除，但 Recovery Counter 会保留真实 panic 信号。业务主动返回 HTTP 500 不会增加 Recovery Counter。

集群抓取与 Dashboard 示例：

- Prometheus：`contrib/httphandler/gin/conf.example/prometheus/scrape.d/chestnut-api.yaml`
- Grafana：
  - `contrib/httphandler/gin/conf.example/grafana/dashboards/chestnut-api/chestnut-api-overview-dashboard.json`
  - `contrib/httphandler/gin/conf.example/grafana/dashboards/chestnut-api/chestnut-api-route-analysis-dashboard.json`
  - `contrib/httphandler/gin/conf.example/grafana/dashboards/chestnut-api/chestnut-api-instance-runtime-dashboard.json`

把 scrape 文件复制到 Qilin 监控 package 的 `config/prometheus/scrape.d/`，逐台替换文档 target、`environment`、`cluster` 和 `node`；把三个 Dashboard JSON 复制到 `config/grafana/dashboards/chestnut-api/`。Dashboard 使用已 provision 的 Prometheus datasource UID `prometheus`，稳定 UID 分别为 `chestnut-api-overview`、`chestnut-api-route-analysis` 和 `chestnut-api-instance-runtime`。概览页每 30 秒刷新，路由分析及实例与运行时页面每分钟刷新；三个页面通过 `chestnut` 标签下拉导航，并保留当前时间范围和同名变量。文件为只读交付，不需要在 Grafana UI 手工创建面板。

三个 Dashboard 只保留 `node` 筛选变量，值直接从 Prometheus `up{job="chestnut-api"}` 读取；`job` 固定为 `chestnut-api`，`service`、`instance` 和 `route` 不作为筛选项，路由仅在分析面板中按指标标签聚合。仍在 scrape 配置中的失败 target 会以 `up=0` 保留在变量和“节点采集明细”中。健康区只使用 target labels：全部健康、单节点 down、全部 down 分别显示对应可用数、不可用数、可用率和正常/异常明细，只有当前筛选范围不存在任何 `up` 序列时才显示“无节点采集数据”。不要在 target labels 重复添加 `service`，否则 Prometheus 默认 `honor_labels=false` 会产生 `exported_service` 冲突。运行时 `go_*`、`process_*` 指标没有应用 `service` 标签，Dashboard 对这些指标使用 target 标签筛选。

`/metrics` 会暴露路由模板和进程状态，生产环境必须通过网络、反向代理、防火墙或既有白名单限制为 Prometheus/管理网络访问，示例不提供应用层鉴权或凭据。

从旧应用指标迁移时，`de_card_*` 不会双写为 `chestnut_*`。滚动发布期间，新 Dashboard 的请求面板只覆盖已升级节点；先用基于 `up` 的 target 健康明细确认所有目标可见且可用，再下线旧查询。应用回滚会恢复旧指标增长；若只把 `prometheus.enabled` 改为 `false`，还需同步停用对应 scrape job，避免 target 被报告为 down。Prometheus 历史时序不需要删除或回填。

### 定时任务

```go
package main

import (
	"context"

	"github.com/bpcoder16/Chestnut/v4/modules/cronserver"
)

// 实现 cronserver.Interface 接口
type DemoCron struct{}

func (c *DemoCron) Init(template cronserver.Interface) {}
func (c *DemoCron) Before(name string, maxConcurrencyCnt int) {}
func (c *DemoCron) GetIsRun(ctx context.Context) bool { return true }
func (c *DemoCron) Process(ctx context.Context) {
	// 任务逻辑
}
func (c *DemoCron) Run(ctx context.Context) {}
func (c *DemoCron) Defer(ctx context.Context) {}

func init() {
	cronserver.RegisterCron("demo_cron", &DemoCron{})
}
```

### gRPC 服务

```go
package main

import (
	"github.com/bpcoder16/Chestnut/v4/modules/grpcserver"
	"google.golang.org/grpc"
)

// 实现 grpcserver.Service 接口
type DemoService struct{}

func (s *DemoService) RegisterService(sr grpc.ServiceRegistrar) {
	// pb.RegisterYourServiceServer(sr, s)
}

func main() {
	manager := grpcserver.NewManager(
		"path/to/grpc.yaml",
		&DemoService{},
	)
	_ = manager.Run(ctx)
}
```

### 并发任务管理

```go
import "github.com/bpcoder16/Chestnut/v4/modules/concurrency"

func example(ctx context.Context) {
	taskMap := map[string]func(ctx context.Context) concurrency.ChanResult{
		"user":   func(ctx context.Context) concurrency.ChanResult { /* ... */ },
		"order":  func(ctx context.Context) concurrency.ChanResult { /* ... */ },
	}
	resultMap, err := concurrency.Manager(ctx, taskMap, "FetchUserOrder")
	_ = resultMap["user"].Result
	_ = resultMap["order"].Result
}
```

### Redis 分布式锁

```go
import "github.com/bpcoder16/Chestnut/v4/modules/lock/nonblock"

func example(ctx context.Context, redisClient *redis.Client) {
	// 单个锁
	success := nonblock.RedisLock(ctx, redisClient, "lock:key", time.Minute)
	if success {
		defer nonblock.RedisUnlock(ctx, redisClient, "lock:key")
		// 执行业务逻辑
	}

	// 批量锁（任一获取失败自动释放已获取的锁）
	unlock, allLocked := nonblock.BizBatchLock(ctx, redisClient, "lock:%s", "res1", "res2")
	if allLocked {
		defer unlock()
		// 执行业务逻辑
	}
}
```

### 日志系统

```go
import "github.com/bpcoder16/Chestnut/v4/logit"

func example() {
	// 记录调试日志
	logit.DebugW("Example", "Run")

	// 记录错误日志
	err := someFunction()
	if err != nil {
		logit.ErrorW("操作失败", err)
	}
}
```

### 数据库操作

```go
import "github.com/bpcoder16/Chestnut/v4/contrib/orm/mysql"

func dbExample() {
	db := mysql.MasterDB()

	var users []User
	result := db.Where("status = ?", "active").Find(&users)
	if result.Error != nil {
		// 处理错误
	}
}
```

### Redis 缓存

```go
import "github.com/bpcoder16/Chestnut/v4/contrib/goredis"

func redisExample() {
	client := redis.DefaultClient()

	err := client.Set(ctx, "key", "value", time.Hour).Err()
	val, err := client.Get(ctx, "key").Result()
}
```

## 项目结构

```
├── appconfig/        # 应用配置管理
│   └── env/          # 环境变量与运行模式
├── bootstrap/        # 应用启动和初始化
├── cmd/              # 命令行工具
├── contrib/          # 第三方集成组件
│   ├── aliyun/       # 阿里云服务集成（OSS）
│   ├── cron/         # 定时任务调度器封装（gocron v2）
│   ├── esclientv7/   # Elasticsearch 客户端
│   ├── gomongodb/    # MongoDB 客户端
│   ├── goredis/      # Redis 客户端
│   ├── httphandler/  # HTTP 处理器
│   │   └── gin/      # Gin 框架集成（中间件、Swagger、PProf）
│   ├── log/          # 日志适配器
│   │   └── zap/      # Zap 日志适配
│   ├── lru/          # LRU 缓存
│   ├── orm/          # ORM 数据库支持（MySQL / SQLite / ClickHouse）
│   ├── validator/    # 数据验证器（多语言）
│   └── websocket/    # WebSocket 支持
├── core/             # 核心功能模块
│   ├── asynctask/    # 异步任务
│   ├── cdefer/       # 延迟执行
│   ├── file/         # 文件操作
│   ├── gtask/        # 协程任务管理
│   ├── log/          # 结构化日志
│   ├── signauth/     # 签名认证
│   └── utils/        # 通用工具函数
├── default/          # 默认实现（数据库/缓存/锁/HTTP客户端初始化）
├── logit/            # 全局日志工具
└── modules/          # 功能模块
    ├── concurrency/  # 并发任务管理
    ├── cronserver/   # 定时任务服务
    ├── crypto/       # 加密工具（AES-GCM）
    ├── dbtoes/       # DB 到 ES 数据同步
    ├── ginwebsocket/ # WebSocket（Gin 集成）
    ├── gormcommon/   # GORM 通用工具
    ├── grpcserver/   # gRPC 服务器
    ├── httpserver/   # HTTP 服务器
    ├── image/        # 图片处理（阿里云 OSS）
    ├── lock/         # 锁机制（本地锁池、Redis 分布式锁）
    └── zaplogger/    # Zap 日志初始化
```

## 📄 许可证

本项目采用 MIT 许可证 - 详情请参阅 [LICENSE](LICENSE) 文件。
