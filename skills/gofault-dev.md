# gofault-dev skill

## 触发条件
任何涉及 gofault 框架开发、接手的 Agent 必须加载本 skill。

## 开发流程

### 1. 读取交接文档
第一件事：读取 `memory/AGENTS.md`，了解当前进度、红线约束、目录结构。

### 2. 本地验证
```bash
cd /Users/Wang/Code/webfault-lib/gofault-all
go test ./...                          # 运行所有测试
go build ./...                         # 确认可构建
python3 tools/commit-gen.py --help    # 确认 commit-gen 可用
```

### 3. 开发新版本
```bash