// Package dirmap computes per-directory ownership from a commit list.
package dirmap

import (
	"fmt"
	"io"
	"math"
	"path"
	"sort"
	"strings"

	"github.com/yourusername/git-summary/internal/gitlog"
)

// DirStat holds activity info for one directory.
type DirStat struct {
	Path     string
	Changes  int     // total file-touches in this dir
	TopOwner string  // author with most touches
	OwnerPct float64 // that author's share within this dir
	Authors  []AuthorShare
}

// AuthorShare is one author's touch count in a directory.
type AuthorShare struct {
	Name    string
	Changes int
	Pct     float64
}

// Result is the full directory breakdown.
type Result struct {
	Dirs []DirStat
}

// Compute builds the directory ownership map from commits.
func Compute(commits []gitlog.Commit, maxDepth int) Result {
	if maxDepth <= 0 {
		maxDepth = 2
	}

	type key struct{ dir, author string }
	counts := map[key]int{}
	dirTotal := map[string]int{}

	for _, c := range commits {
		for _, f := range c.Files {
			dir := dirAtDepth(f, maxDepth)
			k := key{dir, c.Author}
			counts[k]++
			dirTotal[dir]++
		}
	}

	// Build DirStats
	var dirs []DirStat
	seen := map[string]bool{}
	for k, cnt := range counts {
		dir := k.dir
		if seen[dir] {
			continue
		}
		seen[dir] = true

		total := dirTotal[dir]

		authorMap := map[string]int{}
		for kk, v := range counts {
			if kk.dir == dir {
				authorMap[kk.author] += v
			}
		}

		var shares []AuthorShare
		for name, n := range authorMap {
			shares = append(shares, AuthorShare{
				Name:    name,
				Changes: n,
				Pct:     float64(n) / float64(total) * 100,
			})
		}
		sort.Slice(shares, func(i, j int) bool {
			return shares[i].Changes > shares[j].Changes
		})

		ds := DirStat{
			Path:    dir,
			Changes: total,
			Authors: shares,
		}
		if len(shares) > 0 {
			ds.TopOwner = shares[0].Name
			ds.OwnerPct = shares[0].Pct
			_ = cnt
		}
		dirs = append(dirs, ds)
	}

	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i].Changes > dirs[j].Changes
	})

	// Cap at top 15 dirs
	if len(dirs) > 15 {
		dirs = dirs[:15]
	}

	return Result{Dirs: dirs}
}

// dirAtDepth returns the directory path at a maximum depth.
func dirAtDepth(filePath string, depth int) string {
	filePath = strings.ReplaceAll(filePath, "\\", "/")
	dir := path.Dir(filePath)
	if dir == "." {
		return "(root)"
	}
	parts := strings.Split(dir, "/")
	if len(parts) > depth {
		parts = parts[:depth]
	}
	return strings.Join(parts, "/")
}

// Renderer

const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	dim    = "\033[2m"
	green  = "\033[32m"
	yellow = "\033[33m"
	cyan   = "\033[36m"
	red    = "\033[31m"
)

func col(code, s string) string { return code + s + reset }

// Render prints the directory breakdown to w.
func Render(w io.Writer, r Result, noColor bool) {
	p := func(f string, a ...any) { fmt.Fprintf(w, f+"\n", a...) }
	c := func(code, s string) string {
		if noColor {
			return s
		}
		return col(code, s)
	}

	p("")
	p(c(bold+cyan, "  Directory Breakdown"))
	p(c(dim, "  "+strings.Repeat("─", 60)))
	p("  %-28s %7s  %-18s  %s",
		c(dim, "directory"),
		c(dim, "changes"),
		c(dim, "top owner"),
		c(dim, "ownership bar"),
	)
	p(c(dim, "  "+strings.Repeat("─", 60)))

	maxChanges := 0
	if len(r.Dirs) > 0 {
		maxChanges = r.Dirs[0].Changes
	}

	barWidth := 18
	for _, d := range r.Dirs {
		barLen := 0
		if maxChanges > 0 {
			barLen = int(math.Round(float64(d.Changes) / float64(maxChanges) * float64(barWidth)))
		}
		if barLen < 1 && d.Changes > 0 {
			barLen = 1
		}
		bar := strings.Repeat("█", barLen) + strings.Repeat("░", barWidth-barLen)

		ownerLabel := truncate(d.TopOwner, 14)
		if d.OwnerPct < 100 {
			ownerLabel = fmt.Sprintf("%-14s %3.0f%%", truncate(d.TopOwner, 14), d.OwnerPct)
		}

		dirColor := cyan
		if d.OwnerPct >= 90 {
			dirColor = yellow
		}

		p("  %-28s %7d  %-18s  %s",
			c(dirColor, truncate(d.Path, 28)),
			d.Changes,
			c(dim, ownerLabel),
			c(green, bar),
		)

		// Show top 3 authors if there are multiple
		if len(d.Authors) > 1 {
			for i, a := range d.Authors {
				if i >= 3 {
					break
				}
				p("  %s %-22s %3.0f%%",
					c(dim, "    ├"),
					c(dim, truncate(a.Name, 22)),
					a.Pct,
				)
			}
		}
	}
	p("")
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
