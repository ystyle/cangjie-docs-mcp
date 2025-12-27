# 仓颉语言文档检索系统 - Claude Code配置教程

这是一个专为Claude Code设计的仓颉语言文档检索MCP服务器，让你能够直接在Claude Code中高效查询仓颉编程语言的所有文档。

## 🚀 Claude Code配置

### 第一步：下载可mcp可执行文件
- [下载](https://github.com/ystyle/cangje-docs-mcp/releases)仓颉文件mcp可执行文

### 第二步：准备文档目录

#### 方式一：下载官方文档

```bash
# 克隆官方文档仓库
git clone https://gitcode.com/Cangjie/CangjieCorpus.git
```

#### 方式二：使用现有文档

将仓颉语言文档复制到可执行文件所在目录：

```bash
# 确保在项目根目录下有CangjieCorpus目录
# 或者使用 -dir 参数指定文档目录
```

### 第三步：配置Claude Code

1. 打开Claude Code设置
2. 找到MCP服务器配置
3. 添加以下配置：

```json
{
  "mcpServers": {
    "cangjie-docs": {
      "command": "/path/to/cangje-docs-mcp",
      "args": ["-dir", "/path/to/CangjieCorpus"]
    }
  }
}
```

### 第四步：重启Claude Code

重启Claude Code以加载新的MCP服务器。

> 在`CLAUDE.md` 添加 本仓库里的 `cj_syntax.md`, 相当于在上下文常驻一份仓颉基础语法（如果针对仓颉微调过的llm就可以不需要这个）


## 💡 在Claude Code中使用

### 基础查询

```
请帮我查找仓颉语言中函数定义的语法
```

### 分类搜索

```
我想了解仓颉语言的基础数据类型
```

### 学习路径

```
我是初学者，请给我推荐仓颉语言的学习顺序
```

### 具体问题

```
仓颉语言中如何处理并发编程？
```

### API查询

```
仓颉标准库中有哪些文件操作相关的API？
```

## 🎯 常用查询示例

### 语法查询
- "仓颉语言的变量声明语法"
- "如何定义一个类"
- "仓颉的控制流语句"

### 概念学习
- "解释仓颉语言的泛型编程"
- "仓颉的错误处理机制"
- "仓颉的包管理系统"

### 实践问题
- "如何在仓颉中读写文件"
- "仓颉网络编程的例子"
- "仓颉与JavaScript交互"

### API参考
- "仓颉String类的方法"
- "仓颉数组操作API"
- "仓颉HTTP客户端使用"

## 🔧 高级功能

### 指定分类搜索
```
请在manual分类中搜索关于继承的内容
```

### 获取相关建议
```
基于"函数定义"这个主题，推荐相关的学习内容
```

### 导航浏览
```
显示仓颉文档的完整目录结构
```

## ⚡ 智能文档分割

系统内置了智能文档分割功能，自动将大文档拆分成易于管理的小文档，优化AI处理效率。

### 分割策略

- **自动检测**: 超过15KB的文档自动触发分割
- **按章节分割**: 优先按二级标题(##)分割，保持内容完整性
- **递归处理**: 超过10KB的章节按三级标题(###)进一步细分
- **保留关联**: 每个子文档保留父文档ID，方便追溯完整文档

### 实际效果

以典型的仓颉文档为例：

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
- 📊 92.3%的文档大小在0-5KB范围

### 优势

- ✅ **精准搜索**: 搜索结果直接定位到相关章节，无需翻阅大文档
- ✅ **高效处理**: 文档大小适中，大幅降低AI处理压力
- ✅ **节省Token**: 只加载需要的章节内容，避免浪费
- ✅ **保持完整**: 通过父文档ID可追溯完整文档结构

## 📋 检查配置

### 验证版本信息

```bash
./cangje-docs-mcp -version
```

应该显示：
```
仓颉语言文档检索系统 v1.0.0
基于MCP协议的本地文档检索服务器
文档版本: v1.0.0 (发布日期: 2025-07-01)
```

### 查看帮助

```bash
./cangje-docs-mcp -help
```

## 🛠️ 故障排除

### 常见问题

**Q: Claude Code找不到MCP服务器**
A: 检查配置文件中的路径是否正确，确保可执行文件有执行权限

**Q: 搜索结果为空**
A: 确认文档目录存在且包含仓颉文档，检查文档目录权限

**Q: 启动时提示文档目录不存在**
A: 使用 `-dir` 参数指定正确的文档目录路径

**Q: MCP服务器启动失败**
A: 检查配置格式是否正确，确保使用了正确的参数格式

### 正确的MCP配置格式

```json
{
  "mcpServers": {
    "cangje-docs": {
      "type": "stdio",
      "command": "/path/to/cangje-docs-mcp",
      "args": ["-dir", "/path/to/CangjieCorpus"],
      "env": {}
    }
  }
}
```

**重要**: 确保包含 `"type": "stdio"` 和 `"env": {}` 字段。

### 调试模式

```bash
# 直接运行MCP服务器查看启动日志
./cangje-docs-mcp -dir /path/to/CangjieCorpus

# 检查版本信息
./cangje-docs-mcp -version

# 查看帮助信息
./cangje-docs-mcp -help
```

## 📖 文档目录结构

支持新版仓颉语料文档结构（推荐）：

```
CangjieCorpus/
├── README.md
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

也支持旧版目录结构：

```
CangjieCorpus/
├── manual/              # 基础手册
├── libs/                # 标准库API
│   └── std/
├── tools/               # 开发工具
└── extra/               # 额外内容
```

## 🎉 开始使用

配置完成后，你就可以在Claude Code中自然地查询仓颉语言的所有文档内容了！系统会自动理解你的问题并提供相关的文档内容和建议。
