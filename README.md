# 🌰 Chestnut -- Go 语言业务开发通用功能集合库

<div align="center">
  
![Version](https://img.shields.io/badge/version-v2.0.0-blue)
![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)
![License](https://img.shields.io/badge/License-MIT-green)
  
</div>

## 📖 项目简介

Chestnut 是一个功能丰富的 Go 语言业务开发通用功能集合库，旨在提供一套完整的工具和组件，帮助开发者快速构建高性能、可扩展的应用程序。该库集成了常用的数据库操作、缓存管理、日志系统、HTTP 服务、WebSocket 支持等功能，大幅提高开发效率。

## ✨ 核心特性

- **模块化设计**：各功能模块相互独立，可按需引入
- **丰富的组件**：支持多种数据库、缓存、消息队列等
- **高性能**：核心组件经过性能优化，支持高并发场景
- **易扩展**：提供统一的接口和抽象，方便扩展和自定义
- **完善的日志**：集成结构化日志系统，便于问题排查
- **配置灵活**：支持多种配置方式，适应不同环境需求

## 🔍 主要模块

### 🔄 核心模块 (core)

- **asynctask**：异步任务处理
- **cdefer**：延迟执行管理
- **file**：文件操作工具
- **gtask**：协程任务管理
- **log**：日志系统
- **signauth**：签名认证
- **utils**：通用工具函数

### 🛠️ 数据存储 (contrib/orm)

- **MySQL**：关系型数据库支持
- **SQLite**：轻量级数据库支持
- **ClickHouse**：列式数据库支持
- **MongoDB**：文档数据库支持
- **Redis**：内存数据库和缓存支持
- **Elasticsearch**：搜索引擎支持

### 🌐 网络服务 (modules)

- **httpserver**：HTTP 服务器
- **ginwebsocket**：WebSocket 支持
- **cronserver**：定时任务服务

### 🔐 其他功能

- **配置管理**：灵活的配置加载和环境变量支持
- **加密工具**：常用加密算法实现
- **并发控制**：锁机制和并发管理工具
- **验证器**：数据验证和多语言支持

## 🚀 快速开始

### 安装

```bash
go get github.com/bpcoder16/Chestnut/v2
```

### 基本使用

```go
package main

import (
	"context"
	"path"

    "github.com/bpcoder16/Chestnut/v2/appconfig"
    "github.com/bpcoder16/Chestnut/v2/bootstrap"
    "github.com/bpcoder16/Chestnut/v2/core/cdefer"
    "github.com/bpcoder16/Chestnut/v2/core/gtask"
    "github.com/bpcoder16/Chestnut/v2/modules/httpserver"
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
			//gin.HTTPHandler(
			//	route.Api(),
			//),
		).Run(ctx)
	})

    g.Wait()
}
```

## 📚 模块使用示例

### 日志系统

```go
import "github.com/bpcoder16/Chestnut/v2/logit"

func example() {
    // 记录信息日志
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
import (
    "github.com/bpcoder16/Chestnut/v2/contrib/orm/mysql"
)

func dbExample() {
    // 获取 MySQL 客户端
    db := mysql.MasterDB()
    
    // 执行查询
    var users []User
    result := db.Where("status = ?", "active").Find(&users)
    if result.Error != nil {
        // 处理错误
    }
}
```

### Redis 缓存

```go
import "github.com/bpcoder16/Chestnut/v2/contrib/goredis"

func redisExample() {
    // 获取 Redis 客户端
    client := redis.DefaultClient()
    
    // 设置缓存
    err := client.Set(ctx, "key", "value", time.Hour).Err()
    if err != nil {
        // 处理错误
    }
    
    // 获取缓存
    val, err := client.Get(ctx, "key").Result()
    if err != nil {
        // 处理错误
    }
}
```

## 📋 项目结构

```
├── appconfig/        # 应用配置管理
├── bootstrap/        # 应用启动和初始化
├── cmd/              # 命令行工具
├── contrib/          # 第三方集成组件
│   ├── aliyun/       # 阿里云服务集成
│   ├── esclientv7/   # Elasticsearch 客户端
│   ├── gomongodb/    # MongoDB 客户端
│   ├── goredis/      # Redis 客户端
│   ├── httphandler/  # HTTP 处理器
│   ├── log/          # 日志适配器
│   ├── lru/          # LRU 缓存
│   ├── orm/          # ORM 数据库支持
│   ├── validator/    # 数据验证器
│   └── websocket/    # WebSocket 支持
├── core/             # 核心功能模块
├── default/          # 默认实现
├── logit/            # 日志工具
└── modules/          # 功能模块
```

## 📄 许可证

本项目采用 MIT 许可证 - 详情请参阅 [LICENSE](LICENSE) 文件。
