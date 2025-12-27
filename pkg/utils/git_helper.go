package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"cangje-docs-mcp/pkg/types"
)

const (
	// 仓颉文档仓库URL
	CangjieRepoURL = "https://gitcode.com/Cangjie/CangjieCorpus.git"
)

// GetDefaultDocumentDir 获取默认文档目录
func GetDefaultDocumentDir() (string, error) {
	// Windows: 使用可执行文件同目录
	if runtime.GOOS == "windows" {
		exePath, err := os.Executable()
		if err != nil {
			return "", fmt.Errorf("无法获取可执行文件路径: %w", err)
		}
		exeDir := filepath.Dir(exePath)
		return filepath.Join(exeDir, "CangjieCorpus"), nil
	}

	// 其他系统: 使用 ~/.config/cangje-docs-mcp/CangjieCorpus
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("无法获取用户主目录: %w", err)
	}
	configDir := filepath.Join(homeDir, ".config", "cangje-docs-mcp")
	return filepath.Join(configDir, "CangjieCorpus"), nil
}

// EnsureDocuments 确保文档存在并可访问
// 如果文档不存在，会自动克隆；如果存在，会自动更新
func EnsureDocuments(docDir string, autoUpdate bool) error {
	// 检查文档目录是否存在
	_, statErr := os.Stat(docDir)

	// 如果文档目录存在
	if statErr == nil {
		if autoUpdate {
			fmt.Fprintf(os.Stderr, "正在更新仓颉文档...\n")
			if err := UpdateDocuments(docDir); err != nil {
				fmt.Fprintf(os.Stderr, "警告: 文档更新失败: %v\n", err)
				fmt.Fprintf(os.Stderr, "将继续使用现有文档\n")
				return nil // 更新失败不阻塞启动
			}
			fmt.Fprintf(os.Stderr, "✓ 文档更新完成\n")
		}
		return nil
	}

	// 文档目录不存在，需要克隆
	if os.IsNotExist(statErr) {
		fmt.Fprintf(os.Stderr, "仓颉文档不存在，正在从远程仓库克隆...\n")
		fmt.Fprintf(os.Stderr, "仓库: %s\n", CangjieRepoURL)
		fmt.Fprintf(os.Stderr, "目标目录: %s\n", docDir)

		if err := CloneDocuments(docDir); err != nil {
			return fmt.Errorf("克隆文档失败: %w", err)
		}

		fmt.Fprintf(os.Stderr, "✓ 文档克隆完成\n")
		return nil
	}

	return fmt.Errorf("无法访问文档目录: %w", statErr)
}

// CloneDocuments 克隆仓颉文档仓库
func CloneDocuments(docDir string) error {
	// 创建父目录
	parentDir := filepath.Dir(docDir)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 检查 git 是否可用
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("系统未安装 git，请先安装 git: %w", err)
	}

	// 执行 git clone
	cmd := exec.Command("git", "clone", "--depth", "1", CangjieRepoURL, docDir)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// 清理可能部分创建的目录
		os.RemoveAll(docDir)
		return err
	}

	return nil
}

// UpdateDocuments 更新仓颉文档仓库
func UpdateDocuments(docDir string) error {
	// 检查 git 是否可用
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("系统未安装 git: %w", err)
	}

	// 检查是否是 git 仓库
	gitDir := filepath.Join(docDir, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		return fmt.Errorf("不是有效的 git 仓库: %w", err)
	}

	// 执行 git fetch
	cmd := exec.Command("git", "-C", docDir, "fetch", "--all")
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git fetch 失败: %w", err)
	}

	// 检测默认分支名称（可能是 main 或 master）
	branch := getDefaultBranch(docDir)

	// 执行 git reset --hard (强制更新)
	cmd = exec.Command("git", "-C", docDir, "reset", "--hard", "origin/"+branch)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git reset 失败: %w", err)
	}

	// 清理未跟踪的文件
	cmd = exec.Command("git", "-C", docDir, "clean", "-fd")
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		// 清理失败不阻塞
		fmt.Fprintf(os.Stderr, "警告: git clean 失败: %v\n", err)
	}

	return nil
}

// getDefaultBranch 获取仓库的默认分支名称
func getDefaultBranch(docDir string) string {
	// 尝试获取当前分支的远程跟踪分支
	cmd := exec.Command("git", "-C", docDir, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}")
	output, err := cmd.Output()
	if err == nil {
		remoteBranch := strings.TrimSpace(string(output))
		if parts := strings.Split(remoteBranch, "/"); len(parts) == 2 {
			return parts[1]
		}
	}

	// 尝试 origin/main
	cmd = exec.Command("git", "-C", docDir, "rev-parse", "--verify", "origin/main")
	if err := cmd.Run(); err == nil {
		return "main"
	}

	// 尝试 origin/master
	cmd = exec.Command("git", "-C", docDir, "rev-parse", "--verify", "origin/master")
	if err := cmd.Run(); err == nil {
		return "master"
	}

	// 默认返回 main
	return "main"
}

// GetDocumentDir 获取文档目录（优先使用指定的，否则使用默认的）
func GetDocumentDir(specifiedDir string) (string, error) {
	if specifiedDir != "" {
		return specifiedDir, nil
	}

	// 使用默认目录
	defaultDir, err := GetDefaultDocumentDir()
	if err != nil {
		// 回退到旧的默认路径
		return types.DefaultDocumentRootPath, nil
	}

	return defaultDir, nil
}
