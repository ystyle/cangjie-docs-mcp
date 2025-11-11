package search

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"cangje-docs-mcp/pkg/types"
)

// SearchEngine 搜索引擎
type SearchEngine struct {
	documents    map[string]*types.Document
	keywordIndex map[string][]string // 关键词到文档ID的映射
}

// NewSearchEngine 创建新的搜索引擎
func NewSearchEngine() *SearchEngine {
	return &SearchEngine{
		documents:    make(map[string]*types.Document),
		keywordIndex: make(map[string][]string),
	}
}

// BuildIndex 构建搜索索引
func (se *SearchEngine) BuildIndex(documents map[string]*types.Document) {
	se.documents = documents
	se.buildKeywordIndex()
}

// buildKeywordIndex 构建关键词索引
func (se *SearchEngine) buildKeywordIndex() {
	se.keywordIndex = make(map[string][]string)

	for docID, doc := range se.documents {
		// 为标题建立索引
		titleWords := se.extractWords(doc.Title)
		for _, word := range titleWords {
			se.addToIndex(word, docID)
		}

		// 为描述建立索引
		descriptionWords := se.extractWords(doc.Description)
		for _, word := range descriptionWords {
			se.addToIndex(word, docID)
		}

		// 为关键词建立索引
		for _, keyword := range doc.Keywords {
			se.addToIndex(strings.ToLower(keyword), docID)
		}

		// 为内容建立索引（只取前1000个字符）
		content := doc.Content
		if len(content) > 1000 {
			content = content[:1000]
		}
		contentWords := se.extractWords(content)
		for _, word := range contentWords {
			se.addToIndex(word, docID)
		}

		// 为文件名建立索引
		filename := strings.TrimSuffix(doc.RelativePath, ".md")
		filenameWords := se.extractWords(filename)
		for _, word := range filenameWords {
			se.addToIndex(word, docID)
		}
	}
}

// addToIndex 添加到索引
func (se *SearchEngine) addToIndex(keyword, docID string) {
	if _, exists := se.keywordIndex[keyword]; !exists {
		se.keywordIndex[keyword] = []string{}
	}

	// 避免重复添加
	for _, id := range se.keywordIndex[keyword] {
		if id == docID {
			return
		}
	}

	se.keywordIndex[keyword] = append(se.keywordIndex[keyword], docID)
}

// Search 执行搜索
func (se *SearchEngine) Search(req types.SearchRequest) []types.SearchResult {
	query := strings.ToLower(strings.TrimSpace(req.Query))
	if query == "" {
		return []types.SearchResult{}
	}

	maxResults := req.MaxResults
	if maxResults <= 0 {
		maxResults = types.DefaultMaxResults
	}

	minConfidence := req.MinConfidence
	if minConfidence <= 0 {
		minConfidence = types.DefaultMinConfidence
	}

	// 提取查询词
	queryWords := se.extractWords(query)

	// 收集候选文档
	candidateDocs := make(map[string]*DocumentScore)

	// 1. 精确匹配
	for _, doc := range se.documents {
		if se.matchesQuery(doc, query, req.Category) {
			score := se.calculateScore(doc, query, queryWords, "exact")
			candidateDocs[doc.ID] = &DocumentScore{
				Document: doc,
				Score:    score,
				MatchType: "exact",
			}
		}
	}

	// 2. 关键词索引匹配
	for _, word := range queryWords {
		if docIDs, exists := se.keywordIndex[word]; exists {
			for _, docID := range docIDs {
				if doc, exists := se.documents[docID]; exists && se.matchesCategory(doc, req.Category) {
					if existing, exists := candidateDocs[docID]; exists {
						// 累加分数
						existing.Score += types.TitleMatchWeight
					} else {
						score := se.calculateScore(doc, query, queryWords, "keyword")
						candidateDocs[docID] = &DocumentScore{
							Document: doc,
							Score:    score,
							MatchType: "keyword",
						}
					}
				}
			}
		}
	}

	// 3. 模糊匹配（如果候选文档不够）
	if len(candidateDocs) < maxResults {
		for _, doc := range se.documents {
			if _, exists := candidateDocs[doc.ID]; !exists && se.matchesCategory(doc, req.Category) {
				score := se.calculateFuzzyScore(doc, query, queryWords)
				if score > 0 {
					candidateDocs[doc.ID] = &DocumentScore{
						Document: doc,
						Score:    score,
						MatchType: "fuzzy",
					}
				}
			}
		}
	}

	// 转换为结果列表并排序
	var results []types.SearchResult
	for _, docScore := range candidateDocs {
		if docScore.Score >= minConfidence {
			results = append(results, types.SearchResult{
				Document:  *docScore.Document,
				Score:     docScore.Score,
				MatchType: docScore.MatchType,
				MatchText: se.extractMatchText(docScore.Document, query),
			})
		}
	}

	// 按分数排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// 限制结果数量
	if len(results) > maxResults {
		results = results[:maxResults]
	}

	return results
}

// DocumentScore 文档分数结构
type DocumentScore struct {
	Document  *types.Document
	Score     float64
	MatchType string
}

// matchesQuery 检查文档是否匹配查询
func (se *SearchEngine) matchesQuery(doc *types.Document, query string, category types.DocumentCategory) bool {
	if !se.matchesCategory(doc, category) {
		return false
	}

	// 检查标题
	if strings.Contains(strings.ToLower(doc.Title), query) {
		return true
	}

	// 检查描述
	if strings.Contains(strings.ToLower(doc.Description), query) {
		return true
	}

	// 检查关键词
	for _, keyword := range doc.Keywords {
		if strings.Contains(strings.ToLower(keyword), query) {
			return true
		}
	}

	// 检查文件名
	if strings.Contains(strings.ToLower(doc.RelativePath), query) {
		return true
	}

	return false
}

// matchesCategory 检查文档是否匹配分类
func (se *SearchEngine) matchesCategory(doc *types.Document, category types.DocumentCategory) bool {
	if category == "" {
		return true
	}
	return doc.Category == category
}

// calculateScore 计算文档分数
func (se *SearchEngine) calculateScore(doc *types.Document, query string, queryWords []string, matchType string) float64 {
	var score float64

	lowerTitle := strings.ToLower(doc.Title)
	lowerDesc := strings.ToLower(doc.Description)
	lowerQuery := strings.ToLower(query)

	switch matchType {
	case "exact":
		// 精确匹配
		if strings.Contains(lowerTitle, lowerQuery) {
			score += types.ExactMatchWeight
		}
		if strings.Contains(lowerDesc, lowerQuery) {
			score += types.DescriptionWeight
		}
		for _, keyword := range doc.Keywords {
			if strings.Contains(strings.ToLower(keyword), lowerQuery) {
				score += types.ExactMatchWeight
				break
			}
		}

	case "keyword":
		// 关键词匹配
		for _, word := range queryWords {
			if strings.Contains(lowerTitle, word) {
				score += types.TitleMatchWeight
			}
			if strings.Contains(lowerDesc, word) {
				score += types.DescriptionWeight
			}
		}

	case "fuzzy":
		// 模糊匹配
		score = se.calculateFuzzyScore(doc, query, queryWords)
	}

	// 文件名匹配加分
	if strings.Contains(strings.ToLower(doc.RelativePath), lowerQuery) {
		score += types.FilenameMatchWeight
	}

	return score
}

// calculateFuzzyScore 计算模糊匹配分数
func (se *SearchEngine) calculateFuzzyScore(doc *types.Document, query string, queryWords []string) float64 {
	var score float64
	content := strings.ToLower(doc.Content)

	// 计算查询词在内容中出现的次数
	for _, word := range queryWords {
		count := strings.Count(content, word)
		if count > 0 {
			score += types.ContentMatchWeight * float64(count)
		}
	}

	// 部分匹配加分
	for _, word := range queryWords {
		if strings.Contains(strings.ToLower(doc.Title), word) {
			score += types.TitleMatchWeight * 0.5
		}
		if strings.Contains(strings.ToLower(doc.Description), word) {
			score += types.DescriptionWeight * 0.5
		}
	}

	return score
}

// extractWords 提取单词
func (se *SearchEngine) extractWords(text string) []string {
	// 使用正则表达式提取中文和英文单词
	re := regexp.MustCompile(`[\p{Han}a-zA-Z]+`)
	matches := re.FindAllString(text, -1)

	var words []string
	for _, match := range matches {
		word := strings.ToLower(match)
		if len(word) > 1 && !se.isStopWord(word) {
			words = append(words, word)
		}
	}

	return words
}

// isStopWord 检查是否为停用词
func (se *SearchEngine) isStopWord(word string) bool {
	stopWords := map[string]bool{
		"的": true, "了": true, "在": true, "是": true, "我": true, "有": true,
		"和": true, "就": true, "不": true, "人": true, "都": true, "一": true,
		"一个": true, "上": true, "也": true, "很": true, "到": true, "说": true,
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "as": true, "is": true, "are": true, "was": true,
		"were": true, "be": true, "been": true, "being": true, "have": true, "has": true,
		"had": true, "do": true, "does": true, "did": true, "will": true, "would": true,
		"could": true, "should": true, "may": true, "might": true, "can": true, "this": true,
		"that": true, "these": true, "those": true, "i": true, "you": true, "he": true,
		"she": true, "it": true, "we": true, "they": true, "what": true, "which": true,
		"who": true, "when": true, "where": true, "why": true, "how": true, "all": true,
		"each": true, "every": true, "both": true, "few": true, "more": true, "most": true,
		"other": true, "some": true, "such": true, "only": true, "own": true, "same": true,
		"so": true, "than": true, "too": true, "very": true, "just": true, "now": true,
	}

	return stopWords[word]
}

// extractMatchText 提取匹配文本片段
func (se *SearchEngine) extractMatchText(doc *types.Document, query string) string {
	content := doc.Content
	lowerContent := strings.ToLower(content)
	lowerQuery := strings.ToLower(query)

	// 在内容中查找查询词
	if index := strings.Index(lowerContent, lowerQuery); index != -1 {
		start := index - 50
		if start < 0 {
			start = 0
		}
		end := index + len(query) + 50
		if end > len(content) {
			end = len(content)
		}

		matchText := content[start:end]
		if start > 0 {
			matchText = "..." + matchText
		}
		if end < len(content) {
			matchText = matchText + "..."
		}
		return matchText
	}

	// 如果在内容中没找到，返回描述
	if doc.Description != "" {
		return doc.Description
	}

	// 否则返回标题
	return doc.Title
}

// GetSuggestions 获取建议文档
func (se *SearchEngine) GetSuggestions(req types.SuggestionRequest) []types.Suggestion {
	maxSuggestions := req.MaxSuggestions
	if maxSuggestions <= 0 {
		maxSuggestions = types.DefaultMaxSuggestions
	}

	var suggestions []types.Suggestion

	switch req.SuggestionType {
	case "learning_path":
		suggestions = se.getLearningPathSuggestions(req.Context, maxSuggestions)
	case "related":
		suggestions = se.getRelatedSuggestions(req.Context, maxSuggestions)
	case "prerequisite":
		suggestions = se.getPrerequisiteSuggestions(req.Context, maxSuggestions)
	default:
		suggestions = se.getRelatedSuggestions(req.Context, maxSuggestions)
	}

	return suggestions
}

// getLearningPathSuggestions 获取学习路径建议
func (se *SearchEngine) getLearningPathSuggestions(context string, maxSuggestions int) []types.Suggestion {
	var suggestions []types.Suggestion

	// 根据上下文确定学习阶段
	stage := se.determineLearningStage(context)

	if paths, exists := types.LearningPaths[stage]; exists {
		for i, path := range paths {
			if i >= maxSuggestions {
				break
			}

			// 查找匹配的文档
			for _, doc := range se.documents {
				if strings.Contains(doc.RelativePath, path) {
					suggestions = append(suggestions, types.Suggestion{
						Document:  *doc,
						Reason:    fmt.Sprintf("学习路径 - %s 阶段", stage),
						Relevance: float64(len(paths) - i) / float64(len(paths)),
						Type:      "learning_path",
					})
					break
				}
			}
		}
	}

	return suggestions
}

// getRelatedSuggestions 获取相关建议
func (se *SearchEngine) getRelatedSuggestions(context string, maxSuggestions int) []types.Suggestion {
	var suggestions []types.Suggestion

	// 如果上下文是文档ID，查找该文档
	var targetDoc *types.Document
	if doc, exists := se.documents[context]; exists {
		targetDoc = doc
	} else {
		// 否则搜索相关文档
		searchReq := types.SearchRequest{
			Query:      context,
			MaxResults: 1,
		}
		results := se.Search(searchReq)
		if len(results) > 0 {
			targetDoc = &results[0].Document
		}
	}

	if targetDoc == nil {
		return suggestions
	}

	// 查找同分类的文档
	for _, doc := range se.documents {
		if doc.ID == targetDoc.ID {
			continue
		}

		var relevance float64
		var reason string

		// 同分类
		if doc.Category == targetDoc.Category {
			relevance += 0.5
			reason = "同分类文档"
		}

		// 同子分类
		if doc.Subcategory == targetDoc.Subcategory && doc.Subcategory != "" {
			relevance += 0.3
			reason += " - 同子分类"
		}

		// 关键词重叠
		for _, kw1 := range doc.Keywords {
			for _, kw2 := range targetDoc.Keywords {
				if kw1 == kw2 {
					relevance += 0.2
					break
				}
			}
		}

		if relevance > 0 {
			suggestions = append(suggestions, types.Suggestion{
				Document:  *doc,
				Reason:    reason,
				Relevance: relevance,
				Type:      "related",
			})
		}
	}

	// 按相关性排序并限制数量
	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].Relevance > suggestions[j].Relevance
	})

	if len(suggestions) > maxSuggestions {
		suggestions = suggestions[:maxSuggestions]
	}

	return suggestions
}

// getPrerequisiteSuggestions 获取前置知识建议
func (se *SearchEngine) getPrerequisiteSuggestions(context string, maxSuggestions int) []types.Suggestion {
	var suggestions []types.Suggestion

	// 根据当前文档推荐基础内容
	if doc, exists := se.documents[context]; exists {
		// 如果当前文档不是基础文档，推荐基础文档
		if doc.Category != types.CategoryManual || doc.Subcategory != "first_understanding" {
			for _, candidate := range se.documents {
				if candidate.Category == types.CategoryManual &&
					(candidate.Subcategory == "first_understanding" || candidate.Subcategory == "basic_data_type") {
					suggestions = append(suggestions, types.Suggestion{
						Document:  *candidate,
						Reason:    "前置基础知识",
						Relevance: 0.8,
						Type:      "prerequisite",
					})
				}
			}
		}
	} else {
		// 根据查询内容推荐基础文档
		searchReq := types.SearchRequest{
			Query:      context,
			Category:   types.CategoryManual,
			MaxResults: maxSuggestions,
		}
		results := se.Search(searchReq)
		for i, result := range results {
			suggestions = append(suggestions, types.Suggestion{
				Document:  result.Document,
				Reason:    "相关基础知识",
				Relevance: result.Score / 10.0, // 归一化分数
				Type:      "prerequisite",
			})
			if i >= maxSuggestions-1 {
				break
			}
		}
	}

	return suggestions
}

// determineLearningStage 确定学习阶段
func (se *SearchEngine) determineLearningStage(context string) string {
	lowerContext := strings.ToLower(context)

	if strings.Contains(lowerContext, "入门") || strings.Contains(lowerContext, "基础") ||
		strings.Contains(lowerContext, "beginner") || strings.Contains(lowerContext, "basic") {
		return "beginner"
	}

	if strings.Contains(lowerContext, "高级") || strings.Contains(lowerContext, "进阶") ||
		strings.Contains(lowerContext, "advanced") {
		return "advanced"
	}

	return "intermediate"
}