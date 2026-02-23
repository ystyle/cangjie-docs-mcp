# 仓颉语言文档检索系统 - Claude Code配置教程

这是一个专为Claude Code设计的仓颉语言文档检索MCP服务器，让你能够直接在Claude Code中高效查询仓颉编程语言的所有文档。

## ✨ 核心特性

- 🚀 **开箱即用** - 首次运行自动下载文档，无需手动配置
- 🔄 **自动更新** - 每次启动自动更新到最新版本
- ⚡ **智能分割** - 大文档自动分割成小文档，优化AI处理效率
- 🎯 **精准搜索** - 支持全文搜索、分类搜索、章节定位
- 📚 **完整覆盖** - 支持手册、标准库(std/stdx)、OpenHarmony、工具文档
- 💰 **Token 高效** - 树形文本和表格格式，节省 70%+ Token 消耗

## 🚀 快速开始

### 第一步：下载可执行文件

从 [Releases](https://github.com/ystyle/cangje-docs-mcp/releases) 下载对应平台的可执行文件。

### 第二步：配置Claude Code

打开Claude Code设置，添加以下MCP服务器配置：

```json
{
  "mcpServers": {
    "cangjie-docs": {
      "type": "stdio",
      "command": "/path/to/cangje-docs-mcp",
      "args": [],
      "env": {}
    }
  }
}
```

### 第三步：安装 Skill（推荐）

安装 Skill 后，Claude 可以更智能地检索仓颉文档，支持多种搜索模式：

```bash
# GitHub（推荐）
npx skills add ystyle/cangjie-docs-mcp

# 国内用户（AtomGit 镜像）
npx skills add https://atomgit.com/Cangjie-SIG/cangjie-docs-mcp
```

Skill 支持 4 种搜索模式：
- ⚡ **直接搜索**：精确 API 查询（如 "String.split"）
- 🧠 **PageIndex 智能检索**：模糊查询（如 "怎么截取字符串"）
- ⚖️ **混合模式**：自动切换最优策略
- 🧭 **探索模式**：学习引导

### 第四步：重启 Claude Code

重启后即可使用！系统会自动：
- ✅ 下载仓颉文档到默认位置
- ✅ 每次启动时更新文档
- ✅ 自动分割大文档优化处理

就这么简单！🎉

> 💡 **推荐配置**：如果使用的不是针对仓颉微调的模型（如 Qwen3-Code、DeepSeek-Coder 等通用代码模型），建议在 `CLAUDE.md` 中添加本仓库的 `cj_syntax.md` 文件内容。这相当于在上下文中常驻一份仓颉基础语法，可以显著提升模型对仓颉语法的理解准确性。

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

## ⚙️ 高级配置

### 自定义文档目录

如果需要使用自定义文档目录：

```json
{
  "mcpServers": {
    "cangjie-docs": {
      "type": "stdio",
      "command": "/path/to/cangje-docs-mcp",
      "args": ["-dir", "/custom/path/to/CangjieCorpus"],
      "env": {}
    }
  }
}
```

### 禁用自动更新

如需禁用自动更新（使用本地已有文档）：

```json
{
  "mcpServers": {
    "cangjie-docs": {
      "type": "stdio",
      "command": "/path/to/cangje-docs-mcp",
      "args": ["-no-update"],
      "env": {}
    }
  }
}
```

### 完整参数说明

```bash
# 查看版本和文档目录
./cangje-docs-mcp -version

# 查看帮助信息
./cangje-docs-mcp -help

# 使用自定义文档目录
./cangje-docs-mcp -dir /path/to/docs

# 禁用自动更新
./cangje-docs-mcp -no-update
```

## 💡 在Claude Code中使用

配置完成后，你可以这样使用：

### 基础查询
```
请帮我查找仓颉语言中函数定义的语法
```

### 分类搜索
```
我想了解仓颉语言的基础数据类型
```

### API查询
```
仓颉标准库中有哪些文件操作相关的API？
```

### 学习路径
```
我是初学者，请给我推荐仓颉语言的学习顺序
```

## ⚡ 智能文档分割

系统内置了智能文档分割功能，自动将大文档拆分成易于管理的小文档：

### 分割策略

- **自动检测**: 超过15KB的文档自动触发分割
- **按章节分割**: 按二级标题(##)分割，保持内容完整性
- **非递归**: 不再递归分割，保持结构体/类的完整性
- **保留关联**: 每个子文档保留父文档ID，方便追溯完整文档

### 实际效果

| 原始文档 | 大小 | 分割结果 |
|---------|------|---------|
| cj-common-types.md | 430KB | → 190个子文档 |
| overflow_package_interfaces.md | 310KB | → 已分割 |
| ast_package_classes.md | 232KB | → 92个子文档 |
| math_package_funcs.md | 122KB | → 163个子文档 |

**统计数据**（基于完整文档库）：
- 📊 总文档数：4,266个（含分割后的子文档）
- 📊 被分割文档：201个
- 📊 生成的子文档：3,383个
- 📊 **92.3%的文档大小在0-5KB范围**

### 优势

- ✅ **精准搜索**: 搜索结果直接定位到相关章节
- ✅ **高效处理**: 文档大小适中，大幅降低AI处理压力
- ✅ **节省Token**: 只加载需要的章节内容
- ✅ **保持完整**: 通过父文档ID可追溯完整文档结构

## 📖 支持的文档结构

系统支持新版仓颉语料文档结构：

```
CangjieCorpus/
├── manual/              # 基础手册
│   ├── source_zh_cn/
│   └── ...
├── libs/                # 标准库API
│   ├── std/             # 标准库
│   └── stdx/            # 扩展标准库
├── ohos/                # OpenHarmony文档
│   └── zh-cn/
├── tools/               # 开发工具
│   └── source_zh_cn/
└── extra/               # 额外内容
```

也兼容旧版目录结构。

## 🛠️ 故障排除

### 常见问题

**Q: 首次启动很慢？**
A: 正常现象，系统正在下载仓颉文档（约8MB），后续启动会很快（仅更新增量）。

**Q: 提示"系统未安装 git"？**
A: 需要先安装 Git。Windows: [git-scm.com](https://git-scm.com/)，Linux: `sudo apt install git`

**Q: 想使用离线模式？**
A: 在MCP配置中添加 `"-no-update"` 参数即可禁用自动更新。

**Q: 文档下载失败？**
A: 检查网络连接，确保能访问 gitcode.com。也可以手动下载后使用 `-dir` 参数指定目录。

**Q: Claude Code找不到MCP服务器？**
A: 检查配置文件中的可执行文件路径是否正确，确保有执行权限。

### 调试模式

```bash
# 查看版本和文档目录
./cangje-docs-mcp -version

# 手动测试文档下载和更新
./cangje-docs-mcp

# 使用自定义目录测试
./cangje-docs-mcp -dir /path/to/CangjieCorpus
```

## 🎉 开始使用

配置完成后，你就可以在Claude Code中自然地查询仓颉语言的所有文档内容了！系统会自动理解你的问题并提供相关的文档内容和建议。

### 💡 进阶提示

**推荐配置**：如果使用的不是针对仓颉微调的模型（如 Qwen3-Code、DeepSeek-Coder、GPT-4 等通用代码模型），建议在 `CLAUDE.md` 中添加本仓库的 `cj_syntax.md` 文件内容。

**为什么需要？**
- 通用代码模型对仓颉语法的训练数据有限
- `cj_syntax.md` 提供了仓颉核心语法和特性说明
- 相当于在上下文中常驻一份语法参考，提升准确性

**如何添加？**
```bash
# 将 cj_syntax.md 的内容复制到你的 CLAUDE.md 文件中
cat cj_syntax.md >> ~/.claude/CLAUDE.md
```

这样模型就能更好地理解仓颉语法，提供更准确的代码建议！

---

## 📋 更新日志

### v1.2.0 (2025-01-08)

**重大修复**
- ✅ 修复文档分割过度：移除递归分割，保持结构体/类的完整性
  - 现在只在二级标题(##)层级分割，不再按三级标题(###)递归分割
  - 修复前：`struct String` (37KB) 被分割成数百个方法碎片
  - 修复后：`struct String` 保持为完整文档，包含70+个方法
- ✅ 修复子分类精度：从 "std" 改为 "std/core" 等完整路径
  - `list_documents(subcategory="std/core")` 现在能正确过滤
- ✅ 修复文档统计：移除 Prerequisites 过滤，显示所有文档
  - std/core: 5 docs → **115 docs** ✅
  - std/collection: 10 docs → **68 docs** ✅
  - std/math: 2 docs → **174 docs** ✅
- ✅ 修复子分类匹配：使用完整字符串匹配而非首部分匹配

**影响范围**
- 导航树视图统计现在准确（包含所有分割后的文档）
- 文档列表现在显示完整的结构体/类文档
- 搜索功能现在能找到完整的类型定义

### v1.1.0 (2025-12-30)

**Bug 修复**
- 修复搜索功能：大文档的子章节内容未被正确索引
- 修复导航树：子分类和目录节点未正确显示
- 修复文档列表：`max_items` 参数未生效

**新功能/优化**
- 🌳 **导航树视图**：从 JSON 改为树形文本格式（类似 `tree` 命令）
  - 添加 `level` 参数控制显示深度（1=子分类，2=目录，3=文档）
  - 显示所有文档（包括分割后的子文档），统计准确
- 📊 **文档列表**：从 JSON 改为 Markdown 表格格式
  - 支持自定义返回数量（`max_items` 参数）
  - 自动截断过长描述，优化显示
- 💰 **Token 效率提升**：
  - 导航树：节省 ~75% Token（20000 → 5000）
  - 文档列表：节省 ~60% Token（8000 → 3000）
  - 总体节省：~82% Token 消耗

**使用示例**
```bash
# 只看子分类（~200 tokens）
get_document_overview(category="libs", view_type="navigation", level=1)

# 查看包目录（~2000 tokens）
get_document_overview(category="libs", view_type="navigation", level=2)

# 列出前10个文档（表格格式）
list_documents(category="libs", subcategory="std", max_items=10)
```

---

**项目地址**: [github.com/ystyle/cangje-docs-mcp](https://github.com/ystyle/cangje-docs-mcp)
