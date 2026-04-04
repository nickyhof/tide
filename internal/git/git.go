package git

import (
	"os/exec"
	"path/filepath"
	"strings"
)

// FileStatus represents the staging state of a file.
type FileStatus struct {
	Path    string
	AbsPath string
	Index   byte // status in index (staged): ' ', 'M', 'A', 'D', 'R', '?', etc.
	Work    byte // status in worktree: ' ', 'M', 'D', '?', etc.
}

// IsStaged returns true if the file has staged changes.
func (f FileStatus) IsStaged() bool {
	return f.Index != ' ' && f.Index != '?'
}

// IsModified returns true if the file has unstaged changes.
func (f FileStatus) IsModified() bool {
	return f.Work == 'M' || f.Work == 'D'
}

// IsUntracked returns true if the file is untracked.
func (f FileStatus) IsUntracked() bool {
	return f.Index == '?' && f.Work == '?'
}

// StatusLabel returns a human-readable label.
func (f FileStatus) StatusLabel() string {
	switch {
	case f.Index == '?' && f.Work == '?':
		return "Untracked"
	case f.Index == 'A':
		return "Added"
	case f.Index == 'M':
		return "Staged"
	case f.Index == 'D':
		return "Deleted"
	case f.Index == 'R':
		return "Renamed"
	case f.Work == 'M':
		return "Modified"
	case f.Work == 'D':
		return "Deleted"
	default:
		return string([]byte{f.Index, f.Work})
	}
}

// Repo provides git operations for a repository.
type Repo struct {
	Root string // absolute path to repo root
}

// Open finds the git repo root for the given directory. Returns nil if not a git repo.
func Open(dir string) *Repo {
	out, err := runGit(dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return nil
	}
	root := strings.TrimSpace(out)
	if root == "" {
		return nil
	}
	return &Repo{Root: root}
}

// Branch returns the current branch name (or HEAD short ref if detached).
func (r *Repo) Branch() string {
	out, err := runGit(r.Root, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

// Status returns the list of files with changes.
func (r *Repo) Status() []FileStatus {
	out, err := runGit(r.Root, "status", "--porcelain", "-uall")
	if err != nil {
		return nil
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	var files []FileStatus
	for _, line := range lines {
		if len(line) < 4 {
			continue
		}
		idx := line[0]
		work := line[1]
		path := line[3:]
		// Handle renames: "R  old -> new"
		if idx == 'R' {
			if parts := strings.SplitN(path, " -> ", 2); len(parts) == 2 {
				path = parts[1]
			}
		}
		files = append(files, FileStatus{
			Path:    path,
			AbsPath: filepath.Join(r.Root, path),
			Index:   idx,
			Work:    work,
		})
	}
	return files
}

// StagedCount returns how many files have staged changes.
func (r *Repo) StagedCount() int {
	n := 0
	for _, f := range r.Status() {
		if f.IsStaged() {
			n++
		}
	}
	return n
}

// ModifiedCount returns how many files have unstaged changes.
func (r *Repo) ModifiedCount() int {
	n := 0
	for _, f := range r.Status() {
		if f.IsModified() {
			n++
		}
	}
	return n
}

// UntrackedCount returns how many untracked files exist.
func (r *Repo) UntrackedCount() int {
	n := 0
	for _, f := range r.Status() {
		if f.IsUntracked() {
			n++
		}
	}
	return n
}

// DiffFile returns the diff for a specific file.
// If staged is true, returns the staged diff (--cached).
func (r *Repo) DiffFile(path string, staged bool) string {
	args := []string{"diff", "--no-color"}
	if staged {
		args = append(args, "--cached")
	}
	args = append(args, "--", path)
	out, err := runGit(r.Root, args...)
	if err != nil {
		return ""
	}
	return out
}

// DiffUntracked returns the contents of an untracked file (shown as an add diff).
func (r *Repo) DiffUntracked(path string) string {
	args := []string{"diff", "--no-color", "--no-index", "/dev/null", path}
	out, _ := runGit(r.Root, args...)
	return out
}

// Stage stages a file.
func (r *Repo) Stage(path string) error {
	_, err := runGit(r.Root, "add", "--", path)
	return err
}

// Unstage unstages a file.
func (r *Repo) Unstage(path string) error {
	_, err := runGit(r.Root, "reset", "HEAD", "--", path)
	return err
}

func runGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	return string(out), err
}
