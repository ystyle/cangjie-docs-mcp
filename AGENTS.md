# AGENTS.md — cangje-docs-mcp

## 项目概述

仓颉语言文档检索 MCP 服务器。单 binary，stdio 协议通信，启动时从 `https://gitcode.com/Cangjie/CangjieCorpus.git` clone 文档到本地。

## 构建与开发

```bash
go build -o cangje-docs-mcp main.go           # 编译
go build -ldflags "-s -w" -o cangje-docs-mcp main.go  # 生产编译（去除调试信息）
go test -v ./...                               # 测试
./cangje-docs-mcp -version                     # 版本+文档目录
./cangje-docs-mcp -dir /path/to/docs           # 指定文档目录
./cangje-docs-mcp -no-update                   # 离线模式（不 git pull）
```

注意：go.mod 里是 `go 1.25.4`，但 CI 用 `go 1.21` 构建——如有依赖兼容问题，降级版本号即可。

## 架构

| 包 | 职责 |
|---|---|
| `pkg/mcp` | MCP server 启动 + 4 个工具注册/处理 |
| `pkg/scanner` | 文档扫描、自动分割（>15KB 按 `##` 分割，不递归） |
| `pkg/search` | 3 级搜索（精确→索引→模糊），中英文关键词提取 |
| `pkg/types` | 类型定义、常量、搜索权重、学习路径 |
| `pkg/utils` | Git clone/fetch/reset（依赖宿主机 git） |

## MCP 工具（5 个）

- `cangjie_docs_overview` — 3 种视图，支持 `version` 参数
- `cangjie_list_docs` — 路径渐进，支持 `version` 参数
- `cangjie_search` — 空格分隔关键词 AND 匹配，支持 `version` 参数
- `cangjie_get_doc` — 支持 section 提取、format 控制，支持 `version` 参数
- `cangjie_list_versions` — 列出可用版本（semver 过滤），标注 latest/default

## 关键细节

- **文档自动分割**：超过 15KB 按 `##` 分割，子文档 `Prerequisites` 包含父文档 ID。遍历时注意过滤 `len(doc.Prerequisites) > 0` 可跳过分片。
- **日志**：`slog` 只输出 `LevelError` 到 stderr，防止干扰 MCP stdio 通信。
- **依赖**：运行时需要宿主机安装 `git` 命令，首次启动会完整克隆文档仓库（不再 `--depth 1`）。
- **文档缓存位置**：Linux `~/.config/cangje-docs-mcp/CangjieCorpus`，Windows `可执行文件同目录/CangjieCorpus`。
- **无测试文件**：`_test.go` 不存在，CI 中 `go test -v ./...` 仅验证编译通过。
- **CI**：`tag push v*` 触发，构建 4 平台 binary 并上传 release。
- **预编译 binary 被 gitignore**：根目录 `cangje-docs-mcp`（Linux x86_64 ELF）是编译产物，不应提交。
- **技能目录**：`skills/cangjie-docs-navigator/` — 外部 skill，不参与 Go 编译。

## 搜索权重（`pkg/types/constants.go`）

- 精确匹配 10.0 / 标题 8.0 / 描述 6.0 / 文件名 5.0 / 内容 3.0
- 默认 `max_results=10`, `min_confidence=0.3`
