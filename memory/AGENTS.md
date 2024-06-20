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