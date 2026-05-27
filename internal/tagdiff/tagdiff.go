// Package tagdiff generates a changelog between two git tags or refs.
package tagdiff

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"sort"
	"strings"
	"time"
)

// Entry is a single commit in the changelog.
type Entry struct {
	Hash      string
	Author    string
	Timestamp time.Time
	Subject   string
	Category  string
}

// Result holds the full tag diff.
type Result struct {
	From    string
	To      string
	Entries []Entry
	Stats   struct {
		Total    int
		ByCategory map[string]int
	}
}

// Run computes the diff between fromRef and toRef in repoPath.
func Run(repoPath, fromRef, toRef string) (*Result, error) {
	if toRef == "" {
		toRef = "HEAD"
	}

	if err := validateRef(repoPath, fromRef); err != nil {
		return nil, fmt.Errorf("ref %q not found: %w", fromRef, err)
	}

	rangeArg := fromRef + ".." + toRef
	out, err := runGit(repoPath, "log", rangeArg,
		"--format=%H|%an|%at|%s",
		"--no-merges",
	)
	if err != nil {
		return nil, fmt.Errorf("git log failed: %w", err)
	}

	entries := parseEntries(out)

	res := &Result{From: fromRef, To: toRef, Entries: entries}
	res.Stats.Total = len(entries)
	res.Stats.ByCategory = map[string]int{}
	for _, e := range entries {
		res.Stats.ByCategory[e.Category]++
	}
	return res, nil
}

// ListTags returns all tags in the repo, newest first.
func ListTags(repoPath string) ([]string, error) {
	out, err := runGit(repoPath, "tag", "--sort=-version:refname")
	if err != nil {
		return nil, err
	}
	var tags []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if t := strings.TrimSpace(line); t != "" {
			tags = append(tags, t)
		}
	}
	return tags, nil
}

func validateRef(repoPath, ref string) error {
	_, err := runGit(repoPath, "rev-parse", "--verify", ref)
	return err
}

func parseEntries(data []byte) []Entry {
	var entries []Entry
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 4)
		if len(parts) < 4 {
			continue
		}
		var ts int64
		fmt.Sscanf(parts[2], "%d", &ts)
		subject := parts[3]
		e := Entry{
			Hash:      parts[0][:7],
			Author:    parts[1],
			Timestamp: time.Unix(ts, 0),
			Subject:   subject,
			Category:  categorise(subject),
		}
		entries = append(entries, e)
	}
	return entries
}

// categorise detects conventional commit prefixes.
func categorise(subject string) string {
	lower := strings.ToLower(subject)
	prefixes := []struct {
		prefix string
		cat    string
	}{
		{"feat", "feat"},
		{"feature", "feat"},
		{"fix", "fix"},
		{"bug", "fix"},
		{"hotfix", "fix"},
		{"docs", "docs"},
		{"doc", "docs"},
		{"chore", "chore"},
		{"ci", "chore"},
		{"build", "chore"},
		{"refactor", "refactor"},
		{"perf", "perf"},
		{"test", "test"},
		{"style", "style"},
		{"revert", "revert"},
	}
	for _, p := range prefixes {
		if strings.HasPrefix(lower, p.prefix+"(") ||
			strings.HasPrefix(lower, p.prefix+":") ||
			strings.HasPrefix(lower, p.prefix+" ") {
			return p.cat
		}
	}
	return "other"
}

// Renderers

const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	dim    = "\033[2m"
	green  = "\033[32m"
	yellow = "\033[33m"
	cyan   = "\033[36m"
	red    = "\033[31m"
	blue   = "\033[34m"
)

func col(code, s string) string { return code + s + reset }

var categoryMeta = map[string]struct {
	icon  string
	label string
	color string
}{
	"feat":     {"✨", "New features", cyan},
	"fix":      {"🐛", "Bug fixes", green},
	"docs":     {"📝", "Documentation", blue},
	"refactor": {"♻", "Refactoring", yellow},
	"perf":     {"⚡", "Performance", yellow},
	"chore":    {"🔧", "Chores", dim},
	"test":     {"🧪", "Tests", dim},
	"style":    {"🎨", "Style", dim},
	"revert":   {"⏪", "Reverts", red},
	"other":    {"•", "Other changes", dim},
}

var categoryOrder = []string{"feat", "fix", "perf", "refactor", "docs", "test", "style", "chore", "revert", "other"}

// RenderTerminal prints a coloured changelog to w.
func RenderTerminal(w io.Writer, r *Result, noColor bool) {
	p := func(f string, a ...any) { fmt.Fprintf(w, f+"\n", a...) }
	c := func(code, s string) string {
		if noColor {
			return s
		}
		return col(code, s)
	}

	p("")
	p(c(bold+cyan, fmt.Sprintf("  Changelog: %s → %s", r.From, r.To)))
	p(c(dim, "  "+strings.Repeat("─", 50)))
	p("  %s total commits", c(bold, fmt.Sprintf("%d", r.Stats.Total)))
	p("")

	bycat := map[string][]Entry{}
	for _, e := range r.Entries {
		bycat[e.Category] = append(bycat[e.Category], e)
	}

	for _, cat := range categoryOrder {
		entries, ok := bycat[cat]
		if !ok || len(entries) == 0 {
			continue
		}
		meta := categoryMeta[cat]
		p("  %s %s %s",
			meta.icon,
			c(bold+meta.color, meta.label),
			c(dim, fmt.Sprintf("(%d)", len(entries))))
		for _, e := range entries {
			p("  %s %s  %s",
				c(dim, e.Hash),
				e.Subject,
				c(dim, "— "+e.Author))
		}
		p("")
	}
}

// RenderMarkdown writes a markdown changelog to w.
func RenderMarkdown(w io.Writer, r *Result) {
	p := func(f string, a ...any) { fmt.Fprintf(w, f+"\n", a...) }

	p("# Changelog")
	p("")
	p("## %s → %s", r.From, r.To)
	p("")
	p("_%d commits_", r.Stats.Total)
	p("")

	bycat := map[string][]Entry{}
	for _, e := range r.Entries {
		bycat[e.Category] = append(bycat[e.Category], e)
	}

	for _, cat := range categoryOrder {
		entries, ok := bycat[cat]
		if !ok || len(entries) == 0 {
			continue
		}
		meta := categoryMeta[cat]
		p("### %s %s", meta.icon, strings.Title(meta.label))
		p("")
		for _, e := range entries {
			p("- `%s` %s (%s)", e.Hash, e.Subject, e.Author)
		}
		p("")
	}

	p("### Summary by type")
	p("")
	p("| Type | Count |")
	p("|------|-------|")

	type pair struct{ k string; v int }
	var pairs []pair
	for k, v := range r.Stats.ByCategory {
		pairs = append(pairs, pair{k, v})
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].v > pairs[j].v })
	for _, kv := range pairs {
		meta := categoryMeta[kv.k]
		p("| %s %s | %d |", meta.icon, kv.k, kv.v)
	}

	p("")
	p("---")
	p("_Generated by [git-summary](https://github.com/mamatqurtifa/git-summary)_")
}

func runGit(dir string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(errb.String())
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("%s", msg)
	}
	return out.Bytes(), nil
}
