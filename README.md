# 仓颉语言文档检索 MCP 服务器

为 AI 编程助手提供仓颉语言文档检索能力的 [MCP](https://modelcontextprotocol.io/) 服务器，支持 **DeepSeek Harness、Claude Code、opencode、Codex** 等主流 AI 客户端。首次运行自动克隆官方文档仓库，之后每次启动自动增量更新。

## ✨ 核心特性

- 🚀 **开箱即用** - 首次运行自动下载文档，无需手动配置
- 🔄 **自动更新** - 每次启动自动更新到最新版本
- ⚡ **智能分割** - 大文档自动分割成小文档，优化 AI 处理效率
- 🎯 **精准搜索** - 支持全文搜索、分类搜索、章节定位
- 📚 **完整覆盖** - 支持手册、标准库（std/stdx）、OpenHarmony、工具文档
- 🔀 **多版本切换** - 支持 `1.0.0`、`1.1.0` 等多版本文档同时查询，Agent 按需切换
- 💰 **Token 高效** - 树形文本和表格格式，节省 70%+ Token 消耗

## 🚀 快速开始

### 第一步：下载可执行文件

从 [Releases](https://github.com/ystyle/cangje-docs-mcp/releases) 下载对应平台的可执行文件，或自行编译：

```bash
go build -o cangje-docs-mcp main.go
```

### 第二步：配置 MCP 服务器

选择你的 AI 客户端，按对应章节配置：

| 客户端 | 配置文件 | 配置教程 |
|--------|---------|---------|
| **DeepSeek Harness** | `~/.dsh/profiles/web/cordis.patch.yml` | [配置](#deepseek-harness) |
| **Claude Code** | `.mcp.json` / `claude mcp add` | [配置](#claude-code) |
| **opencode** | `opencode.json` | [配置](#opencode) |
| **Codex** | `~/.codex/config.toml` | [配置](#codex) |

### 第三步：安装 Skill（推荐）

安装 Skill 后，AI 可以更智能地检索仓颉文档，支持多种搜索模式：

```bash
# GitHub（推荐）
npx skills add ystyle/cangjie-docs-mcp

# 国内用户（AtomGit 镜像）
npx skills add https://atomgit.com/Cangjie-SIG/cangjie-docs-mcp.git
```

Skill 支持 4 种搜索模式：
- ⚡ **直接搜索**：精确 API 查询（如 "String.split"）
- 🧠 **PageIndex 智能检索**：模糊查询（如 "怎么截取字符串"）
- ⚖️ **混合模式**：自动切换最优策略
- 🧭 **探索模式**：学习引导

### 第四步：重启客户端

重启后即可使用！系统会自动：
- ✅ 下载仓颉文档到默认位置
- ✅ 每次启动时更新文档
- ✅ 自动分割大文档优化处理

## 🔌 MCP 客户端配置

> 以下配置均假设可执行文件已加入 `PATH`；否则请将 `cangje-docs-mcp` 替换为完整路径。

### DeepSeek Harness

在 `~/.dsh/profiles/web/cordis.patch.yml` 中注册 MCP 客户端插件（若文件已存在，将 `mcp-cangje-docs` 这一行合并到现有的 `insert` 列表中）：

```yaml
- insert:
    - id: mcp-cangje-docs
      name: "@deepseek-ai/dsh-mcp-client"
      config:
        serverName: cangje-docs
        transport: stdio
        command: cangje-docs-mcp
```

保存后重启 DSH（HMR 会自动重连）。工具将以 `mcp__cangje-docs__*` 命名空间注册，例如 `mcp__cangje-docs__cangjie_search`。

### Claude Code

**方式一：项目级配置**（`.mcp.json` 放在项目根目录，可随仓库共享）：

```json
{
  "mcpServers": {
    "cangjie-docs": {
      "command": "/path/to/cangje-docs-mcp",
      "args": []
    }
  }
}
```

**方式二：命令行添加**：

```bash
claude mcp add cangjie-docs -- /path/to/cangje-docs-mcp
```

### opencode

编辑项目根目录的 `opencode.json`，或全局 `~/.config/opencode/opencode.json`：

```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "cangjie-docs": {
      "type": "local",
      "command": ["/path/to/cangje-docs-mcp"],
      "enabled": true
    }
  }
}
```

也可以用命令行添加：

```bash
opencode mcp add cangjie-docs -- /path/to/cangje-docs-mcp
```

### Codex

编辑 `~/.codex/config.toml`：

```toml
[mcp_servers.cangjie-docs]
command = "/path/to/cangje-docs-mcp"
args = []
enabled = true
```

也可以用命令行添加：

```bash
codex mcp add cangjie-docs -- /path/to/cangje-docs-mcp
```

## 📂 文档存储位置

系统会根据操作系统自动选择文档存储位置：

**Windows:**
```
可执行文件同目录\CangjieCorpus
```

**Linux/macOS:**
```
~/.config/cangje-docs-mcp/CangjieCorpus
```

## ⚙️ 命令行参数

```bash
# 查看版本和文档目录
./cangje-docs-mcp -version

# 查看帮助信息
./cangje-docs-mcp -help

# 使用自定义文档目录
./cangje-docs-mcp -dir /path/to/docs

# 禁用自动更新（离线模式，使用本地已有文档）
./cangje-docs-mcp -no-update
```

参数也可以加到 MCP 配置的 `args` 中，例如：

```json
{
  "mcpServers": {
    "cangjie-docs": {
      "command": "/path/to/cangje-docs-mcp",
      "args": ["-dir", "/custom/path/to/CangjieCorpus"]
    }
  }
}
```

## 💡 使用示例

配置完成后，直接在对话中提问即可：

| 场景 | 示例提问 |
|------|---------|
| 基础查询 | 请帮我查找仓颉语言中函数定义的语法 |
| 分类搜索 | 我想了解仓颉语言的基础数据类型 |
| API 查询 | 仓颉标准库中有哪些文件操作相关的 API？ |
| 学习路径 | 我是初学者，请给我推荐仓颉语言的学习顺序 |
| 多版本查询 | 查询仓颉 1.0.0 版本的文档 |

## ⚡ 智能文档分割

系统内置智能文档分割功能，自动将大文档拆分成易于管理的小文档：

- **自动检测**: 超过 15KB 的文档自动触发分割
- **按章节分割**: 按二级标题（`##`）分割，保持内容完整性
- **非递归**: 不递归分割，保持结构体/类的完整性
- **保留关联**: 每个子文档保留父文档 ID，方便追溯完整文档

**统计**（基于完整文档库）：
- 📊 总文档数：4,266 个（含分割后的子文档）
- 📊 被分割文档：201 个，生成子文档 3,383 个
- 📊 **92.3% 的文档大小在 0-5KB 范围**

## 📖 支持的文档结构

系统支持新版仓颉语料文档结构（也兼容旧版目录结构）：

```
CangjieCorpus/
├── manual/              # 基础手册
│   └── source_zh_cn/
├── libs/                # 标准库 API
│   ├── std/             # 标准库
│   └── stdx/            # 扩展标准库
├── ohos/                # OpenHarmony 文档
│   └── zh-cn/
├── tools/               # 开发工具
│   └── source_zh_cn/
└── extra/               # 额外内容
```

## 🛠️ 故障排除

**Q: 首次启动很慢？**
A: 正常现象，系统正在下载仓颉文档（约 8MB），后续启动会很快（仅增量更新）。

**Q: 提示"系统未安装 git"？**
A: 需要先安装 Git。Windows: [git-scm.com](https://git-scm.com/)，Linux: `sudo apt install git`。

**Q: 想使用离线模式？**
A: 在 MCP 配置的 `args` 中添加 `"-no-update"` 即可禁用自动更新。

**Q: 文档下载失败？**
A: 检查网络连接，确保能访问 gitcode.com。也可以手动下载后使用 `-dir` 参数指定目录。

**Q: 找不到 MCP 服务器 / 工具未出现？**
A: 检查配置文件中的可执行文件路径是否正确，确保有执行权限，并确认已重启客户端。

**Q: 如何手动验证服务是否正常？**
A: 直接运行 `./cangje-docs-mcp` 观察是否成功克隆/更新文档，或使用 `-version` 查看版本和文档目录。

## 🧠 进阶提示：常驻语法参考

如果使用不是针对仓颉微调的通用代码模型（如 Qwen3-Code、DeepSeek-Coder、GPT-4），建议在上下文中常驻一份仓颉语法参考：

```bash
# 将 cj_syntax.md 内容追加到 Claude Code 的全局配置
cat cj_syntax.md >> ~/.claude/CLAUDE.md
```

**为什么需要？**
- 通用代码模型对仓颉语法的训练数据有限
- `cj_syntax.md` 提供了仓颉核心语法和特性说明
- 相当于在上下文中常驻语法参考，显著提升模型对仓颉语法的理解准确性

## 📋 更新日志

### v1.4.0 (2026-04-27)

**新功能**
- 🔀 **多版本切换**：支持查询不同版本的仓颉文档（`1.0.0`、`1.1.0`、`0.53.18`）
  - 新增 `cangjie_list_versions` 工具，列出可用版本并标注 latest/default
  - 现有 4 个工具全部添加 `version` 参数，不传则使用默认版本
  - 使用 git worktree 按需加载版本，只在首次查询时创建
  - semver 排序自动识别最新版本，排除 OpenHarmony 等非标准分支
- 🏗️ **完整克隆**：从浅克隆改为完整克隆，为 worktree 多版本提供基础

### v1.2.0 (2025-01-08)

**重大修复**
- ✅ 修复文档分割过度：移除递归分割，保持结构体/类的完整性
- ✅ 修复子分类精度：从 "std" 改为 "std/core" 等完整路径
- ✅ 修复文档统计：移除 Prerequisites 过滤，显示所有文档
- ✅ 修复子分类匹配：使用完整字符串匹配而非首部分匹配

### v1.1.0 (2025-12-30)

**新功能/优化**
- 🌳 **导航树视图**：JSON 改为树形文本格式，新增 `level` 参数控制显示深度
- 📊 **文档列表**：JSON 改为 Markdown 表格格式，支持 `max_items` 参数
- 💰 **Token 效率提升**：导航树节省 ~75%、文档列表节省 ~60%，总体节省 ~82%

**Bug 修复**
- 修复搜索功能：大文档子章节内容未被正确索引
- 修复导航树：子分类和目录节点未正确显示
- 修复文档列表：`max_items` 参数未生效

---

**项目地址**: [github.com/ystyle/cangje-docs-mcp](https://github.com/ystyle/cangje-docs-mcp)
