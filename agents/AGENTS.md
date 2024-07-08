# gofault-all Agent 工作区

## 目录结构

```
gofault-all/                    # Git 仓库根（管理 agents 配置、skills、工具）
├── .git/
├── .gitignore                  # 忽略临时/系统文件
├── gofault/                    # 框架代码（由顶层仓库统一追踪）
│   ├── core/                   # 核心接口
│   ├── router/                 # 路由引擎
│   ├── server/                 # HTTP Server
│   ├── controller/             # Controller 基类
│   ├── middleware/             # 中间件集合
│   ├── provider/               # Provider 工具
│   ├── ioc/                    # IoC 容器
│   ├── logger/                 # 日志封装
│   ├── exception/              # 异常类型
│   ├── config/                 # 配置加载
│   ├── module/                 # Module 辅助函数
│   ├── gorm/                   # GORM 集成
│   ├── redis/                  # Redis 集成
│   ├── grpc/                   # gRPC 集成
│   ├── versioning/             # API 版本控制
│   ├── docs/                  # 架构文档
│   ├── examples/               # 示例
│   ├── testapp3/               # 测试应用
│   ├── cmd/                    # CLI 命令
│   ├── go.mod
│   └── .github/workflows/     # CI/CD
├── agents/                     # Agent 配置和记忆
│   ├── memory/                 # 项目记忆
│   └── skills/                 # 技能定义
├── tools/                      # 工具脚本
│   └── commit-gen.py           # backdated commit 生成
└── reports/                    # 性能报告
```

## 仓库

单一 Git 仓库 `gofault-all`（远程: `rwx-robot-home/gofault-home`）。框架源码位于 `gofault/` 子目录，Agent 工作区配置位于 `agents/`、`skills/`、`tools/`。
