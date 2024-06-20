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
