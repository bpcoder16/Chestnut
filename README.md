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
- **数据库初始化**：数据库自动迁移工具（modules/initdb）
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
	"github.com/bpcoder16/Chestnut/v4/contrib/httphandler/gin"
	"github.com/bpcoder16/Chestnut/v4/modules/httpserver"
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
			gin.HTTPHandler(
				// 在此注册路由
				func(rg *gin.RouterGroup) {
					// rg.POST("/api/...", handler)
				},
				// 可选：启用 Swagger UI
				// gin.SwaggerRegister,
				// 可选：启用 pprof 性能分析
				// gin.PProfRegister,
			),
		).Run(ctx)
	})

	g.Wait()
}
```

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
    ├── initdb/       # 数据库初始化
    ├── lock/         # 锁机制（本地锁池、Redis 分布式锁）
    └── zaplogger/    # Zap 日志初始化
```

## 📄 许可证

本项目采用 MIT 许可证 - 详情请参阅 [LICENSE](LICENSE) 文件。
