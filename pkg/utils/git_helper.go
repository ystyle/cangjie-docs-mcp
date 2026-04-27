package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"cangje-docs-mcp/pkg/types"
)

const (
	WorktreeDirPrefix = "CangjieCorpus-"
	CangjieRepoURL    = "https://gitcode.com/Cangjie/CangjieCorpus.git"
)

func CheckGit() error {
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("系统未安装 git，请先安装 git: %w", err)
	}
	return nil
}

func FullCloneDocuments(docDir string) error {
	if err := CheckGit(); err != nil {
		return err
	}
	parentDir := filepath.Dir(docDir)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}
	cmd := exec.Command("git", "clone", CangjieRepoURL, docDir)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		os.RemoveAll(docDir)
		return err
	}
	return nil
}

func FetchAll(docDir string) error {
	if err := CheckGit(); err != nil {
		return err
	}
	cmd := exec.Command("git", "-C", docDir, "fetch", "--all")
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func UpdateDefaultBranch(docDir string) error {
	if err := CheckGit(); err != nil {
		return err
	}
	headBranch, err := getRemoteHEAD(docDir)
	if err != nil {
		headBranch = "main"
	}
	cmd := exec.Command("git", "-C", docDir, "reset", "--hard", "origin/"+headBranch)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	cmd = exec.Command("git", "-C", docDir, "clean", "-fd")
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	cmd.Run()
	return nil
}

func ListVersionBranches(docDir string) ([]types.VersionBranch, error) {
	if err := CheckGit(); err != nil {
		return nil, err
	}
	cmd := exec.Command("git", "-C", docDir, "ls-remote", "--heads", "origin")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ls-remote 失败: %w", err)
	}

	branches := parseBranches(string(output))
	defaultBranch, _ := getRemoteHEAD(docDir)
	sorted := sortSemverBranches(branches)

	var versions []types.VersionBranch
	for _, b := range sorted {
		v := types.VersionBranch{Name: b}
		if b == defaultBranch {
			v.IsDefault = true
		}
		versions = append(versions, v)
	}
	if len(versions) > 0 {
		versions[0].IsLatest = true
	}
	return versions, nil
}

func EnsureWorktree(docDir, version string) (string, error) {
	worktreeDir := worktreePath(docDir, version)
	if _, err := os.Stat(worktreeDir); err == nil {
		return worktreeDir, nil
	}
	if err := CheckGit(); err != nil {
		return "", err
	}
	cmd := exec.Command("git", "-C", docDir, "fetch", "origin", version)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("fetch 分支 %s 失败: %w", version, err)
	}
	cmd = exec.Command("git", "-C", docDir, "show-ref", "--verify", "--quiet", "refs/heads/"+version)
	if err := cmd.Run(); err != nil {
		cmd = exec.Command("git", "-C", docDir, "branch", version, "origin/"+version)
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("创建本地分支 %s 失败: %w", version, err)
		}
	}
	cmd = exec.Command("git", "-C", docDir, "worktree", "add", worktreeDir, version)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("创建 worktree %s 失败: %w", version, err)
	}
	return worktreeDir, nil
}

func GetVersionDocDir(docDir, version string) string {
	defaultBranch, _ := getRemoteHEAD(docDir)
	if version == "" || version == defaultBranch {
		return docDir
	}
	return worktreePath(docDir, version)
}

func GetRemoteHEAD(docDir string) (string, error) {
	return getRemoteHEAD(docDir)
}

// GetDefaultDocumentDir 获取默认文档目录
func GetDefaultDocumentDir() (string, error) {
	if runtime.GOOS == "windows" {
		exePath, err := os.Executable()
		if err != nil {
			return "", fmt.Errorf("无法获取可执行文件路径: %w", err)
		}
		exeDir := filepath.Dir(exePath)
		return filepath.Join(exeDir, "CangjieCorpus"), nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("无法获取用户主目录: %w", err)
	}
	configDir := filepath.Join(homeDir, ".config", "cangje-docs-mcp")
	return filepath.Join(configDir, "CangjieCorpus"), nil
}

// GetDocumentDir 获取文档目录（优先使用指定的，否则使用默认的）
func GetDocumentDir(specifiedDir string) (string, error) {
	if specifiedDir != "" {
		return specifiedDir, nil
	}

	defaultDir, err := GetDefaultDocumentDir()
	if err != nil {
		return types.DefaultDocumentRootPath, nil
	}

	return defaultDir, nil
}

func EnsureDocuments(docDir string, autoUpdate bool) error {
	if err := CheckGit(); err != nil {
		return err
	}
	_, statErr := os.Stat(docDir)
	if statErr == nil {
		if autoUpdate {
			fmt.Fprintf(os.Stderr, "正在更新仓颉文档...\n")
			if err := FetchAll(docDir); err != nil {
				fmt.Fprintf(os.Stderr, "警告: 文档更新失败: %v\n", err)
				return nil
			}
			if err := UpdateDefaultBranch(docDir); err != nil {
				fmt.Fprintf(os.Stderr, "警告: 默认分支更新失败: %v\n", err)
			}
			fmt.Fprintf(os.Stderr, "✓ 文档更新完成\n")
		}
		return nil
	}
	if os.IsNotExist(statErr) {
		fmt.Fprintf(os.Stderr, "仓颉文档不存在，正在从远程仓库完整克隆...\n")
		if err := FullCloneDocuments(docDir); err != nil {
			return fmt.Errorf("完整克隆文档失败: %w", err)
		}
		fmt.Fprintf(os.Stderr, "✓ 文档克隆完成\n")
		return nil
	}
	return fmt.Errorf("无法访问文档目录: %w", statErr)
}

func worktreePath(docDir, version string) string {
	return docDir + "-" + version
}

func parseBranches(output string) []string {
	var branches []string
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		ref := parts[1]
		if !strings.HasPrefix(ref, "refs/heads/") {
			continue
		}
		name := strings.TrimPrefix(ref, "refs/heads/")
		if types.VersionBranchPattern.MatchString(name) {
			branches = append(branches, name)
		}
	}
	return branches
}

func sortSemverBranches(branches []string) []string {
	sort.Slice(branches, func(i, j int) bool {
		return compareSemver(branches[i], branches[j]) > 0
	})
	return branches
}

func compareSemver(a, b string) int {
	pa := strings.Split(a, ".")
	pb := strings.Split(b, ".")
	for i := 0; i < 3; i++ {
		va, _ := strconv.Atoi(pa[i])
		vb, _ := strconv.Atoi(pb[i])
		if va != vb {
			return va - vb
		}
	}
	return 0
}

func getRemoteHEAD(docDir string) (string, error) {
	cmd := exec.Command("git", "-C", docDir, "symbolic-ref", "--short", "refs/remotes/origin/HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	full := strings.TrimSpace(string(output))
	parts := strings.Split(full, "/")
	if len(parts) > 1 {
		return parts[len(parts)-1], nil
	}
	return full, nil
}
