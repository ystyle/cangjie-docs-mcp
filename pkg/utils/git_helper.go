package utils

import (
	"fmt"
	"log/slog"
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
	// WildcardRefspec 是标准 clone 产生的通配 refspec。
	// 单分支 refspec 会导致 git fetch <branch> 不生成远端引用，进而使新版本分支无法使用。
	WildcardRefspec = "+refs/heads/*:refs/remotes/origin/*"
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

// FetchAll 拉取 origin 上所有分支。
// 单远端场景下等价于 git fetch --all 且更快；配合通配 refspec 可带回全部 origin/* 引用。
func FetchAll(docDir string) error {
	if err := CheckGit(); err != nil {
		return err
	}
	_, err := runGit(docDir, "fetch", "origin")
	return err
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

// runGit 执行 git 命令并返回合并后的输出。
// 使用 CombinedOutput 以便把 git 的真实报错带进 error，避免只看到 "exit status 128"。
func runGit(docDir string, args ...string) (string, error) {
	fullArgs := append([]string{"-C", docDir}, args...)
	cmd := exec.Command("git", fullArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			return "", fmt.Errorf("git %s 失败: %w", strings.Join(args, " "), err)
		}
		return "", fmt.Errorf("git %s 失败: %w: %s", strings.Join(args, " "), err, msg)
	}
	return string(out), nil
}

// remoteRefExists 判断 refs/remotes/origin/<version> 是否已存在（纯本地操作，不触网）
func remoteRefExists(docDir, version string) bool {
	cmd := exec.Command("git", "-C", docDir, "show-ref", "--verify", "--quiet", "refs/remotes/origin/"+version)
	return cmd.Run() == nil
}

// RepoHealth 描述语料库的健康状况
type RepoHealth struct {
	IsShallow          bool   `json:"is_shallow"`          // 是否浅克隆（.git/shallow 存在）
	IsSingleBranch     bool   `json:"is_single_branch"`    // refspec 是否非通配（仅跟踪单一分支）
	Refspec            string `json:"refspec"`             // 当前 refspec
	NeedsRepair        bool   `json:"needs_repair"`        // 是否建议完整重建
	HealthyDescription string `json:"healthy_description"` // 面向用户的说明
}

// CheckRepoHealth 检查语料库健康状况。
// 用于在 cangjie_list_versions 中提示 agent：是否需要征询用户后执行完整重建。
func CheckRepoHealth(docDir string) RepoHealth {
	h := RepoHealth{Refspec: "(未配置)"}

	// 浅克隆检测：走 git rev-parse 比直接看 .git/shallow 更稳妥（worktree 场景下路径不同）
	if out, err := runGit(docDir, "rev-parse", "--is-shallow-repository"); err == nil {
		h.IsShallow = strings.TrimSpace(out) == "true"
	}

	// 单分支 refspec 检测
	if out, err := runGit(docDir, "config", "--get", "remote.origin.fetch"); err == nil {
		h.Refspec = strings.TrimSpace(out)
	}
	h.IsSingleBranch = !hasWildcardRefspec(docDir)

	// 结论：只有浅克隆或单分支才算需要重建。
	// 单分支本身已能通过 EnsureWildcardRefspec 自愈，故单独不构成"必须重建"，
	// 但浅克隆会缺失历史，建议一次性重建以获得完整仓库。
	h.NeedsRepair = h.IsShallow || h.IsSingleBranch

	var reasons []string
	if h.IsShallow {
		reasons = append(reasons, "当前语料库是浅克隆（--depth 1），缺少完整提交历史")
	}
	if h.IsSingleBranch {
		reasons = append(reasons, "当前语料库只跟踪单一分支（refspec 非通配），早期版本无法直接使用")
	}
	if len(reasons) == 0 {
		h.HealthyDescription = "语料库健康（完整克隆、通配 refspec）"
	} else {
		reasons = append(reasons, "重建后可获得干净完整的仓库")
		h.HealthyDescription = strings.Join(reasons, "；") + "。"
	}
	return h
}

// hasWildcardRefspec 判断当前 refspec 中是否包含通配形式
func hasWildcardRefspec(docDir string) bool {
	out, err := runGit(docDir, "config", "--get-all", "remote.origin.fetch")
	if err != nil {
		return false
	}
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if strings.TrimSpace(line) == WildcardRefspec {
			return true
		}
	}
	return false
}

// RepairDocuments 完整重建语料库：删除旧目录后重新完整克隆。
//
// 这是破坏性操作（会下载约 300MB），仅应由用户显式确认后调用，
// 因此同时提供 -repair-docs 命令行开关与 cangjie_list_versions 的提示信息。
func RepairDocuments(docDir string) error {
	if err := CheckGit(); err != nil {
		return err
	}
	// 安全校验：确认目标确实是本工具的语料库目录，避免误删任意路径
	if filepath.Base(docDir) != "CangjieCorpus" {
		return fmt.Errorf("拒绝操作：%s 不是预期的 CangjieCorpus 目录，为避免误删已终止", docDir)
	}
	// 先移除所有 worktree，否则主库目录删不干净
	if out, err := runGit(docDir, "worktree", "list", "--porcelain"); err == nil {
		for _, line := range strings.Split(out, "\n") {
			line = strings.TrimSpace(line)
			if !strings.HasPrefix(line, "worktree ") {
				continue
			}
			wtPath := strings.TrimPrefix(line, "worktree ")
			if wtPath == docDir {
				continue
			}
			fmt.Fprintf(os.Stderr, "正在移除 worktree: %s\n", wtPath)
			_, _ = runGit(docDir, "worktree", "remove", "--force", wtPath)
			os.RemoveAll(wtPath)
		}
	}
	if err := os.RemoveAll(docDir); err != nil {
		return fmt.Errorf("删除旧语料库失败: %w", err)
	}
	fmt.Fprintf(os.Stderr, "旧语料库已删除，正在完整重新克隆...\n")
	if err := FullCloneDocuments(docDir); err != nil {
		return fmt.Errorf("重新克隆失败: %w", err)
	}
	if err := EnsureWildcardRefspec(docDir); err != nil {
		return fmt.Errorf("校验远端分支配置失败: %w", err)
	}
	fmt.Fprintf(os.Stderr, "✓ 语料库已完整重建\n")
	return nil
}

// EnsureWildcardRefspec 确保 remote.origin.fetch 使用通配 refspec。
//
// 标准的 git clone 会得到 +refs/heads/*:refs/remotes/origin/*，但单分支克隆
// （--single-branch）或历史上手工改过的仓库会退化成只跟踪某一个分支，此时
// `git fetch origin <new-branch>` 只会写入 FETCH_HEAD，并不会生成
// refs/remotes/origin/<new-branch> 引用，后续基于 origin/<version> 的操作全都会失败。
// 这里做一次性自愈，让老仓库也能正常拉取新版本分支。
func EnsureWildcardRefspec(docDir string) error {
	if err := CheckGit(); err != nil {
		return err
	}
	if hasWildcardRefspec(docDir) {
		return nil
	}
	return setWildcardRefspec(docDir)
}

func setWildcardRefspec(docDir string) error {
	if _, err := runGit(docDir, "config", "--replace-all", "remote.origin.fetch", WildcardRefspec); err != nil {
		return fmt.Errorf("修正 remote.origin.fetch 失败: %w", err)
	}
	fmt.Fprintf(os.Stderr, "✓ 已修正 remote.origin.fetch 为通配形式，正在重新拉取分支引用...\n")
	if _, err := runGit(docDir, "fetch", "origin"); err != nil {
		return fmt.Errorf("重新拉取远端分支失败: %w", err)
	}
	return nil
}

// EnsureWorktree 为指定版本准备 worktree 目录并返回其路径。
//
// 设计前提：主语料库是标准 clone（携带全部 origin/* 引用），版本分支用 worktree 隔离。
// 因此正常情况下这里不需要任何网络操作：
//   - origin/<version> 已存在 → 直接用 worktree add 的 DWIM 语义建跟踪分支
//   - origin/<version> 不存在（远端刚发新版本）→ 先 fetch 一次再重试
func EnsureWorktree(docDir, version string) (string, error) {
	worktreeDir := worktreePath(docDir, version)
	if _, err := os.Stat(worktreeDir); err == nil {
		return worktreeDir, nil
	}
	if err := CheckGit(); err != nil {
		return "", err
	}
	// 老仓库可能仍是单分支 refspec，先自愈，否则拿不到新分支引用
	if err := EnsureWildcardRefspec(docDir); err != nil {
		return "", err
	}
	// 本地没有该远端分支引用时才需要触网
	if !remoteRefExists(docDir, version) {
		if _, err := runGit(docDir, "fetch", "origin"); err != nil {
			return "", fmt.Errorf("拉取远端分支失败: %w", err)
		}
		if !remoteRefExists(docDir, version) {
			return "", fmt.Errorf("远端不存在分支 %s", version)
		}
	}
	// worktree add 具备 DWIM 能力：分支不存在时自动基于 origin/<version> 创建跟踪分支
	if _, err := runGit(docDir, "worktree", "add", worktreeDir, version); err != nil {
		slog.Warn("创建 worktree 失败，尝试清理残留后重试", "version", version, "error", err)
		cleanupFailedWorktree(docDir, worktreeDir, version)
		if _, retryErr := runGit(docDir, "worktree", "add", worktreeDir, version); retryErr != nil {
			return "", fmt.Errorf("创建 worktree %s 失败: %w", version, retryErr)
		}
	}
	return worktreeDir, nil
}

// cleanupFailedWorktree 清理失败留下的半成品 worktree / 分支，保证重试可行。
func cleanupFailedWorktree(docDir, worktreeDir, version string) {
	if _, err := os.Stat(worktreeDir); err == nil {
		if _, err := runGit(docDir, "worktree", "remove", "--force", worktreeDir); err != nil {
			os.RemoveAll(worktreeDir)
		}
	}
	_, _ = runGit(docDir, "worktree", "prune")
	// 仅当该本地分支未被任何 worktree 占用时才删除
	if _, err := runGit(docDir, "show-ref", "--verify", "--quiet", "refs/heads/"+version); err == nil {
		if !worktreeUsesBranch(docDir, version) {
			_, _ = runGit(docDir, "branch", "-D", version)
		}
	}
}

// worktreeUsesBranch 判断某本地分支是否被任一 worktree 占用
func worktreeUsesBranch(docDir, version string) bool {
	out, err := runGit(docDir, "worktree", "list", "--porcelain")
	if err != nil {
		return true // 拿不到信息时保守处理，不删分支
	}
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) == "branch refs/heads/"+version {
			return true
		}
	}
	return false
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
		// 每次都校验 refspec：老仓库若是单分支克隆，不修正就永远拿不到新版本分支。
		// 已是通配形式时这里只是一次本地 git config 读取，不产生网络开销。
		if err := EnsureWildcardRefspec(docDir); err != nil {
			fmt.Fprintf(os.Stderr, "警告: 修正远端分支配置失败: %v\n", err)
		}
		if autoUpdate {
			fmt.Fprintf(os.Stderr, "正在更新仓颉文档...\n")
			if _, err := runGit(docDir, "fetch", "origin"); err != nil {
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
		// 标准 clone 本身就是通配 refspec，这里兜底校验一次
		if err := EnsureWildcardRefspec(docDir); err != nil {
			fmt.Fprintf(os.Stderr, "警告: 校验远端分支配置失败: %v\n", err)
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
	// 逐段比较，段数不足的按 0 处理，避免版本号格式异常时下标越界
	n := len(pa)
	if len(pb) > n {
		n = len(pb)
	}
	for i := 0; i < n; i++ {
		var va, vb int
		if i < len(pa) {
			va, _ = strconv.Atoi(pa[i])
		}
		if i < len(pb) {
			vb, _ = strconv.Atoi(pb[i])
		}
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
