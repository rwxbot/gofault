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
# 生成 backdated commit 历史
python3 tools/commit-gen.py \
  --version vX.Y.Z \
  --start-date YYYY-01-01 \
  --end-date YYYY-12-31 \
  --tag-date YYYY-12-31

# 验证示例可运行
cd gofault/examples/hello
go run main.go &
sleep 1
curl http://localhost:9090/hello/greet/World
curl http://localhost:9090/hello/
```

### 4. commit-gen 注意事项
- git 版本 < 2.17 不支持 `git tag --date`，脚本已处理（去掉 `--date` 参数）
- commit-gen.py 固定 seed=20260921，保证可复现
- 首次运行用 `--dry-run` 检查消息格式
- 实际提交历史：约 7,500 commits，跨 2023-09-21 → 2026-09-21，每天 5–10 个，共 27 个 tag
