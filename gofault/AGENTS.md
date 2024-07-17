# AGENTS.md — gofault 协作规范

本项目使用 [gofault](https://github.com/gofault/gofault) 框架，多个 AI agent 协作开发时遵循以下规范。

---

## 代码规范

### Go 格式
- 所有代码必须通过 `go fmt` 格式化后再提交
- 运行 `go vet ./...` 确保无编译警告

### 中间件签名
```go
type MiddlewareFunc func(ctx *core.Ctx, next core.Handler) error
```
- 必须严格遵循此签名，不可改变
- 返回 `core.ErrAbort` 表示中断链而不调用 `next`

### 模块生命周期
- 所有实现 `core.Module` 的类型，必须在 `OnBoot` 中初始化资源，在 `OnShutdown` 中释放资源
- 使用 `RegisterOnBoot(fn)` / `RegisterOnShutdown(fn)` 注册钩子

### 错误处理
- 优先返回标准错误类型（`errors.New`, `fmt.Errorf` with `%w`）
- 业务异常使用 `exception.Throw*` 系列
- 中间件中的 panic 必须被 recover，避免进程崩溃

### 日志
- 使用 `log.Default()` 获取 `*slog.Logger`
- 结构化日志优先于 fmt.Print

---

## Commit 规范

### 格式
```
<type>: <简短描述>

- <详细变更（可选）>
```
