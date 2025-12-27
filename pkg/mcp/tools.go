package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"cangje-docs-mcp/pkg/types"
	"github.com/mark3labs/mcp-go/mcp"
)

// registerTools 注册所有工具
func (s *CangJieDocServer) registerTools() {
	// 文档总览工具
	overviewTool := mcp.NewTool("get_document_overview",
		mcp.WithDescription("获取文档总览和导航结构"),
		mcp.WithString("view_type",
			mcp.Description("视图类型 (默认overview)"),
			mcp.Enum("overview", "map", "navigation", "tree"),
		),
		mcp.WithString("category",
			mcp.Required(),
			mcp.Description("指定分类 (manual/libs/tools/extra/ohos)"),
			mcp.Enum("manual", "libs", "tools", "extra", "ohos"),
		),
		mcp.WithNumber("max_items",
			mcp.Description("最大显示条目数 (默认50)"),
		),
	)
	s.server.AddTool(overviewTool, s.handleGetDocumentOverview)

	// 文档列表工具
	listTool := mcp.NewTool("list_documents",
		mcp.WithDescription("列出分类或子分类的文档"),
		mcp.WithString("category",
			mcp.Required(),
			mcp.Description("主分类"),
			mcp.Enum("manual", "libs", "tools", "extra", "ohos"),
		),
		mcp.WithString("subcategory",
			mcp.Description("子分类，留空则列出整个分类"),
		),
		mcp.WithString("sort_by",
			mcp.Description("排序方式 (默认title)"),
			mcp.Enum("title", "difficulty", "last_modified"),
		),
		mcp.WithBoolean("include_preview",
			mcp.Description("是否包含内容预览 (默认false)"),
		),
	)
	s.server.AddTool(listTool, s.handleListDocuments)

	// 搜索文档工具
	searchTool := mcp.NewTool("search_documents",
		mcp.WithDescription("搜索仓颉语言文档。支持单个关键词或多个关键词（用空格分隔，使用AND逻辑）"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("搜索查询词。单个关键词或多关键词（空格分隔，AND匹配）"),
		),
		mcp.WithString("category",
			mcp.Description("可选的分类过滤 (manual/libs/tools/extra/ohos)"),
			mcp.Enum("manual", "libs", "tools", "extra", "ohos"),
		),
		mcp.WithNumber("max_results",
			mcp.Description("最大结果数 (默认10)"),
		),
		mcp.WithNumber("min_confidence",
			mcp.Description("最小置信度 (默认0.3)"),
		),
	)
	s.server.AddTool(searchTool, s.handleSearchDocuments)

	// 获取文档内容工具
	contentTool := mcp.NewTool("get_document_content",
		mcp.WithDescription("获取指定文档的完整内容"),
		mcp.WithString("doc_id",
			mcp.Required(),
			mcp.Description("文档ID"),
		),
		mcp.WithBoolean("include_metadata",
			mcp.Description("是否包含元数据 (默认true)"),
		),
		mcp.WithString("format",
			mcp.Description("输出格式 (默认markdown)"),
			mcp.Enum("markdown", "json", "plain"),
		),
		mcp.WithString("section",
			mcp.Description("获取特定章节 (如 '1.1', '2.3')"),
		),
	)
	s.server.AddTool(contentTool, s.handleGetDocumentContent)
}

// handleSearchDocuments 处理文档搜索
func (s *CangJieDocServer) handleSearchDocuments(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 使用新的API获取参数
	query, err := request.RequireString("query")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// 可选参数
	var category types.DocumentCategory
	if cat, ok := request.GetArguments()["category"].(string); ok && cat != "" {
		category = types.DocumentCategory(cat)
	}

	maxResults := types.DefaultMaxResults
	if mr, ok := request.GetArguments()["max_results"].(float64); ok {
		maxResults = int(mr)
	}

	minConfidence := types.DefaultMinConfidence
	if mc, ok := request.GetArguments()["min_confidence"].(float64); ok {
		minConfidence = mc
	}

	// 构建搜索请求
	searchReq := types.SearchRequest{
		Query:        query,
		MaxResults:   maxResults,
		MinConfidence: minConfidence,
		Category:     category,
	}

	// 执行搜索
	results := s.searchEngine.Search(searchReq)

	// 格式化结果
	var formattedResults []map[string]interface{}
	for _, result := range results {
		formattedResults = append(formattedResults, map[string]interface{}{
			"document": map[string]interface{}{
				"id":           result.Document.ID,
				"title":        result.Document.Title,
				"category":     result.Document.Category,
				"subcategory":  result.Document.Subcategory,
				"description":  result.Document.Description,
				"difficulty":   result.Document.Difficulty,
				"keywords":     result.Document.Keywords,
				"relative_path": result.Document.RelativePath,
			},
			"score":     result.Score,
			"match_type": result.MatchType,
			"match_text": result.MatchText,
		})
	}

	response := map[string]interface{}{
		"query":   query,
		"count":   len(results),
		"results": formattedResults,
	}

	data, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal response: %v", err)), nil
	}

	return mcp.NewToolResultText(string(data)), nil
}

// handleGetDocumentOverview 处理文档总览请求
func (s *CangJieDocServer) handleGetDocumentOverview(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取必填参数
	categoryStr, err := request.RequireString("category")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	category := types.DocumentCategory(categoryStr)

	// 获取可选参数
	viewType := "overview"
	if vt, ok := request.GetArguments()["view_type"].(string); ok && vt != "" {
		viewType = vt
	}

	maxItems := 50
	if mi, ok := request.GetArguments()["max_items"].(float64); ok {
		maxItems = int(mi)
	}

	// 根据视图类型生成不同的响应
	var response interface{}

	switch viewType {
	case "map":
		// 生成文档地图
		response = s.generateDocumentMap(category, maxItems)
	case "navigation", "tree":
		// 生成导航树
		response = s.generateNavigationTree(category, maxItems)
	default: // overview
		// 生成总览
		response = s.generateOverview(category, maxItems)
	}

	data, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal response: %v", err)), nil
	}

	return mcp.NewToolResultText(string(data)), nil
}

// handleListDocuments 处理文档列表请求
func (s *CangJieDocServer) handleListDocuments(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取必需参数
	category, err := request.RequireString("category")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// 获取可选参数
	subcategory := ""
	if sub, ok := request.GetArguments()["subcategory"].(string); ok {
		subcategory = sub
	}

	sortBy := "title"
	if sb, ok := request.GetArguments()["sort_by"].(string); ok {
		sortBy = sb
	}

	includePreview := false
	if ip, ok := request.GetArguments()["include_preview"].(bool); ok {
		includePreview = ip
	}

	// 筛选文档
	var documents []*types.Document
	for _, doc := range s.documents {
		if string(doc.Category) == category {
			if subcategory == "" || doc.Subcategory == subcategory {
				documents = append(documents, doc)
			}
		}
	}

	// 排序
	s.sortDocuments(documents, sortBy)

	// 限制结果数量
	if len(documents) > 100 {
		documents = documents[:100]
	}

	// 格式化结果
	var formattedDocs []map[string]interface{}
	for _, doc := range documents {
		docInfo := map[string]interface{}{
			"id":           doc.ID,
			"title":        doc.Title,
			"subcategory":  doc.Subcategory,
			"description":  doc.Description,
			"difficulty":   doc.Difficulty,
			"keywords":     doc.Keywords,
			"relative_path": doc.RelativePath,
			"last_modified": doc.LastModified.Format("2006-01-02 15:04:05"),
		}

		if includePreview {
			// 包含内容预览（前200字符）
			preview := doc.Content
			if len(preview) > 200 {
				preview = preview[:200] + "..."
			}
			docInfo["content_preview"] = preview
		}

		formattedDocs = append(formattedDocs, docInfo)
	}

	response := map[string]interface{}{
		"category":     category,
		"subcategory":  subcategory,
		"sort_by":      sortBy,
		"count":        len(formattedDocs),
		"documents":    formattedDocs,
	}

	data, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal response: %v", err)), nil
	}

	return mcp.NewToolResultText(string(data)), nil
}

// handleGetDocumentContent 处理获取文档内容
func (s *CangJieDocServer) handleGetDocumentContent(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取必需参数
	docID, err := request.RequireString("doc_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// 获取可选参数
	includeMetadata := true
	if im, ok := request.GetArguments()["include_metadata"].(bool); ok {
		includeMetadata = im
	}

	format := "markdown"
	if f, ok := request.GetArguments()["format"].(string); ok && f != "" {
		format = f
	}

	section := ""
	if sec, ok := request.GetArguments()["section"].(string); ok {
		section = sec
	}

	// 查找文档
	doc, exists := s.documents[docID]
	if !exists {
		return mcp.NewToolResultError(fmt.Sprintf("document not found: %s", docID)), nil
	}

	// 处理内容
	content := doc.Content
	if section != "" {
		// 提取特定章节
		content = s.extractSection(doc.Content, section)
	}

	// 根据格式返回结果
	if format == "json" {
		response := map[string]interface{}{
			"document_id": docID,
			"title":       doc.Title,
			"category":    doc.Category,
			"subcategory": doc.Subcategory,
			"content":     content,
		}

		if includeMetadata {
			response["metadata"] = map[string]interface{}{
				"description":   doc.Description,
				"difficulty":    doc.Difficulty,
				"keywords":      doc.Keywords,
				"relative_path": doc.RelativePath,
				"file_size":     doc.FileSize,
				"last_modified": doc.LastModified.Format("2006-01-02 15:04:05"),
			}
		}

		data, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to marshal response: %v", err)), nil
		}
		return mcp.NewToolResultText(string(data)), nil
	} else if format == "plain" {
		// 纯文本格式
		if includeMetadata {
			content = fmt.Sprintf(`标题: %s
分类: %s/%s
难度: %s
描述: %s

%s`, doc.Title, string(doc.Category), doc.Subcategory, doc.Difficulty, doc.Description, content)
		}
		return mcp.NewToolResultText(content), nil
	} else { // markdown
		// Markdown格式
		if includeMetadata {
			content = fmt.Sprintf(`# %s

## 元数据
- **分类**: %s
- **子分类**: %s
- **难度**: %s
- **文件路径**: %s
- **最后修改**: %s
- **关键词**: %s

## 描述
%s

## 内容
%s`,
				doc.Title,
				string(doc.Category),
				doc.Subcategory,
				doc.Difficulty,
				doc.RelativePath,
				doc.LastModified.Format("2006-01-02 15:04:05"),
				strings.Join(doc.Keywords, ", "),
				doc.Description,
				content)
		}
		return mcp.NewToolResultText(content), nil
	}
}

// 辅助函数

// generateOverview 生成文档总览
func (s *CangJieDocServer) generateOverview(category types.DocumentCategory, maxItems int) map[string]interface{} {
	// 统计信息
	totalDocs := len(s.documents)
	categoryStats := make(map[types.DocumentCategory]int)
	subcategoryStats := make(map[string]map[string]int)

	for _, doc := range s.documents {
		categoryStats[doc.Category]++
		if subcategoryStats[string(doc.Category)] == nil {
			subcategoryStats[string(doc.Category)] = make(map[string]int)
		}
		subcategoryStats[string(doc.Category)][doc.Subcategory]++
	}

	// 构建响应
	response := map[string]interface{}{
		"total_documents": totalDocs,
		"categories":      make([]map[string]interface{}, 0),
		"generated_at":    time.Now().Format("2006-01-02 15:04:05"),
	}

	// 添加分类信息
	categories := []types.DocumentCategory{types.CategoryManual, types.CategoryLibs, types.CategoryTools, types.CategoryExtra, types.CategoryOhos}
	for _, cat := range categories {
		if category != "" && cat != category {
			continue
		}

		catInfo := map[string]interface{}{
			"name":        cat,
			"display_name": types.CategoryNames[cat],
			"count":       categoryStats[cat],
			"subcategories": make([]map[string]interface{}, 0),
		}

		// 添加子分类信息
		if subcats, ok := subcategoryStats[string(cat)]; ok {
			for subcat, count := range subcats {
				catInfo["subcategories"] = append(catInfo["subcategories"].([]map[string]interface{}), map[string]interface{}{
					"name":  subcat,
					"count": count,
				})
			}
		}

		response["categories"] = append(response["categories"].([]map[string]interface{}), catInfo)
	}

	return response
}

// generateDocumentMap 生成文档地图
func (s *CangJieDocServer) generateDocumentMap(category types.DocumentCategory, maxItems int) map[string]interface{} {
	// 构建分类->子分类->文档的层次结构
	docMap := make(map[string]map[string][]map[string]interface{})

	for _, doc := range s.documents {
		if category != "" && doc.Category != category {
			continue
		}

		catStr := string(doc.Category)
		if docMap[catStr] == nil {
			docMap[catStr] = make(map[string][]map[string]interface{})
		}
		if docMap[catStr][doc.Subcategory] == nil {
			docMap[catStr][doc.Subcategory] = make([]map[string]interface{}, 0)
		}

		// 限制每个子分类的文档数量
		if len(docMap[catStr][doc.Subcategory]) < maxItems/5 {
			docMap[catStr][doc.Subcategory] = append(docMap[catStr][doc.Subcategory], map[string]interface{}{
				"id":          doc.ID,
				"title":       doc.Title,
				"description": doc.Description,
				"difficulty":  doc.Difficulty,
				"keywords":    doc.Keywords,
			})
		}
	}

	return map[string]interface{}{
		"map_type":     "document_hierarchy",
		"categories":   docMap,
		"total_docs":   len(s.documents),
		"generated_at": time.Now().Format("2006-01-02 15:04:05"),
	}
}

// generateNavigationTree 生成导航树
func (s *CangJieDocServer) generateNavigationTree(category types.DocumentCategory, maxItems int) map[string]interface{} {
	type TreeNode struct {
		Name        string      `json:"name"`
		Type        string      `json:"type"` // category/subcategory/document
		ID          string      `json:"id,omitempty"`
		Description string      `json:"description,omitempty"`
		Children    []TreeNode  `json:"children,omitempty"`
	}

	var roots []TreeNode

	// 构建树结构
	treeMap := make(map[string]*TreeNode)
	categoryMap := make(map[string]*TreeNode)

	for _, doc := range s.documents {
		if category != "" && doc.Category != category {
			continue
		}

		catStr := string(doc.Category)

		// 创建分类节点（如果不存在）
		if _, exists := treeMap[catStr]; !exists {
			catNode := &TreeNode{
				Name:     types.CategoryNames[doc.Category],
				Type:     "category",
				ID:       catStr,
				Children: make([]TreeNode, 0),
			}
			treeMap[catStr] = catNode
			categoryMap[catStr] = catNode
			roots = append(roots, *catNode)
		}

		// 创建子分类节点（如果不存在）
		subcatKey := catStr + "/" + doc.Subcategory
		if _, exists := treeMap[subcatKey]; !exists {
			subcatNode := &TreeNode{
				Name:     doc.Subcategory,
				Type:     "subcategory",
				ID:       subcatKey,
				Children: make([]TreeNode, 0),
			}
			treeMap[subcatKey] = subcatNode
			categoryMap[catStr].Children = append(categoryMap[catStr].Children, *subcatNode)
		}

		// 添加文档节点（限制数量）
		if len(treeMap[subcatKey].Children) < maxItems/10 {
			docNode := TreeNode{
				Name:        doc.Title,
				Type:        "document",
				ID:          doc.ID,
				Description: doc.Description,
			}
			treeMap[subcatKey].Children = append(treeMap[subcatKey].Children, docNode)
		}
	}

	return map[string]interface{}{
		"tree_type":    "navigation",
		"roots":        roots,
		"total_nodes":  len(treeMap),
		"generated_at": time.Now().Format("2006-01-02 15:04:05"),
	}
}

// sortDocuments 排序文档
func (s *CangJieDocServer) sortDocuments(documents []*types.Document, sortBy string) {
	switch sortBy {
	case "title":
		sort.Slice(documents, func(i, j int) bool {
			return documents[i].Title < documents[j].Title
		})
	case "difficulty":
		// 按难度级别排序：beginner < intermediate < advanced
		difficultyOrder := map[string]int{
			"beginner":     1,
			"intermediate": 2,
			"advanced":     3,
		}
		sort.Slice(documents, func(i, j int) bool {
			orderI := difficultyOrder[documents[i].Difficulty]
			orderJ := difficultyOrder[documents[j].Difficulty]
			if orderI != orderJ {
				return orderI < orderJ
			}
			return documents[i].Title < documents[j].Title
		})
	case "last_modified":
		sort.Slice(documents, func(i, j int) bool {
			return documents[i].LastModified.After(documents[j].LastModified)
		})
	}
}

// extractSection 提取文档的特定章节
func (s *CangJieDocServer) extractSection(content, section string) string {
	// 简单的章节提取，支持 # ## ### 等标题格式
	lines := strings.Split(content, "\n")
	var sectionLines []string
	var inSection bool
	sectionPattern := regexp.MustCompile(`^(#{1,6})\s+` + regexp.QuoteMeta(section))

	for _, line := range lines {
		if sectionPattern.MatchString(line) {
			inSection = true
			sectionLines = append(sectionLines, line)
			continue
		}

		if inSection {
			// 检查是否到了下一个同级或更高级标题
			if strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "######") {
				// 检查标题级别
				currentLevel := 0
				for _, char := range line {
					if char == '#' {
						currentLevel++
					} else {
						break
					}
				}

				sectionLevel := 1 // 默认假设目标章节是 #
				for _, char := range section {
					if char == '.' {
						sectionLevel++
					}
				}

				if currentLevel <= sectionLevel {
					break // 到达下一个同级或更高级标题，停止
				}
			}
			sectionLines = append(sectionLines, line)
		}
	}

	if len(sectionLines) == 0 {
		return fmt.Sprintf("未找到章节: %s", section)
	}

	return strings.Join(sectionLines, "\n")
}