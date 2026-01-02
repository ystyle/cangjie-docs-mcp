package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
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
		mcp.WithNumber("level",
			mcp.Description("树形显示深度 (仅navigation/tree视图，默认3，0表示全部)"),
		),
	)
	s.server.AddTool(overviewTool, s.handleGetDocumentOverview)

	// 文档列表工具
	listTool := mcp.NewTool("list_documents",
		mcp.WithDescription("列出分类或子分类的文档（支持路径导航，类似 ls 命令）"),
		mcp.WithString("category",
			mcp.Required(),
			mcp.Description("主分类"),
			mcp.Enum("manual", "libs", "tools", "extra", "ohos"),
		),
		mcp.WithString("subcategory",
			mcp.Description("子分类路径（支持多级路径，如 'stdx' 或 'stdx/crypto'），留空显示子分类列表"),
		),
		mcp.WithString("sort_by",
			mcp.Description("排序方式 (默认title)"),
			mcp.Enum("title", "difficulty", "last_modified"),
		),
		mcp.WithBoolean("include_preview",
			mcp.Description("是否包含内容预览 (默认false)"),
		),
		mcp.WithNumber("max_items",
			mcp.Description("最大返回数量 (默认100)"),
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

	level := 3 // 默认显示3层
	if l, ok := request.GetArguments()["level"].(float64); ok {
		level = int(l)
	}

	// 根据视图类型生成不同的响应
	switch viewType {
	case "map":
		// 生成文档地图
		response := s.generateDocumentMap(category, maxItems)
		data, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to marshal response: %v", err)), nil
		}
		return mcp.NewToolResultText(string(data)), nil
	case "navigation", "tree":
		// 生成导航树（文本格式）
		treeText := s.generateNavigationTreeText(category, maxItems, level)
		return mcp.NewToolResultText(treeText), nil
	default: // overview
		// 生成总览
		response := s.generateOverview(category, maxItems)
		data, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to marshal response: %v", err)), nil
		}
		return mcp.NewToolResultText(string(data)), nil
	}
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

	maxItems := 0
	if mi, ok := request.GetArguments()["max_items"].(float64); ok {
		maxItems = int(mi)
	}

	// 解析路径
	pathParts := []string{}
	if subcategory != "" {
		pathParts = strings.Split(subcategory, "/")
	}

	var builder strings.Builder

	// 根据路径深度显示不同内容
	if len(pathParts) == 0 {
		// 深度0：显示子分类列表
		return s.listSubcategories(category, builder)
	} else if len(pathParts) == 1 {
		// 深度1：显示该子分类下的一级目录
		return s.listDirectories(category, pathParts[0], builder)
	} else {
		// 深度2+：显示文档列表
		return s.listDocumentsAtPath(category, subcategory, pathParts, sortBy, includePreview, maxItems, builder)
	}
}

// listSubcategories 列出子分类
func (s *CangJieDocServer) listSubcategories(category string, builder strings.Builder) (*mcp.CallToolResult, error) {
	// 统计每个子分类的文档数
	subcatCounts := make(map[string]int)
	for _, doc := range s.documents {
		if string(doc.Category) == category && len(doc.Prerequisites) == 0 {
			subcatCounts[doc.Subcategory]++
		}
	}

	// 排序子分类
	var subcats []string
	for subcat := range subcatCounts {
		subcats = append(subcats, subcat)
	}
	sort.Strings(subcats)

	// 标题
	builder.WriteString(fmt.Sprintf("📋 %s\n\n", types.CategoryNames[types.DocumentCategory(category)]))
	builder.WriteString("| 子分类 | 文档数 |\n")
	builder.WriteString("|---|---|\n")

	for _, subcat := range subcats {
		builder.WriteString(fmt.Sprintf("| %s | %d |\n", subcat, subcatCounts[subcat]))
	}

	totalDocs := len(subcatCounts)
	builder.WriteString(fmt.Sprintf("\n📊 共 %d 个子分类 | 总计 %d 个原始文档\n",
		totalDocs, countTotalDocs(s, category, "")))

	return mcp.NewToolResultText(builder.String()), nil
}

// listDirectories 列出子分类下的一级目录
func (s *CangJieDocServer) listDirectories(category, subcategory string, builder strings.Builder) (*mcp.CallToolResult, error) {
	// 统计目录下的文档数
	dirCounts := make(map[string]int)
	dirPathMap := make(map[string]string) // 目录名 -> 完整路径前缀

	for _, doc := range s.documents {
		if string(doc.Category) == category && doc.Subcategory == subcategory && len(doc.Prerequisites) == 0 {
			// 解析路径，获取第一级目录
			pathParts := strings.Split(doc.RelativePath, string(filepath.Separator))
			if len(pathParts) > 2 {
				// 跳过子分类本身，获取下一级目录
				dirName := pathParts[2] // 例如 libs/stdx/crypto -> crypto
				dirCounts[dirName]++
				dirPathMap[dirName] = dirName
			}
		}
	}

	// 排序目录
	var dirs []string
	for dir := range dirCounts {
		dirs = append(dirs, dir)
	}
	sort.Strings(dirs)

	// 标题
	builder.WriteString(fmt.Sprintf("📋 %s / %s\n\n", types.CategoryNames[types.DocumentCategory(category)], subcategory))
	builder.WriteString("| 目录 | 文档数 |\n")
	builder.WriteString("|---|---|\n")

	for _, dir := range dirs {
		builder.WriteString(fmt.Sprintf("| %s | %d |\n", dir, dirCounts[dir]))
	}

	totalDirs := len(dirs)
	builder.WriteString(fmt.Sprintf("\n📊 共 %d 个目录 | 使用 '%s/%s/目录名' 深入查看\n",
		totalDirs, category, subcategory))

	return mcp.NewToolResultText(builder.String()), nil
}

// listDocumentsAtPath 列出指定路径下的文档
func (s *CangJieDocServer) listDocumentsAtPath(category, subcategory string, pathParts []string,
	sortBy string, includePreview bool, maxItems int, builder strings.Builder) (*mcp.CallToolResult, error) {

	// 筛选文档
	var documents []*types.Document
	for _, doc := range s.documents {
		if string(doc.Category) == category && len(doc.Prerequisites) == 0 {
			// 首先检查子分类是否匹配
			if len(pathParts) > 0 && doc.Subcategory != pathParts[0] {
				continue
			}

			// 检查路径前缀是否匹配
			docPathParts := strings.Split(doc.RelativePath, string(filepath.Separator))
			if len(docPathParts) >= len(pathParts)+2 {
				// 检查路径是否匹配（跳过子分类部分）
				match := true
				for i, part := range pathParts {
					// docPathParts: [libs, stdx, crypto, xxx.md]
					// pathParts: [stdx, crypto]
					// 需要检查 docPathParts[i+1] == pathParts[i]
					if i+1 >= len(docPathParts) || docPathParts[i+1] != part {
						match = false
						break
					}
				}
				if match {
					documents = append(documents, doc)
				}
			}
		}
	}

	// 排序
	s.sortDocuments(documents, sortBy)

	// 限制结果数量
	maxDocs := 100
	if maxItems > 0 {
		maxDocs = maxItems
	}
	if len(documents) > maxDocs {
		documents = documents[:maxDocs]
	}

	// 标题
	title := fmt.Sprintf("📋 %s", types.CategoryNames[types.DocumentCategory(category)])
	title += fmt.Sprintf(" / %s", subcategory)
	title += fmt.Sprintf(" (%d docs)", len(documents))
	builder.WriteString(title + "\n\n")

	// 表头
	if includePreview {
		builder.WriteString("| ID | 标题 | 难度 | 描述 | 预览 |\n")
		builder.WriteString("|---|---|---|---|---|\n")
	} else {
		builder.WriteString("| ID | 标题 | 难度 | 描述 |\n")
		builder.WriteString("|---|---|---|---|\n")
	}

	// 表格内容
	for _, doc := range documents {
		// 截断描述
		description := doc.Description
		if len(description) > 50 {
			description = description[:47] + "..."
		}

		id := doc.ID
		title := doc.Title
		difficulty := doc.Difficulty

		if includePreview {
			// 包含内容预览
			preview := doc.Content
			if len(preview) > 80 {
				preview = preview[:77] + "..."
			}
			// 转义管道符
			preview = strings.ReplaceAll(preview, "|", "\\|")
			builder.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
				id, title, difficulty, description, preview))
		} else {
			builder.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
				id, title, difficulty, description))
		}
	}

	// 添加统计信息
	builder.WriteString(fmt.Sprintf("\n📊 排序方式: %s | 显示: %d/%d\n",
		sortBy, len(documents), maxDocs))

	return mcp.NewToolResultText(builder.String()), nil
}

// countTotalDocs 统计总文档数
func countTotalDocs(server *CangJieDocServer, category, subcategory string) int {
	count := 0
	for _, doc := range server.documents {
		if string(doc.Category) == category {
			if subcategory == "" || doc.Subcategory == subcategory {
				if len(doc.Prerequisites) == 0 {
					count++
				}
			}
		}
	}
	return count
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

	// 查找文档（支持通过 ID 或 FullPathID 查找）
	doc, exists := s.documents[docID]
	if !exists {
		// 如果通过 ID 找不到，尝试通过 FullPathID 查找
		for _, d := range s.documents {
			if d.FullPathID == docID {
				doc = d
				exists = true
				break
			}
		}
	}

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
		Count       int         `json:"count,omitempty"` // 子节点数量
		Children    []TreeNode  `json:"children,omitempty"`
	}

	var roots []TreeNode

	// 构建树结构
	treeMap := make(map[string]*TreeNode)
	subcatDocCounts := make(map[string]int) // 统计每个子分类的实际文档数

	// 第一遍：统计每个子分类的文档数量
	for _, doc := range s.documents {
		if category != "" && doc.Category != category {
			continue
		}

		catStr := string(doc.Category)
		subcatKey := catStr + "/" + doc.Subcategory
		subcatDocCounts[subcatKey]++
	}

	// 第二遍：构建树结构（只包含原始文档，不包含分割后的子文档）
	for _, doc := range s.documents {
		if category != "" && doc.Category != category {
			continue
		}

		// 跳过分割后的文档：通过Prerequisites字段判断
		// 分割后的文档的Prerequisites包含父文档ID
		if len(doc.Prerequisites) > 0 {
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
		}

		// 创建子分类节点（如果不存在）
		subcatKey := catStr + "/" + doc.Subcategory
		if _, exists := treeMap[subcatKey]; !exists {
			subcatNode := &TreeNode{
				Name:     doc.Subcategory,
				Type:     "subcategory",
				ID:       subcatKey,
				Count:    subcatDocCounts[subcatKey], // 实际文档总数
				Children: make([]TreeNode, 0),
			}
			treeMap[subcatKey] = subcatNode
			treeMap[catStr].Children = append(treeMap[catStr].Children, *subcatNode)
		}

		// 添加文档节点（按目录结构组织）
		// 使用RelativePath作为树结构
		pathParts := strings.Split(doc.RelativePath, string(filepath.Separator))
		if len(pathParts) > 2 {
			// 例如: libs/std/core/core_package_api/core_package_structs.md
			// 构建: std → core → core_package_api → core_package_structs.md
			currentLevel := treeMap[catStr]

			// 遍历路径中的目录（除了最后一层的文件名）
			for i := 2; i < len(pathParts)-1; i++ {
				dirName := pathParts[i]
				dirKey := strings.Join(pathParts[:i+1], "/")

				// 查找或创建目录节点
				var dirNode *TreeNode
				found := false
				for j, child := range currentLevel.Children {
					if child.Name == dirName && child.Type == "directory" {
						dirNode = &currentLevel.Children[j]
						found = true
						break
					}
				}

				if !found {
					dirNode = &TreeNode{
						Name:     dirName,
						Type:     "directory",
						ID:       dirKey,
						Children: make([]TreeNode, 0),
					}
					currentLevel.Children = append(currentLevel.Children, *dirNode)
				}

				currentLevel = dirNode
			}

			// 添加文档节点
			docNode := TreeNode{
				Name:        doc.Title,
				Type:        "document",
				ID:          doc.ID,
				Description: doc.Description,
			}
			currentLevel.Children = append(currentLevel.Children, docNode)
		}
	}

	// 从treeMap重新构建roots（递归复制，确保包含所有children）
	roots = make([]TreeNode, 0)
	for _, catStr := range []string{string(category)} {
		if catNode, exists := treeMap[catStr]; exists {
			// 递归复制节点及其children
			rootCopy := *catNode
			rootCopy.Children = make([]TreeNode, len(catNode.Children))
			copy(rootCopy.Children, catNode.Children)

			// 递归复制每个子分类的children
			for i, subcat := range rootCopy.Children {
				subcatKey := catStr + "/" + subcat.Name
				if subcatNode, exists := treeMap[subcatKey]; exists {
					subcatCopy := *subcatNode
					subcatCopy.Children = make([]TreeNode, len(subcatNode.Children))
					copy(subcatCopy.Children, subcatNode.Children)
					rootCopy.Children[i] = subcatCopy
				}
			}

			roots = append(roots, rootCopy)
		}
	}

	// 计算实际显示的节点数
	totalNodes := 0
	for _, node := range treeMap {
		totalNodes++
		totalNodes += len(node.Children)
	}

	return map[string]interface{}{
		"tree_type":    "navigation",
		"roots":        roots,
		"total_nodes":  totalNodes,
		"total_docs":   len(s.documents),
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

// generateNavigationTreeText 生成导航树的文本格式（节省 tokens）
func (s *CangJieDocServer) generateNavigationTreeText(category types.DocumentCategory, maxItems int, level int) string {
	type TreeNode struct {
		Name        string
		Type        string
		ID          string
		Description string
		Count       int
		Children    []*TreeNode
	}

	// 使用指针的树结构
	treeMap := make(map[string]*TreeNode)

	// 辅助函数：创建或获取节点
	getOrCreateNode := func(key string, name string, nodeType string) *TreeNode {
		if node, exists := treeMap[key]; exists {
			return node
		}
		newNode := &TreeNode{
			Name:     name,
			Type:     nodeType,
			ID:       key,
			Children: make([]*TreeNode, 0),
		}
		treeMap[key] = newNode
		return newNode
	}

	// 统计子分类文档数量
	subcatDocCounts := make(map[string]int)
	for _, doc := range s.documents {
		if category != "" && doc.Category != category {
			continue
		}
		if len(doc.Prerequisites) > 0 {
			continue
		}
		catStr := string(doc.Category)
		subcatKey := catStr + "/" + doc.Subcategory
		subcatDocCounts[subcatKey]++
	}

	// 遍历文档构建树
	for _, doc := range s.documents {
		if category != "" && doc.Category != category {
			continue
		}
		if len(doc.Prerequisites) > 0 {
			continue
		}

		catStr := string(doc.Category)
		pathParts := strings.Split(doc.RelativePath, string(filepath.Separator))

		if len(pathParts) <= 2 {
			continue
		}

		// 创建或获取分类节点
		catKey := catStr
		catNode := getOrCreateNode(catKey, types.CategoryNames[doc.Category], "category")

		// 创建或获取子分类节点
		subcatKey := catStr + "/" + doc.Subcategory
		subcatNode := getOrCreateNode(subcatKey, doc.Subcategory, "subcategory")
		if subcatNode.Count == 0 {
			subcatNode.Count = subcatDocCounts[subcatKey]
		}

		// 确保子分类是分类的子节点
		found := false
		for _, child := range catNode.Children {
			if child == subcatNode {
				found = true
				break
			}
		}
		if !found {
			catNode.Children = append(catNode.Children, subcatNode)
		}

		// 构建目录路径（从子分类开始）
		currentNode := subcatNode
		for i := 2; i < len(pathParts)-1; i++ {
			dirKey := strings.Join(pathParts[:i+1], "/")
			dirName := pathParts[i]
			dirNode := getOrCreateNode(dirKey, dirName, "directory")

			// 确保目录是当前节点的子节点
			found = false
			for _, child := range currentNode.Children {
				if child == dirNode {
					found = true
					break
				}
			}
			if !found {
				currentNode.Children = append(currentNode.Children, dirNode)
			}

			currentNode = dirNode
		}

		// 添加文档节点
		docNode := &TreeNode{
			Name:        doc.Title,
			Type:        "document",
			ID:          doc.ID,
			Description: doc.Description,
		}
		currentNode.Children = append(currentNode.Children, docNode)
	}

	// 生成树形文本
	var builder strings.Builder

	totalDocs := 0
	for _, doc := range s.documents {
		if category == "" || doc.Category == category {
			totalDocs++
		}
	}

	builder.WriteString(fmt.Sprintf("📚 %s (%d docs)\n\n", types.CategoryNames[category], totalDocs))

	// 递归生成树形文本
	var printTree func([]*TreeNode, string, int)
	printTree = func(nodes []*TreeNode, prefix string, currentDepth int) {
		if level > 0 && currentDepth > level {
			return
		}

		for i, node := range nodes {
			isLast := i == len(nodes)-1
			var connector string
			if isLast {
				connector = "└── "
			} else {
				connector = "├── "
			}

			var nodeStr string
			switch node.Type {
			case "subcategory":
				if node.Count > 0 {
					nodeStr = fmt.Sprintf("%s (%d docs)", node.Name, node.Count)
				} else {
					nodeStr = node.Name
				}
			case "document":
				if node.Description != "" {
					desc := node.Description
					if len(desc) > 60 {
						desc = desc[:57] + "..."
					}
					nodeStr = fmt.Sprintf("%s - %s", node.Name, desc)
				} else {
					nodeStr = node.Name
				}
			default:
				nodeStr = node.Name
			}

			builder.WriteString(prefix + connector + nodeStr + "\n")

			if len(node.Children) > 0 {
				var newPrefix string
				if isLast {
					newPrefix = prefix + "    "
				} else {
					newPrefix = prefix + "│   "
				}
				printTree(node.Children, newPrefix, currentDepth+1)
			}
		}
	}

	// 从分类节点开始输出
	catKey := string(category)
	if catNode, exists := treeMap[catKey]; exists {
		printTree(catNode.Children, "", 1)
	}

	return builder.String()
}