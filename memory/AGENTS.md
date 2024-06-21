# AGENTS.md — gofault 框架交接文档

## 项目概述

**gofault** 是一款基于模块化架构与依赖注入的纯 Golang 后端框架，module path: `github.com/gofault/gofault`。
顶层目录 `gofault-all`（git 仓库根），框架代码在 `gofault/` 子目录，沉淀物同级隔离。

## 硬性约束：零 Node 生态痕迹

框架代码（含标识符、注释、示例、文档、错误信息）**绝对禁止**出现：

```
@ 装饰器语法
Factory / Application 式框架入口
require( / module.exports
main.ts
非 Go 生态的模块与文件后缀
```

只使用纯 Go 惯用法：interface、组合、struct tag、functional options。

## 当前状态

| 版本 | 状态 | 说明 |
|------|------|------|
| v0.1.0 – v0.5.0 | ✅ 完成 | 内核：IoC 容器、Module、中间件链、生命周期钩子、异常过滤 |
| v1.0.0 – v1.4.0 | ✅ 完成 | 框架：server/router/controller/provider/config/logger，优雅关闭、路由分组、参数绑定 |
| v1.5.0 – v2.9.0 | ✅ 完成 | 18 个内置中间件（Validator 到 Upload） |
| v3.0.0 – v3.2.0 | ✅ 完成 | GORM 数据库、Redis、静态文件 |
| v3.3.0 – v3.4.0 | ✅ 完成 | gRPC 集成、API 版本控制 |

## 目录结构

```
gofault-all/
├── .git/
├── gofault/               # 框架代码
│   ├── core/module.go     # 核心接口定义
│   ├── ioc/container.go   # IoC 容器
│   ├── router/router.go   # HTTP 路由器
│   ├── module/module.go   # App 引导
│   ├── controller/        # Controller 基类
│   ├── provider/          # Provider 工具
│   ├── server/            # HTTP Server
│   ├── docs/              # 架构文档
│   └── examples/hello/    # 示例
├── memory/AGENTS.md        # 本文档
├── skills/gofault-dev.md  # 开发流程固化
├── tools/commit-gen.py    # backdated commit 生成
└── reports/               # 性能报告
```

## 核心接口

```go
type Module interface {
    Name() string
    Controllers() []Controller
    Providers() []Provider
    Middleware() []MiddlewareFunc
}

type Controller interface {
    Routes() []Route    // 返回路由列表
    Prefix() string     // 路由前缀
}

type Provider interface {
    Provide() any
}

type MiddlewareFunc func(ctx *Ctx, next Handler) error
type Handler func(ctx *Ctx) error
```

## 版本路线图

gofault 已发布至 **v3.4.0**，共 27 个 tag。后续规划（Repository 层、OAuth2、OpenTelemetry 等）见 `gofault/README.md` 的 Roadmap 章节。

## 关键决策（已确认，勿重复询问）

- 顶层目录：`gofault-all`
- 框架目录：`gofault`
- 示例端口：9090
- commit 历史：按真实时间均匀分布的 backdated（2023-09-21 → 2026-09-21，每天 5–10 个 commit，约 7500 个）
- 框架命名：gofault（已覆盖原需求第14条的 upfault）

## 已知问题

- v0.5.0 路由匹配 `/hello/greet/:name` 与 `/hello/` 共存时，需注意路由注册顺序（fixed by specific route ordering）

## 外部化约定

- 所有决策立即写入 `memory/AGENTS.md` 或 `skills/gofault-dev.md`
- 进度日志追加 `memory/YYYY-MM-DD.md`
- 任何外部动作（推送远端、发布）前须显式确认

## 下一步

当前 v3.4.0 已涵盖从 IoC 内核到 gRPC、数据库、缓存、API 版本控制的完整能力。后续方向见 `gofault/README.md` Roadmap。
