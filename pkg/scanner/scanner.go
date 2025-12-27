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

		// 检查是否需要分割大文档
		splitDocs := s.splitDocumentIfNeeded(doc)

		// 添加分割后的文档到索引
		for _, splitDoc := range splitDocs {
			documents[splitDoc.ID] = splitDoc
		}

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
		// 处理 libs/std 和 libs/stdx
		if len(parts) > 1 {
			return types.CategoryLibs, parts[1]
		}
		return types.CategoryLibs, ""
	case "tools":
		return types.CategoryTools, ""
	case "extra":
		return types.CategoryExtra, ""
	case "ohos":
		// 处理 ohos/zh-cn 等子目录
		if len(parts) > 1 {
			return types.CategoryOhos, parts[1]
		}
		return types.CategoryOhos, ""
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

// splitDocumentIfNeeded 检查文档是否需要分割，如果需要则返回分割后的文档列表
func (s *Scanner) splitDocumentIfNeeded(doc *types.Document) []*types.Document {
	// 如果文档分割未启用或文档小于阈值，直接返回原文档
	if !types.EnableDocumentSplitting || len(doc.Content) < types.LargeDocumentThreshold {
		return []*types.Document{doc}
	}

	// 解析文档的TOC
	toc := s.parseDocumentTOC(doc.Content, doc.ID)

	// 如果只有少数几个章节，且每个章节都不太大，不需要分割
	if len(toc.Sections) <= 5 {
		maxSectionSize := 0
		for _, section := range toc.Sections {
			if section.CharCount > maxSectionSize {
				maxSectionSize = section.CharCount
			}
		}
		if maxSectionSize < types.MaxSectionSize {
			return []*types.Document{doc}
		}
	}

	// 需要分割：为每个主要章节创建独立的文档
	var splitDocs []*types.Document

	for i, section := range toc.Sections {
		// 只处理二级标题(##)及以下的章节
		if section.Level > 2 {
			continue
		}

		// 如果章节内容仍然太大，递归分割
		if section.CharCount > types.MaxSectionSize {
			subSections := s.splitLargeSection(doc, section, toc, i)
			splitDocs = append(splitDocs, subSections...)
		} else {
			// 创建章节文档
			sectionDoc := s.createSectionDocument(doc, section, i)
			splitDocs = append(splitDocs, sectionDoc)
		}
	}

	// 如果分割失败或没有产生任何文档，返回原文档
	if len(splitDocs) == 0 {
		return []*types.Document{doc}
	}

	return splitDocs
}

// parseDocumentTOC 解析文档目录结构
func (s *Scanner) parseDocumentTOC(content, docID string) *types.DocumentTOC {
	lines := strings.Split(content, "\n")
	toc := &types.DocumentTOC{
		DocID:    docID,
		Sections: []types.DocumentSection{},
	}

	var currentSection *types.DocumentSection
	var sectionContent strings.Builder

	for i, line := range lines {
		// 检查是否是标题行
		if strings.HasPrefix(line, "#") {
			// 保存上一个章节
			if currentSection != nil {
				currentSection.Content = sectionContent.String()
				currentSection.CharCount = len(currentSection.Content)
				toc.Sections = append(toc.Sections, *currentSection)
				toc.TotalSize += currentSection.CharCount
			}

			// 解析新章节
			level := 0
			for _, ch := range line {
				if ch == '#' {
					level++
				} else {
					break
				}
			}

			title := strings.TrimSpace(line[level:])
			sectionID := fmt.Sprintf("%s_section_%d", docID, len(toc.Sections)+1)

			currentSection = &types.DocumentSection{
				ID:         sectionID,
				Title:      title,
				Level:      level,
				LineNumber: i + 1,
			}

			sectionContent = strings.Builder{}
		} else if currentSection != nil {
			// 添加内容到当前章节
			if sectionContent.Len() > 0 {
				sectionContent.WriteString("\n")
			}
			sectionContent.WriteString(line)
		}
	}

	// 保存最后一个章节
	if currentSection != nil {
		currentSection.Content = sectionContent.String()
		currentSection.CharCount = len(currentSection.Content)
		toc.Sections = append(toc.Sections, *currentSection)
		toc.TotalSize += currentSection.CharCount
	}

	// 提取文档标题（第一个一级标题）
	if len(toc.Sections) > 0 && toc.Sections[0].Level == 1 {
		toc.Title = toc.Sections[0].Title
	}

	toc.IsSplit = toc.TotalSize >= types.LargeDocumentThreshold

	return toc
}

// splitLargeSection 递归分割过大的章节
func (s *Scanner) splitLargeSection(doc *types.Document, section types.DocumentSection, toc *types.DocumentTOC, sectionIndex int) []*types.Document {
	// 在原始内容中查找该章节下的子章节
	lines := strings.Split(section.Content, "\n")
	var subDocs []*types.Document

	var currentSubSection *types.DocumentSection
	var subContent strings.Builder
	subSectionIndex := 0

	for _, line := range lines {
		// 查找三级标题(###)及以下的子章节
		if strings.HasPrefix(line, "###") {
			// 保存上一个子章节
			if currentSubSection != nil && subContent.Len() > 0 {
				currentSubSection.Content = subContent.String()
				currentSubSection.CharCount = len(currentSubSection.Content)

				// 如果子章节仍然太大，跳过它（避免无限递归）
				if currentSubSection.CharCount < types.MaxSectionSize*2 {
					subDoc := s.createSubSectionDocument(doc, section, *currentSubSection, sectionIndex, subSectionIndex)
					subDocs = append(subDocs, subDoc)
					subSectionIndex++
				}
			}

			// 创建新的子章节
			level := 0
			for _, ch := range line {
				if ch == '#' {
					level++
				} else {
					break
				}
			}

			title := strings.TrimSpace(line[level:])
			subSectionID := fmt.Sprintf("%s_sub_%d_%d", doc.ID, sectionIndex, subSectionIndex)

			currentSubSection = &types.DocumentSection{
				ID:    subSectionID,
				Title: title,
				Level: level,
			}

			subContent = strings.Builder{}
		} else if currentSubSection != nil {
			if subContent.Len() > 0 {
				subContent.WriteString("\n")
			}
			subContent.WriteString(line)
		}
	}

	// 保存最后一个子章节
	if currentSubSection != nil && subContent.Len() > 0 {
		currentSubSection.Content = subContent.String()
		currentSubSection.CharCount = len(currentSubSection.Content)

		if currentSubSection.CharCount < types.MaxSectionSize*2 {
			subDoc := s.createSubSectionDocument(doc, section, *currentSubSection, sectionIndex, subSectionIndex)
			subDocs = append(subDocs, subDoc)
		}
	}

	// 如果没有找到合适的子章节，返回包含整个章节的单个文档
	if len(subDocs) == 0 {
		return []*types.Document{s.createSectionDocument(doc, section, sectionIndex)}
	}

	return subDocs
}

// createSectionDocument 创建章节文档
func (s *Scanner) createSectionDocument(doc *types.Document, section types.DocumentSection, index int) *types.Document {
	sectionID := fmt.Sprintf("%s_%s_%d", doc.ID, sanitizeID(section.Title), index)

	return &types.Document{
		ID:           sectionID,
		Title:        section.Title,
		Category:     doc.Category,
		Subcategory:  doc.Subcategory,
		Description:  s.generateSectionDescription(section.Content),
		FilePath:     doc.FilePath,
		RelativePath: doc.RelativePath,
		Keywords:     s.extractKeywords(section.Content),
		Prerequisites: []string{doc.ID}, // 父文档ID
		RelatedDocs:   []string{},
		Difficulty:    doc.Difficulty,
		FileSize:      int64(len(section.Content)),
		LastModified:  doc.LastModified,
		Content:       section.Content,
		ContentPreview: s.generateContentPreview(section.Content),
	}
}

// createSubSectionDocument 创建子章节文档
func (s *Scanner) createSubSectionDocument(doc *types.Document, parentSection types.DocumentSection, subSection types.DocumentSection, parentIndex, subIndex int) *types.Document {
	subSectionID := fmt.Sprintf("%s_%s_%d_%d", doc.ID, sanitizeID(parentSection.Title), parentIndex, subIndex)

	// 合并父章节标题和子章节标题作为标题
	title := fmt.Sprintf("%s - %s", parentSection.Title, subSection.Title)

	return &types.Document{
		ID:           subSectionID,
		Title:        title,
		Category:     doc.Category,
		Subcategory:  doc.Subcategory,
		Description:  s.generateSectionDescription(subSection.Content),
		FilePath:     doc.FilePath,
		RelativePath: doc.RelativePath,
		Keywords:     s.extractKeywords(subSection.Content),
		Prerequisites: []string{doc.ID},
		RelatedDocs:   []string{},
		Difficulty:    doc.Difficulty,
		FileSize:      int64(len(subSection.Content)),
		LastModified:  doc.LastModified,
		Content:       subSection.Content,
		ContentPreview: s.generateContentPreview(subSection.Content),
	}
}

// generateSectionDescription 为章节生成描述
func (s *Scanner) generateSectionDescription(content string) string {
	lines := strings.Split(content, "\n")
	var descLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// 跳过空行、标题和代码块
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "```") {
			continue
		}
		// 跳过代码行
		if strings.HasPrefix(line, "    ") || strings.HasPrefix(line, "\t") {
			continue
		}

		descLines = append(descLines, line)
		if len(descLines) >= 2 {
			break
		}
	}

	desc := strings.Join(descLines, " ")
	if len(desc) > 150 {
		desc = desc[:150] + "..."
	}

	return desc
}

// sanitizeID 清理ID中的特殊字符
func sanitizeID(title string) string {
	// 移除特殊字符，只保留字母、数字、下划线和连字符
	// 使用字符类匹配中文字符
	reg := regexp.MustCompile(`[^a-zA-Z0-9_\p{Han}-]`)
	id := reg.ReplaceAllString(title, "_")
	// 移除连续的下划线
	reg = regexp.MustCompile(`_+`)
	id = reg.ReplaceAllString(id, "_")
	return strings.Trim(id, "_")
}