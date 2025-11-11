package types

import (
	"os"
	"path/filepath"
)

// 默认文档根目录 - 可执行文件所在目录的CangjieCorpus
var DefaultDocumentRootPath = func() string {
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		return filepath.Join(exeDir, "CangjieCorpus")
	}
	return "./CangjieCorpus" // fallback
}()

// 搜索权重配置
const (
	ExactMatchWeight   = 10.0 // 精确关键词匹配
	TitleMatchWeight    = 8.0  // 标题匹配
	DescriptionWeight   = 6.0  // 描述匹配
	ContentMatchWeight  = 3.0  // 内容匹配
	FilenameMatchWeight = 5.0  // 文件名匹配
)

// 默认配置
const (
	DefaultMaxResults   = 10
	DefaultMinConfidence = 0.3
	DefaultMaxSuggestions = 5
)

// 分类映射
var CategoryNames = map[DocumentCategory]string{
	CategoryManual: "基础手册",
	CategoryLibs:   "标准库API",
	CategoryTools:  "开发工具",
	CategoryExtra:  "额外内容",
}

// 分类描述
var CategoryDescriptions = map[DocumentCategory]string{
	CategoryManual: "仓颉语言基础教程和编程概念",
	CategoryLibs:   "仓颉标准库API文档",
	CategoryTools:  "仓颉开发工具文档",
	CategoryExtra:  "高级主题和最佳实践",
}

// 学习路径
var LearningPaths = map[string][]string{
	"beginner": {
		"manual/first_understanding",
		"manual/basic_data_type",
		"manual/basic_programming_concepts",
		"manual/function",
	},
	"intermediate": {
		"manual/class_and_interface",
		"manual/collections",
		"libs/std",
	},
	"advanced": {
		"manual/concurrency",
		"manual/compile_and_build",
		"extra",
	},
}