package types

import "time"

// DocumentCategory 文档分类枚举
type DocumentCategory string

const (
	CategoryManual DocumentCategory = "manual" // 基础手册
	CategoryLibs   DocumentCategory = "libs"   // 标准库API
	CategoryTools  DocumentCategory = "tools"  // 开发工具
	CategoryExtra  DocumentCategory = "extra"  // 额外内容
	CategoryOhos   DocumentCategory = "ohos"   // OpenHarmony
)

// Document 文档结构
type Document struct {
	ID            string           `json:"id"`
	Title         string           `json:"title"`
	Category      DocumentCategory `json:"category"`
	Subcategory   string           `json:"subcategory,omitempty"`
	Description   string           `json:"description"`
	FilePath      string           `json:"file_path"`
	RelativePath  string           `json:"relative_path"`
	Keywords      []string         `json:"keywords"`
	Prerequisites []string         `json:"prerequisites"`
	RelatedDocs   []string         `json:"related_docs"`
	Difficulty    string           `json:"difficulty"`
	FileSize      int64            `json:"file_size"`
	LastModified  time.Time        `json:"last_modified"`
	Content       string           `json:"content,omitempty"`
	ContentPreview string          `json:"content_preview,omitempty"`
}

// CategoryInfo 分类信息
type CategoryInfo struct {
	Name        string                     `json:"name"`
	Description string                     `json:"description"`
	Count       int                        `json:"count"`
	Subcategories map[string]SubcategoryInfo `json:"subcategories,omitempty"`
}

// SubcategoryInfo 子分类信息
type SubcategoryInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Count       int      `json:"count"`
	Documents   []DocumentSummary `json:"documents,omitempty"`
}

// DocumentSummary 文档摘要
type DocumentSummary struct {
	ID          string           `json:"id"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Difficulty  string           `json:"difficulty"`
	Subcategory string           `json:"subcategory,omitempty"`
	Keywords    []string         `json:"keywords"`
}

// SearchResult 搜索结果
type SearchResult struct {
	Document   Document `json:"document"`
	Score      float64  `json:"score"`
	MatchType  string   `json:"match_type"` // exact, title, description, content
	MatchText  string   `json:"match_text"` // 匹配的文本片段
}

// SearchRequest 搜索请求
type SearchRequest struct {
	Query        string           `json:"query"`
	Category     DocumentCategory `json:"category,omitempty"`
	MaxResults   int              `json:"max_results,omitempty"`
	MinConfidence float64         `json:"min_confidence,omitempty"`
}

// SuggestionRequest 建议请求
type SuggestionRequest struct {
	Context         string `json:"context"`
	SuggestionType  string `json:"suggestion_type"` // learning_path, related, prerequisite
	MaxSuggestions  int    `json:"max_suggestions,omitempty"`
}

// Suggestion 建议结果
type Suggestion struct {
	Document    Document `json:"document"`
	Reason      string   `json:"reason"`
	Relevance   float64  `json:"relevance"`
	Type        string   `json:"type"`
}

// NavigationNode 导航节点
type NavigationNode struct {
	Name        string           `json:"name"`
	Path        string           `json:"path"`
	Type        string           `json:"type"` // category, subcategory, document
	Count       int              `json:"count,omitempty"`
	Children    []NavigationNode `json:"children,omitempty"`
	Document    *DocumentSummary `json:"document,omitempty"`
}

// DocumentMap 文档映射结构
type DocumentMap struct {
	Categories map[DocumentCategory]CategoryInfo `json:"categories"`
	TotalDocs  int                               `json:"total_docs"`
}