// Package compare analyses and diffs two branches side by side.
package compare

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/yourusername/git-summary/internal/gitlog"
	"github.com/yourusername/git-summary/internal/stats"
)

// BranchSummary is a stats.Summary tagged with a branch name.
type BranchSummary struct {
	Branch  string
	Summary stats.Summary
}

// Result holds both branches plus divergence info.
type Result struct {
	A            BranchSummary
	B            BranchSummary
	OnlyInA      []string
	OnlyInB      []string
	SharedFiles  []string
}

// Run computes summaries for both branches in the given repo.
func Run(repoPath, branchA, branchB, since string, topN int) (*Result, error) {
	sumA, err := summaryForBranch(repoPath, branchA, since, topN)
	if err != nil {
		return nil, fmt.Errorf("branch %q: %w", branchA, err)
	}
	sumB, err := summaryForBranch(repoPath, branchB, since, topN)
	if err != nil {
		return nil, fmt.Errorf("branch %q: %w", branchB, err)
	}

	res := &Result{
		A: BranchSummary{Branch: branchA, Summary: sumA},
		B: BranchSummary{Branch: branchB, Summary: sumB},
	}
	res.OnlyInA, res.OnlyInB = diffContributors(sumA, sumB)
	res.SharedFiles = sharedFiles(sumA, sumB)
	return res, nil
}

func summaryForBranch(repo, branch, since string, topN int) (stats.Summary, error) {
	commits, err := gitlog.Parse(gitlog.Options{
		RepoPath: repo,
		Branch:   branch,
		Since:    since,
	})
	if err != nil {
		return stats.Summary{}, err
	}
	return stats.Compute(commits, topN), nil
}

func diffContributors(a, b stats.Summary) (onlyA, onlyB []string) {
	setA := map[string]bool{}
	setB := map[string]bool{}
	for _, c := range a.Contributors {
		setA[c.Name] = true
	}
	for _, c := range b.Contributors {
		setB[c.Name] = true
	}
	for name := range setA {
		if !setB[name] {
			onlyA = append(onlyA, name)
		}
	}
	for name := range setB {
		if !setA[name] {
			onlyB = append(onlyB, name)
		}
	}
	sort.Strings(onlyA)
	sort.Strings(onlyB)
	return
}

func sharedFiles(a, b stats.Summary) []string {
	setA := map[string]bool{}
	for _, f := range a.TopFiles {
		setA[f.Path] = true
	}
	var shared []string
	for _, f := range b.TopFiles {
		if setA[f.Path] {
			shared = append(shared, f.Path)
		}
	}
	return shared
}

// Renderer
const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	dim    = "\033[2m"
	cyan   = "\033[36m"
	green  = "\033[32m"
	yellow = "\033[33m"
	red    = "\033[31m"
	blue   = "\033[34m"
)

func col(code, s string) string { return code + s + reset }

func Render(w io.Writer, r *Result, noColor bool) {
	if noColor {
		renderPlain(w, r)
		return
	}
	renderColor(w, r)
}

func renderColor(w io.Writer, r *Result) {
	p := func(f string, a ...any) { fmt.Fprintf(w, f+"\n", a...) }

	width := 28
	divider := "│"

	p("")
	p(col(bold+cyan, "  Branch Comparison"))
	p(col(dim, "  "+strings.Repeat("─", width*2+5)))
	p("  %-*s %s %-*s",
		width, col(bold+yellow, truncate(r.A.Branch, width)),
		col(dim, divider),
		width, col(bold+yellow, truncate(r.B.Branch, width)))
	p(col(dim, "  "+strings.Repeat("─", width*2+5)))

	rows := []struct{ label, a, b string }{
		{"Commits",
			fmt.Sprintf("%d", r.A.Summary.TotalCommits),
			fmt.Sprintf("%d", r.B.Summary.TotalCommits)},
		{"Active days",
			fmt.Sprintf("%d", r.A.Summary.ActiveDays),
			fmt.Sprintf("%d", r.B.Summary.ActiveDays)},
		{"Lines added",
			fmt.Sprintf("+%d", r.A.Summary.TotalAdded),
			fmt.Sprintf("+%d", r.B.Summary.TotalAdded)},
		{"Lines deleted",
			fmt.Sprintf("-%d", r.A.Summary.TotalDeleted),
			fmt.Sprintf("-%d", r.B.Summary.TotalDeleted)},
		{"Contributors",
			fmt.Sprintf("%d", len(r.A.Summary.Contributors)),
			fmt.Sprintf("%d", len(r.B.Summary.Contributors))},
	}

	for _, row := range rows {
		p("  %-18s %-*s %s %-*s",
			col(dim, row.label),
			width-18, col(green, row.a),
			col(dim, divider),
			width, col(green, row.b))
	}

	p(col(dim, "  "+strings.Repeat("─", width*2+5)))
	p("")

	// Top contributors comparison
	p(col(bold+yellow, "  Top contributors"))
	maxC := len(r.A.Summary.Contributors)
	if len(r.B.Summary.Contributors) > maxC {
		maxC = len(r.B.Summary.Contributors)
	}
	for i := 0; i < maxC && i < 5; i++ {
		aName, bName := "", ""
		aCommits, bCommits := "", ""
		if i < len(r.A.Summary.Contributors) {
			c := r.A.Summary.Contributors[i]
			aName = truncate(c.Name, 14)
			aCommits = fmt.Sprintf("%d", c.Commits)
		}
		if i < len(r.B.Summary.Contributors) {
			c := r.B.Summary.Contributors[i]
			bName = truncate(c.Name, 14)
			bCommits = fmt.Sprintf("%d", c.Commits)
		}
		p("  %-14s %-6s %s %-14s %-6s",
			col(bold, aName), col(green, aCommits),
			col(dim, divider),
			col(bold, bName), col(green, bCommits))
	}

	// Unique contributors
	if len(r.OnlyInA) > 0 {
		p("")
		p("  %s only in %s: %s",
			col(dim, "contributors"),
			col(yellow, r.A.Branch),
			col(cyan, strings.Join(r.OnlyInA, ", ")))
	}
	if len(r.OnlyInB) > 0 {
		p("  %s only in %s: %s",
			col(dim, "contributors"),
			col(yellow, r.B.Branch),
			col(cyan, strings.Join(r.OnlyInB, ", ")))
	}

	// Shared hot files
	if len(r.SharedFiles) > 0 {
		p("")
		p("  %s", col(bold+yellow, "Files changed in both branches"))
		for _, f := range r.SharedFiles {
			p("  %s %s", col(yellow, "⚠"), col(dim, f))
		}
	}
	p("")
}

func renderPlain(w io.Writer, r *Result) {
	p := func(f string, a ...any) { fmt.Fprintf(w, f+"\n", a...) }
	p("Branch Comparison: %s vs %s", r.A.Branch, r.B.Branch)
	p("%-20s %-15s %-15s", "Metric", r.A.Branch, r.B.Branch)
	p("%s", strings.Repeat("-", 52))
	p("%-20s %-15d %-15d", "Commits", r.A.Summary.TotalCommits, r.B.Summary.TotalCommits)
	p("%-20s %-15d %-15d", "Active days", r.A.Summary.ActiveDays, r.B.Summary.ActiveDays)
	p("%-20s +%-14d +%-14d", "Lines added", r.A.Summary.TotalAdded, r.B.Summary.TotalAdded)
	p("%-20s -%-14d -%-14d", "Lines deleted", r.A.Summary.TotalDeleted, r.B.Summary.TotalDeleted)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
