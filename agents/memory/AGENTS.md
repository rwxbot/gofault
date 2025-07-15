# gofault 框架交接文档

## 项目概述

**gofault** 是一款基于模块化架构与依赖注入的纯 Golang 后端框架，module path: `github.com/gofault/gofault`。
顶层目录 `gofault-all`（git 仓库根），框架代码在 `gofault/` 子目录，沉淀物同级隔离。

## 硬性约束：零 Node 生态痕迹

框架代码（含标识符、注释、示例、文档、错误信息）**绝对禁止**出现：

- `@` 装饰器语法
- Factory / Application 式框架入口
- `require(` / `module.exports`
- `main.ts`
- 非 Go 生态的模块与文件后缀

只使用纯 Go 惯用法：interface、组合、struct tag、functional options。

## 当前状态

| 版本 | 状态 | 说明 |
|------|------|------|
| v0.1.0 | ✅ 完成 | 最小内核：IoC + Module/Controller + HTTP 路由 |
| v0.5.0 | ✅ 完成 | 构造器注入完善 + Middleware 链 + Exception Filter + 生命周期钩子 |
| v1.0.0–v2.8.0 | ✅ 完成 | 中间件集合 |
| v2.9.0 | ✅ 完成 | File Upload |
| v3.0.0 | ✅ 完成 | GORM |
| v3.1.0 | ✅ 完成 | Redis |
| v3.2.0 | ✅ 完成 | Static Files |
| v3.3.0 | ✅ 完成 | gRPC |
| v3.4.0 | ✅ 完成 | API Versioning |

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

## 关键决策（已确认，勿重复询问）

- 顶层目录：`gofault-all`
- 框架目录：`gofault`（子目录，由顶层仓库统一追踪）
- 示例端口：9090
- commit 历史：按真实时间均匀分布的 backdated（2023-09-21 → 2026-09-21，每天 5–10 个 commit，约 7500 个）
- 框架命名：gofault

## 仓库

单一 Git 仓库 `gofault-all`，框架代码位于 `gofault/` 子目录（不再嵌套独立仓库）。远程：`rwx-robot-home/gofault-home`。
