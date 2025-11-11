package mcp

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"cangje-docs-mcp/pkg/scanner"
	"cangje-docs-mcp/pkg/search"
	"cangje-docs-mcp/pkg/types"
	"github.com/mark3labs/mcp-go/server"
)

// CangJieDocServer 仓颉文档MCP服务器
type CangJieDocServer struct {
	server      *server.MCPServer
	documents   map[string]*types.Document
	searchEngine *search.SearchEngine
	scanner     *scanner.Scanner
}

// NewCangJieDocServer 创建新的仓颉文档服务器
func NewCangJieDocServer(docRoot string) *CangJieDocServer {
	// 配置slog，不输出到stdio避免干扰MCP通信
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError, // 只输出错误日志到stderr
	})))

	// 创建MCP服务器
	mcpServer := server.NewMCPServer(
		"仓颉语言文档检索系统",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	s := &CangJieDocServer{
		server:       mcpServer,
		documents:    make(map[string]*types.Document),
		searchEngine: search.NewSearchEngine(),
		scanner:      scanner.NewScanner(docRoot),
	}

	// 注册工具
	s.registerTools()

	return s
}

// Serve 启动服务器
func (s *CangJieDocServer) Serve(ctx context.Context) error {
	// 初始化文档
	if err := s.initializeDocuments(); err != nil {
		return fmt.Errorf("failed to initialize documents: %w", err)
	}

	// 构建搜索索引
	s.searchEngine.BuildIndex(s.documents)

	slog.Info("服务器已启动", "文档数量", len(s.documents))

	// 启动MCP服务器（stdio协议）
	return server.ServeStdio(s.server)
}

// initializeDocuments 初始化文档
func (s *CangJieDocServer) initializeDocuments() error {
	docRoot := s.scanner.GetDocRoot()
	slog.Info("开始扫描文档目录", "路径", docRoot)

	// 检查文档目录是否存在
	if _, err := os.Stat(docRoot); os.IsNotExist(err) {
		return fmt.Errorf("文档目录不存在: %s", docRoot)
	}

	// 扫描所有文档
	documents, err := s.scanner.ScanAll()
	if err != nil {
		return fmt.Errorf("failed to scan documents: %w", err)
	}

	s.documents = documents

	slog.Info("文档扫描完成", "文档数量", len(documents))

	// 打印分类统计
	categoryStats := make(map[types.DocumentCategory]int)
	for _, doc := range documents {
		categoryStats[doc.Category]++
	}

	for category, count := range categoryStats {
		slog.Info("分类统计", "分类", types.CategoryNames[category], "数量", count)
	}

	return nil
}