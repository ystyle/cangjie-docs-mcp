package scanner

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"cangje-docs-mcp/pkg/types"
)

// Scanner 文档扫描器
type Scanner struct {
	docRoot string
}

// NewScanner 创建新的文档扫描器
func NewScanner(docRoot string) *Scanner {
	if docRoot == "" {
		docRoot = types.DefaultDocumentRootPath
	}
	return &Scanner{
		docRoot: docRoot,
	}
}

// ScanAll 扫描所有文档
func (s *Scanner) ScanAll() (map[string]*types.Document, error) {
	documents := make(map[string]*types.Document)

	err := s.scanDirectory("", documents)
	if err != nil {
		return nil, fmt.Errorf("failed to scan documents: %w", err)
	}

	return documents, nil
}

// scanDirectory 扫描目录
func (s *Scanner) scanDirectory(relativePath string, documents map[string]*types.Document) error {
	fullPath := filepath.Join(s.docRoot, relativePath)

	return filepath.WalkDir(fullPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录和非markdown文件
		if d.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".md") {
			return nil
		}

		// 获取相对路径
		relPath, err := filepath.Rel(s.docRoot, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", path, err)
		}

		// 解析文档
		doc, err := s.parseDocument(path, relPath)
		if err != nil {
			// 记录错误但继续扫描其他文件
			fmt.Printf("Warning: failed to parse document %s: %v\n", path, err)
			return nil
		}

		documents[doc.ID] = doc
		return nil
	})
}

// parseDocument 解析文档文件
func (s *Scanner) parseDocument(fullPath, relativePath string) (*types.Document, error) {
	// 获取文件信息
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// 读取文件内容
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	contentStr := string(content)

	// 解析标题
	title := s.extractTitle(contentStr)

	// 解析元数据
	description := s.extractDescription(contentStr)
	keywords := s.extractKeywords(contentStr)

	// 确定分类和子分类
	category, subcategory := s.determineCategory(relativePath)

	// 生成文档ID
	docID := s.generateID(category, subcategory, filepath.Base(fullPath))

	// 生成内容预览
	contentPreview := s.generateContentPreview(contentStr)

	doc := &types.Document{
		ID:            docID,
		Title:         title,
		Category:      category,
		Subcategory:   subcategory,
		Description:   description,
		FilePath:      fullPath,
		RelativePath:  relativePath,
		Keywords:      keywords,
		Prerequisites: []string{},
		RelatedDocs:   []string{},
		Difficulty:    s.determineDifficulty(relativePath),
		FileSize:      fileInfo.Size(),
		LastModified:  fileInfo.ModTime(),
		Content:       contentStr,
		ContentPreview: contentPreview,
	}

	return doc, nil
}

// extractTitle 提取文档标题
func (s *Scanner) extractTitle(content string) string {
	// 尝试匹配第一个 # 标题
	titleRegex := regexp.MustCompile(`^#\s+(.+)$`)
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if matches := titleRegex.FindStringSubmatch(line); matches != nil {
			return strings.TrimSpace(matches[1])
		}
	}

	// 如果没有找到，使用文件名作为标题
	return "Untitled Document"
}

// extractDescription 提取文档描述
func (s *Scanner) extractDescription(content string) string {
	lines := strings.Split(content, "\n")
	var descriptionLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 跳过标题和空行
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		// 跳过代码块
		if strings.HasPrefix(line, "```") {
			continue
		}

		// 收集描述行
		if len(descriptionLines) < 3 {
			descriptionLines = append(descriptionLines, line)
		} else {
			break
		}
	}

	return strings.Join(descriptionLines, " ")
}

// extractKeywords 提取关键词
func (s *Scanner) extractKeywords(content string) []string {
	var keywords []string

	// 基于内容提取一些常见的技术关键词
	keywordPatterns := []string{
		`\b(函数|function|方法|method)\b`,
		`\b(类|class|对象|object)\b`,
		`\b(接口|interface)\b`,
		`\b(变量|variable|常量|constant)\b`,
		`\b(数组|array|列表|list|集合|set|字典|map)\b`,
		`\b(循环|loop|条件|condition|判断|if|for|while)\b`,
		`\b(字符串|string|整数|integer|浮点|float|布尔|boolean)\b`,
		`\b(并发|concurrency|异步|async|协程|coroutine)\b`,
		`\b(错误|error|异常|exception|处理|handle)\b`,
		`\b(包|package|模块|module|库|library)\b`,
	}

	for _, pattern := range keywordPatterns {
		regex := regexp.MustCompile(pattern)
		matches := regex.FindAllString(content, -1)
		for _, match := range matches {
			match = strings.ToLower(strings.TrimSpace(match))
			if !contains(keywords, match) {
				keywords = append(keywords, match)
			}
		}
	}

	return keywords
}

// determineCategory 确定文档分类
func (s *Scanner) determineCategory(relativePath string) (types.DocumentCategory, string) {
	parts := strings.Split(relativePath, string(filepath.Separator))

	if len(parts) == 0 {
		return types.CategoryManual, ""
	}

	switch parts[0] {
	case "manual":
		if len(parts) > 1 {
			return types.CategoryManual, parts[1]
		}
		return types.CategoryManual, ""
	case "libs":
		return types.CategoryLibs, ""
	case "tools":
		return types.CategoryTools, ""
	case "extra":
		return types.CategoryExtra, ""
	default:
		return types.CategoryManual, ""
	}
}

// determineDifficulty 确定文档难度
func (s *Scanner) determineDifficulty(relativePath string) string {
	if strings.Contains(relativePath, "first_understanding") {
		return "beginner"
	}
	if strings.Contains(relativePath, "basic") {
		return "beginner"
	}
	if strings.Contains(relativePath, "advanced") {
		return "advanced"
	}
	return "intermediate"
}

// generateID 生成文档ID
func (s *Scanner) generateID(category types.DocumentCategory, subcategory, filename string) string {
	base := strings.TrimSuffix(filename, filepath.Ext(filename))
	base = strings.ToLower(strings.ReplaceAll(base, "-", "_"))

	if subcategory != "" {
		return fmt.Sprintf("%s_%s_%s", category, subcategory, base)
	}
	return fmt.Sprintf("%s_%s", category, base)
}

// generateContentPreview 生成内容预览
func (s *Scanner) generateContentPreview(content string) string {
	lines := strings.Split(content, "\n")
	var previewLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// 跳过标题、空行和代码块标记
		if strings.HasPrefix(line, "#") || line == "" || line == "```" {
			continue
		}
		// 跳过代码块
		if strings.HasPrefix(line, "    ") || strings.HasPrefix(line, "\t") {
			continue
		}

		previewLines = append(previewLines, line)
		if len(previewLines) >= 3 {
			break
		}
	}

	preview := strings.Join(previewLines, " ")
	if len(preview) > 200 {
		preview = preview[:200] + "..."
	}

	return preview
}

// GetDocRoot 获取文档根目录
func (s *Scanner) GetDocRoot() string {
	return s.docRoot
}

// contains 检查字符串是否在切片中
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}