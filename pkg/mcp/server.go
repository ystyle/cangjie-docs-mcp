package mcp

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"cangje-docs-mcp/pkg/scanner"
	"cangje-docs-mcp/pkg/search"
	"cangje-docs-mcp/pkg/types"
	"cangje-docs-mcp/pkg/utils"
	"github.com/mark3labs/mcp-go/server"
)

// CangJieDocServer 仓颉文档MCP服务器
type CangJieDocServer struct {
	server         *server.MCPServer
	docRoot        string
	defaultVersion string
	documents      map[string]*types.Document            // default 版本文档
	versionedDocs  map[string]map[string]*types.Document // version → docs
	searchEngine   *search.SearchEngine
	searchEngines  map[string]*search.SearchEngine // version → search engine
	scanner        *scanner.Scanner
	// versionMu 保护 versionedDocs / searchEngines。
	// MCP 工具调用由 worker 池并发执行（默认 5 个），无锁会导致 concurrent map writes 直接崩溃。
	versionMu sync.RWMutex
	// loadMu 串行化版本文档的首次加载，避免同一版本被并发 git worktree add 到同一路径。
	loadMu sync.Mutex
}

// NewCangJieDocServer 创建新的仓颉文档服务器
func NewCangJieDocServer(docRoot string) *CangJieDocServer {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError,
	})))
	mcpServer := server.NewMCPServer(
		"仓颉语言文档检索系统",
		types.ServerVersion,
		server.WithToolCapabilities(true),
	)
	s := &CangJieDocServer{
		server:        mcpServer,
		docRoot:       docRoot,
		documents:     make(map[string]*types.Document),
		versionedDocs: make(map[string]map[string]*types.Document),
		searchEngine:  search.NewSearchEngine(),
		searchEngines: make(map[string]*search.SearchEngine),
		scanner:       scanner.NewScanner(docRoot),
	}
	s.registerTools()
	return s
}

// Serve 启动服务器
func (s *CangJieDocServer) Serve(ctx context.Context) error {
	if err := s.initializeDocuments(); err != nil {
		return fmt.Errorf("failed to initialize documents: %w", err)
	}
	return server.ServeStdio(s.server)
}

// initializeDocuments 初始化文档
func (s *CangJieDocServer) initializeDocuments() error {
	docRoot := s.scanner.GetDocRoot()
	if _, err := os.Stat(docRoot); os.IsNotExist(err) {
		return fmt.Errorf("文档目录不存在: %s", docRoot)
	}
	if defVer, err := utils.GetRemoteHEAD(docRoot); err == nil {
		s.defaultVersion = defVer
		slog.Info("检测到默认版本", "version", s.defaultVersion)
	}
	documents, err := s.scanner.ScanAll()
	if err != nil {
		return fmt.Errorf("扫描文档失败: %w", err)
	}
	s.documents = documents
	s.searchEngine.BuildIndex(s.documents)
	slog.Info("服务器已启动", "文档数量", len(documents), "默认版本", s.defaultVersion)
	return nil
}

// getDocDirForVersion 获取指定版本的文档目录
func (s *CangJieDocServer) getDocDirForVersion(version string) string {
	return utils.GetVersionDocDir(s.docRoot, version)
}

// getDocsForVersion 获取指定版本的文档 map
func (s *CangJieDocServer) getDocsForVersion(version string) map[string]*types.Document {
	if version == "" || version == s.defaultVersion {
		return s.documents
	}
	s.versionMu.RLock()
	defer s.versionMu.RUnlock()
	if docs, ok := s.versionedDocs[version]; ok {
		return docs
	}
	return nil
}

// isVersionLoaded 判断某版本的文档是否已加载到内存。
//
// 这是"就绪"的唯一判据：必须与 getDocsForVersion 的可见性一致。
// 早期实现用 os.Stat(worktree 目录) 判断，worktree 已建好但本进程未加载时会误报未就绪。
func (s *CangJieDocServer) isVersionLoaded(version string) bool {
	if version == "" || version == s.defaultVersion {
		return true
	}
	s.versionMu.RLock()
	defer s.versionMu.RUnlock()
	_, ok := s.versionedDocs[version]
	return ok
}

// ensureVersionDocs 确保指定版本的文档已加载
func (s *CangJieDocServer) ensureVersionDocs(version string) (map[string]*types.Document, error) {
	if version == "" || version == s.defaultVersion {
		return s.documents, nil
	}
	if docs := s.getDocsForVersion(version); docs != nil {
		return docs, nil
	}
	// 串行化加载，避免并发对同一版本重复 git worktree add / 扫描
	s.loadMu.Lock()
	defer s.loadMu.Unlock()
	// 加锁后二次检查，可能已被其他 goroutine 加载完毕
	if docs := s.getDocsForVersion(version); docs != nil {
		return docs, nil
	}
	versionDocDir, err := utils.EnsureWorktree(s.docRoot, version)
	if err != nil {
		return nil, fmt.Errorf("准备版本 %s 的文档目录失败: %w", version, err)
	}
	verScanner := scanner.NewVersionedScanner(versionDocDir, version)
	docs, err := verScanner.ScanAll()
	if err != nil {
		return nil, fmt.Errorf("扫描版本 %s 文档失败: %w", version, err)
	}
	se := search.NewSearchEngine()
	se.BuildIndex(docs)
	s.versionMu.Lock()
	s.versionedDocs[version] = docs
	s.searchEngines[version] = se
	s.versionMu.Unlock()
	slog.Info("版本文档已加载", "version", version, "count", len(docs))
	return docs, nil
}

// getSearchEngineForVersion 获取指定版本的搜索引擎
func (s *CangJieDocServer) getSearchEngineForVersion(version string) *search.SearchEngine {
	if version == "" || version == s.defaultVersion {
		return s.searchEngine
	}
	s.versionMu.RLock()
	defer s.versionMu.RUnlock()
	if se, ok := s.searchEngines[version]; ok {
		return se
	}
	return nil
}
