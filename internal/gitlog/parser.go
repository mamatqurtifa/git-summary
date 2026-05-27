// Package gitlog handles parsing raw git log output into structured commits.
package gitlog

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Commit represents a single parsed git commit.
type Commit struct {
	Hash      string
	Author    string
	Email     string
	Timestamp time.Time
	Subject   string
	Added     int
	Deleted   int
	Files     []string
}

// Options configures what range of commits to fetch.
type Options struct {
	RepoPath string
	Since    string
	Until    string
	Author   string
	Branch   string
}

// Parse runs git log in the given repo and returns structured commits.
func Parse(opts Options) ([]Commit, error) {
	if err := validateRepo(opts.RepoPath); err != nil {
		return nil, err
	}

	args := buildArgs(opts)
	out, err := runGit(opts.RepoPath, args...)
	if err != nil {
		return nil, fmt.Errorf("git log failed: %w", err)
	}

	return parseOutput(out), nil
}

func validateRepo(path string) error {
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("path not found: %s", path)
	}
	_, err := runGit(path, "rev-parse", "--git-dir")
	if err != nil {
		return fmt.Errorf("not a git repository: %s", path)
	}
	return nil
}

// separator between commit records in the log output.
const sep = "---COMMIT---"

func buildArgs(opts Options) []string {
	// Format: hash|author|email|unix timestamp|subject
	format := fmt.Sprintf("--format=%s%%n%%H|%%an|%%ae|%%at|%%s", sep)
	args := []string{
		"log",
		format,
		"--numstat",
	}

	if opts.Since != "" {
		args = append(args, "--since="+opts.Since)
	}
	if opts.Until != "" {
		args = append(args, "--until="+opts.Until)
	}
	if opts.Author != "" {
		args = append(args, "--author="+opts.Author)
	}
	if opts.Branch != "" {
		args = append(args, opts.Branch)
	}

	return args
}

func runGit(dir string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("%s", msg)
	}
	return stdout.Bytes(), nil
}

func parseOutput(data []byte) []Commit {
	raw := string(data)
	blocks := strings.Split(raw, sep)

	var commits []Commit
	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}
		c := parseBlock(block)
		if c != nil {
			commits = append(commits, *c)
		}
	}
	return commits
}

func parseBlock(block string) *Commit {
	lines := strings.Split(strings.TrimSpace(block), "\n")
	if len(lines) < 1 {
		return nil
	}

	var headerLine string
	var rest []string
	for i, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" && strings.Contains(l, "|") {
			headerLine = l
			rest = lines[i+1:]
			break
		}
	}
	if headerLine == "" {
		return nil
	}

	parts := strings.SplitN(headerLine, "|", 5)
	if len(parts) < 4 {
		return nil
	}

	ts, _ := strconv.ParseInt(strings.TrimSpace(parts[3]), 10, 64)
	subject := ""
	if len(parts) == 5 {
		subject = strings.TrimSpace(parts[4])
	}

	c := &Commit{
		Hash:      strings.TrimSpace(parts[0]),
		Author:    strings.TrimSpace(parts[1]),
		Email:     strings.TrimSpace(parts[2]),
		Timestamp: time.Unix(ts, 0),
		Subject:   subject,
	}

	for _, line := range rest {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		cols := strings.Fields(line)
		if len(cols) < 3 {
			continue
		}
		added, _ := strconv.Atoi(cols[0])
		deleted, _ := strconv.Atoi(cols[1])
		c.Added += added
		c.Deleted += deleted
		c.Files = append(c.Files, cols[2])
	}

	return c
}
