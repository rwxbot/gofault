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