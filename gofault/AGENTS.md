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

### Type 前缀
| 前缀 | 含义 |
|------|------|
| `feat:` | 新功能 |
| `fix:` | Bug 修复 |
| `test:` | 测试添加/修复 |
| `docs:` | 文档更新 |
| `refactor:` | 代码重构（不改变功能）|
| `chore:` | 构建、CI、依赖更新 |

### 示例
```bash
feat: v3.4.0 add API Versioning middleware

- Header/PathPrefix/Query three strategies
- VersionSet lifecycle management (Active/Deprecated/EOL)
- 17 tests pass
```

---

## 版本标签规范

### 标签格式
```
v<major>.<minor>.<patch>
```

### 打标签时机
1. 完成一个版本的所有功能开发
2. 所有测试通过（除明确标记为环境依赖的，如 Redis）
3. README 文档同步更新

### 打标签步骤
```bash
# 1. 确保 clean
git status

# 2. 打标签
git tag -a v3.4.0 -m "v3.4.0: API Versioning with Header/PathPrefix/Query strategies"

# 3. 验证
git tag -v v3.4.0
```

---

## 测试规范

### 必须覆盖
- 核心接口行为（Router, Module, Controller, Middleware）
- 边界条件（空输入、非法参数、路径遍历尝试）
- 错误路径（404, 401, 500 等）

### 环境依赖测试
- 需要外部服务（Redis, MySQL 等）的测试使用 `t.Skip()` 跳过
- 跳过格式：`t.Skip("requires Redis: " + err.Error())`
