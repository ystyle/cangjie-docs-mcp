# 仓颉语言文档检索系统 - Claude Code配置教程

这是一个专为Claude Code设计的仓颉语言文档检索MCP服务器，让你能够直接在Claude Code中高效查询仓颉编程语言的所有文档。

## 🚀 快速开始

```bash
# 1. 克隆项目
git clone <repository-url>
cd cangje-docs-mcp

# 2. 下载仓颉文档
git clone https://gitcode.com/Cangjie/CangjieCorpus.git

# 3. 构建项目
go build -o cangje-docs-mcp

# 4. 测试运行
./cangje-docs-mcp -version
```

> **文档源**: [CangjieCorpus](https://gitcode.com/Cangjie/CangjieCorpus) - 仓颉编程语言官方文档仓库

## 🚀 Claude Code配置

### 第一步：构建MCP服务器

```bash
git clone <repository-url>
cd cangje-docs-mcp
go build -o cangje-docs-mcp
```

### 第二步：准备文档目录

#### 方式一：下载官方文档

```bash
# 克隆官方文档仓库
git clone https://gitcode.com/Cangjie/CangjieCorpus.git

# 或者下载压缩包
wget https://gitcode.com/Cangjie/CangjieCorpus/archive/main.zip
unzip main.zip
mv CangjieCorpus-main CangjieCorpus
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

确保你的文档目录结构如下：

```
CangjieCorpus/
├── README.md
├── manual/              # 基础手册
│   ├── first_understanding/
│   ├── basic_data_type/
│   ├── function/
│   └── ...
├── libs/                # 标准库API
│   └── std/
├── tools/               # 开发工具
└── extra/               # 额外内容
```

## 🎉 开始使用

配置完成后，你就可以在Claude Code中自然地查询仓颉语言的所有文档内容了！系统会自动理解你的问题并提供相关的文档内容和建议。