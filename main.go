package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"cangje-docs-mcp/pkg/mcp"
	"cangje-docs-mcp/pkg/utils"
)

func main() {
	// 定义命令行参数
	var docRoot = flag.String("dir", "", "仓颉文档根目录路径 (留空则使用默认位置)")
	var noUpdate = flag.Bool("no-update", false, "禁用自动更新文档")
	var showVersion = flag.Bool("version", false, "显示版本信息")
	var showHelp = flag.Bool("help", false, "显示帮助信息")

	flag.Parse()

	// 显示版本信息
	if *showVersion {
		fmt.Println("仓颉语言文档检索系统 v1.0.0")
		fmt.Println("基于MCP协议的本地文档检索服务器")
		fmt.Println()

		// 获取文档目录
		docDir, _ := utils.GetDocumentDir(*docRoot)
		fmt.Printf("文档目录: %s\n", docDir)

		// 尝试读取文档版本信息
		if docVersion := getDocumentVersion(docDir); docVersion != "" {
			fmt.Printf("文档版本: %s\n", docVersion)
		}
		return
	}

	// 显示帮助信息
	if *showHelp {
		fmt.Println("仓颉语言文档检索系统")
		fmt.Println()
		fmt.Println("用法:")
		fmt.Println("  cangje-docs-mcp [选项]")
		fmt.Println()
		fmt.Println("选项:")
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("说明:")
		fmt.Println("  如果不指定 -dir 参数，系统将自动下载仓颉文档到默认位置：")
		fmt.Println("    - Windows: 可执行文件同目录下的 CangjieCorpus")
		fmt.Println("    - 其他系统: ~/.config/cangje-docs-mcp/CangjieCorpus")
		fmt.Println()
		fmt.Println("  启动时会自动更新文档（除非使用 -no-update 参数）")
		fmt.Println()
		fmt.Println("示例:")
		fmt.Println("  cangje-docs-mcp                                    # 使用默认目录并自动更新")
		fmt.Println("  cangje-docs-mcp -no-update                         # 使用默认目录但不更新")
		fmt.Println("  cangje-docs-mcp -dir /path/to/docs                # 指定文档目录")
		return
	}

	// 获取文档目录
	docDir, err := utils.GetDocumentDir(*docRoot)
	if err != nil {
		log.Fatalf("获取文档目录失败: %v", err)
	}

	// 确保文档存在并更新
	autoUpdate := !*noUpdate
	if err := utils.EnsureDocuments(docDir, autoUpdate); err != nil {
		log.Fatalf("初始化文档失败: %v", err)
	}

	ctx := context.Background()

	server := mcp.NewCangJieDocServer(docDir)

	if err := server.Serve(ctx); err != nil {
		log.Printf("服务器错误: %v", err)
		os.Exit(1)
	}
}

// getDocumentVersion 从文档目录的README.md文件中提取版本信息
func getDocumentVersion(docRoot string) string {
	readmePath := filepath.Join(docRoot, "README.md")

	// 检查文件是否存在
	if _, err := os.Stat(readmePath); err != nil {
		return ""
	}

	// 读取文件内容
	file, err := os.Open(readmePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	// 扫描文件内容寻找版本信息
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() && lineNum < 30 { // 只检查前30行
		line := strings.TrimSpace(scanner.Text())
		lineNum++

		// 匹配仓颉版本信息：仓颉编程语言 v1.0.0（对应官网文档发布日期：2025-07-01）
		if strings.Contains(line, "仓颉编程语言") {
			// 提取版本号
			versionPattern := regexp.MustCompile(`仓颉编程语言\s+([vV]?\d+(?:\.\d+)*)`)
			matches := versionPattern.FindStringSubmatch(line)
			if len(matches) > 1 {
				version := strings.TrimSpace(matches[1])
				// 提取日期信息
				datePattern := regexp.MustCompile(`(\d{4}-\d{2}-\d{2})`)
				dateMatches := datePattern.FindStringSubmatch(line)
				if len(dateMatches) > 1 {
					return fmt.Sprintf("%s (发布日期: %s)", version, dateMatches[1])
				}
				return version
			}
		}

		// 通用版本匹配
		versionPatterns := []string{
			`版本\s*[:：]?\s*([vV]?\d+(?:\.\d+)*)`,
			`Version\s*[:：]?\s*([vV]?\d+(?:\.\d+)*)`,
			`([vV]?\d+(?:\.\d+)*)`,
		}

		for _, pattern := range versionPatterns {
			re := regexp.MustCompile(pattern)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				version := strings.TrimSpace(matches[1])
				if version != "" {
					return version
				}
			}
		}
	}

	return ""
}