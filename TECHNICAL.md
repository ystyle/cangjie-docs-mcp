# 仓颉语言文档检索系统 - 技术文档

## 系统架构

### 整体设计

```
用户 (Claude Code) → MCP协议 → 仓颉文档MCP服务器 → 本地文档系统
                     ↓
              stdio通信协议
                     ↓
              文档检索引擎
```

### 文档源

- **官方文档仓库**: https://gitcode.com/Cangjie/CangjieCorpus
- **文档版本**: 仓颉编程语言 v1.0.0 (发布日期: 2025-07-01)
- **文档格式**: Markdown
- **文档结构**: manual/ (基础手册), libs/ (标准库), tools/ (工具), extra/ (额外内容)

### 核心组件

1. **文档扫描器 (Scanner)** - 自动扫描和解析文档
2. **搜索引擎 (Search)** - 三级搜索算法实现
3. **MCP服务器** - 协议处理和接口暴露
4. **索引系统** - 倒排索引和元数据管理

## 数据结构

### 文档分类

```go
type DocumentCategory string

const (
    CategoryManual DocumentCategory = "manual" // 基础手册
    CategoryLibs   DocumentCategory = "libs"   // 标准库API
    CategoryTools  DocumentCategory = "tools"  // 开发工具
    CategoryExtra  DocumentCategory = "extra"  // 额外内容
)
```

### 文档元数据

```go
type Document struct {
    ID            string           `json:"id"`
    Title         string           `json:"title"`
    Category      DocumentCategory `json:"category"`
    Subcategory   string           `json:"subcategory"`
    Description   string           `json:"description"`
    FilePath      string           `json:"file_path"`
    Keywords      []string         `json:"keywords"`
    Content       string           `json:"content"`
    // ... 其他字段
}
```

## 搜索算法

### 三级搜索策略

1. **精确关键词匹配** (权重: 10.0)
   - 匹配文档关键词标签
   - 完全匹配时直接返回

2. **标题描述匹配** (权重: 8.0/6.0)
   - 在标题和描述中搜索
   - 部分匹配和同义词处理

3. **内容模糊匹配** (权重: 3.0)
   - 在文档内容中搜索
   - 使用简单文本匹配

### 相关性评分

```go
func calculateRelevance(query, document string) float64 {
    score := 0.0

    // 关键词匹配 (最高权重)
    for _, keyword := range document.Keywords {
        if strings.Contains(strings.ToLower(query), strings.ToLower(keyword)) {
            score += ExactMatchWeight
        }
    }

    // 标题匹配
    if strings.Contains(strings.ToLower(document.Title), strings.ToLower(query)) {
        score += TitleMatchWeight
    }

    // 描述匹配
    if strings.Contains(strings.ToLower(document.Description), strings.ToLower(query)) {
        score += DescriptionWeight
    }

    return score
}
```

## MCP接口实现

### Resources

| URI模板 | 功能 | 返回类型 |
|---------|------|----------|
| `cangjie://map` | 完整文档分类结构 | JSON |
| `cangjie://navigation/{category}` | 分类目录树 | JSON |
| `cangjie://navigation/all` | 完整导航树 | JSON |
| `cangjie://category/{category}` | 分类文档列表 | JSON |
| `cangjie://document/{doc_id}` | 文档完整内容 | Markdown |

### Tools

| 工具名 | 功能 | 参数 |
|--------|------|------|
| `search_documents` | 搜索文档 | query, category, max_results, min_confidence |
| `suggest_documents` | 智能建议 | context, suggestion_type, max_suggestions |
| `get_document_content` | 获取文档 | doc_id, include_metadata |

## 性能优化

### 索引构建

```go
// 倒排索引结构
type SearchEngine struct {
    documents    map[string]*types.Document
    keywordIndex map[string][]string // 关键词到文档ID的映射
}
```

### 缓存策略

- **文档内容缓存**: 避免重复文件读取
- **搜索结果缓存**: 缓存常用查询结果
- **索引预构建**: 启动时构建完整索引

### 内存优化

- **延迟加载**: 大文档内容按需加载
- **索引压缩**: 使用压缩算法减少内存占用
- **垃圾回收**: 及时释放不需要的资源

## 配置系统

### 零配置设计

```go
// 默认文档根目录 - 可执行文件所在目录的CangjieCorpus
var DefaultDocumentRootPath = func() string {
    if exe, err := os.Executable(); err == nil {
        exeDir := filepath.Dir(exe)
        return filepath.Join(exeDir, "CangjieCorpus")
    }
    return "./CangjieCorpus" // fallback
}()
```

### 命令行参数

```bash
./cangje-docs-mcp [选项]

选项:
  -dir string    仓颉文档根目录路径
  -version       显示版本信息
  -help          显示帮助信息
```

## 版本检测

### 自动版本提取

```go
func getDocumentVersion(docRoot string) string {
    readmePath := filepath.Join(docRoot, "README.md")

    // 匹配仓颉版本信息：仓颉编程语言 v1.0.0（对应官网文档发布日期：2025-07-01）
    if strings.Contains(line, "仓颉编程语言") {
        versionPattern := regexp.MustCompile(`仓颉编程语言\s+([vV]?\d+(?:\.\d+)*)`)
        // 提取版本和日期信息
    }
}
```

## 错误处理

### 错误类型

1. **文档不存在** - 指定的文档ID无效
2. **目录访问错误** - 无法访问文档目录
3. **索引构建失败** - 文档解析错误
4. **搜索超时** - 搜索操作超时

### 日志记录

```go
log.Printf("开始扫描文档目录: %s", docRoot)
log.Printf("文档扫描完成，共发现 %d 个文档", len(documents))
log.Printf("服务器已启动，已加载 %d 个文档", len(s.documents))
```

## 扩展性

### 插件化设计

- **搜索算法插件**: 可插拔的搜索策略
- **文档解析器**: 支持多种文档格式
- **建议引擎**: 可配置的建议算法

### 多语言支持

- **国际化框架**: 支持多语言界面
- **本地化搜索**: 支持多语言内容搜索
- **文化适配**: 考虑不同语言的文化差异

## 安全考虑

### 输入验证

- **参数检查**: 验证所有输入参数
- **路径遍历防护**: 防止目录遍历攻击
- **内容过滤**: 过滤恶意内容

### 权限控制

- **文件访问权限**: 限制文件系统访问范围
- **资源限制**: 限制内存和CPU使用
- **请求频率限制**: 防止滥用

## 测试策略

### 单元测试

```go
func TestSearchEngine(t *testing.T) {
    engine := NewSearchEngine()
    // 测试搜索功能
}

func TestDocumentScanner(t *testing.T) {
    scanner := NewScanner("/test/path")
    // 测试文档扫描
}
```

### 集成测试

- **MCP协议测试**: 验证MCP接口正确性
- **端到端测试**: 完整流程测试
- **性能测试**: 响应时间和并发测试

### 测试覆盖率

- **代码覆盖率**: 目标 > 90%
- **分支覆盖率**: 目标 > 85%
- **功能覆盖率**: 所有核心功能

## 部署和维护

### 部署方式

1. **二进制部署**: 直接运行编译后的可执行文件
2. **容器化部署**: 使用Docker容器部署
3. **服务化部署**: 作为系统服务运行

### 监控指标

- **性能指标**: 响应时间、吞吐量
- **业务指标**: 搜索成功率、用户满意度
- **系统指标**: CPU、内存、磁盘使用

### 维护流程

```bash
# 更新文档
1. 更新CangjieCorpus目录内容
2. 重启MCP服务器
3. 验证功能正常

# 版本升级
1. 备份当前版本
2. 部署新版本
3. 验证兼容性
```