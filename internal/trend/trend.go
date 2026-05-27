// Package trend analyses contributor activity over time and detects risk signals.
package trend

import (
	"fmt"
	"io"
	"math"
	"sort"
	"strings"

	"github.com/yourusername/git-summary/internal/gitlog"
)

// WeeklySlice is one author's commit count for a single week.
type WeeklySlice struct {
	Week    string
	Commits int
}

// AuthorTrend holds time-series data for one contributor.
type AuthorTrend struct {
	Name       string
	Email      string
	Total      int
	Weekly     []WeeklySlice
	Trend      string
	SharePct   float64
}

// BusFactor is the minimum number of people who hold N% of the knowledge.
type BusFactor struct {
	N          int
	Threshold  float64
	Authors    []string
}

// Result is the full trend report.
type Result struct {
	Authors   []AuthorTrend
	BusFactor BusFactor
	AllWeeks  []string
}

// Compute derives trend data from commits.
func Compute(commits []gitlog.Commit, topN int) Result {
	if len(commits) == 0 {
		return Result{}
	}

	// Build per-author per-week buckets
	type key struct{ email, week string }
	buckets := map[key]int{}
	authorMeta := map[string]struct{ name, email string }{}
	weekSet := map[string]bool{}
	authorTotal := map[string]int{}

	for _, c := range commits {
		y, w := c.Timestamp.ISOWeek()
		wk := fmt.Sprintf("%04d-W%02d", y, w)
		k := key{c.Email, wk}
		buckets[k]++
		weekSet[wk] = true
		authorTotal[c.Email]++
		authorMeta[c.Email] = struct{ name, email string }{c.Author, c.Email}
	}

	// Sorted week list
	var weeks []string
	for wk := range weekSet {
		weeks = append(weeks, wk)
	}
	sort.Strings(weeks)

	totalCommits := len(commits)

	// Build AuthorTrend for each author
	var trends []AuthorTrend
	for email, total := range authorTotal {
		meta := authorMeta[email]
		var slices []WeeklySlice
		for _, wk := range weeks {
			slices = append(slices, WeeklySlice{Week: wk, Commits: buckets[key{email, wk}]})
		}
		at := AuthorTrend{
			Name:     meta.name,
			Email:    email,
			Total:    total,
			Weekly:   slices,
			SharePct: float64(total) / float64(totalCommits) * 100,
			Trend:    detectTrend(slices),
		}
		trends = append(trends, at)
	}

	// Sort by total desc
	sort.Slice(trends, func(i, j int) bool {
		return trends[i].Total > trends[j].Total
	})
	if len(trends) > topN {
		trends = trends[:topN]
	}

	bf := computeBusFactor(trends, totalCommits, 0.80)

	return Result{Authors: trends, BusFactor: bf, AllWeeks: weeks}
}

func detectTrend(slices []WeeklySlice) string {
	if len(slices) < 3 {
		return "stable"
	}
	n := len(slices)
	half := n / 2
	if half < 2 {
		half = 2
	}
	recent := slices[n-half:]
	prior := slices[n-half*2 : n-half]

	recentAvg := avg(recent)
	priorAvg := avg(prior)

	if priorAvg == 0 && recentAvg > 0 {
		return "new"
	}
	if recentAvg == 0 && priorAvg > 0 {
		return "gone"
	}
	if priorAvg == 0 {
		return "stable"
	}
	ratio := recentAvg / priorAvg
	if ratio > 1.4 {
		return "rising"
	}
	if ratio < 0.6 {
		return "falling"
	}
	return "stable"
}

func avg(slices []WeeklySlice) float64 {
	if len(slices) == 0 {
		return 0
	}
	sum := 0
	for _, s := range slices {
		sum += s.Commits
	}
	return float64(sum) / float64(len(slices))
}

func computeBusFactor(trends []AuthorTrend, total int, threshold float64) BusFactor {
	cumulative := 0.0
	var authors []string
	for _, t := range trends {
		cumulative += t.SharePct / 100
		authors = append(authors, t.Name)
		if cumulative >= threshold {
			break
		}
	}
	return BusFactor{
		N:         len(authors),
		Threshold: threshold * 100,
		Authors:   authors,
	}
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
	blue   = "\033[34m"
)

func col(code, s string) string { return code + s + reset }

var trendMeta = map[string]struct {
	icon  string
	color string
}{
	"rising":  {"↑", green},
	"falling": {"↓", red},
	"stable":  {"→", dim},
	"new":     {"★", cyan},
	"gone":    {"✗", red},
}

// Render prints the trend report to w.
func Render(w io.Writer, r Result, noColor bool) {
	p := func(f string, a ...any) { fmt.Fprintf(w, f+"\n", a...) }
	c := func(code, s string) string {
		if noColor {
			return s
		}
		return col(code, s)
	}

	p("")
	p(c(bold+cyan, "  Contributor Trends"))
	p(c(dim, "  "+strings.Repeat("─", 60)))

	if len(r.AllWeeks) == 0 {
		p("  No data.")
		return
	}

	displayWeeks := r.AllWeeks
	if len(displayWeeks) > 8 {
		displayWeeks = displayWeeks[len(displayWeeks)-8:]
	}

	p("  %-20s %6s %6s   %-8s  %s",
		c(dim, "author"),
		c(dim, "commits"),
		c(dim, "share"),
		c(dim, "trend"),
		c(dim, "last 8 weeks →"),
	)
	p(c(dim, "  "+strings.Repeat("─", 60)))

	for _, at := range r.Authors {
		tm := trendMeta[at.Trend]
		trendStr := fmt.Sprintf("%s %s", tm.icon, at.Trend)

		sparkline := buildSparkline(at.Weekly, displayWeeks)

		name := truncate(at.Name, 20)
		p("  %-20s %6d %5.1f%%   %-10s  %s",
			c(bold, name),
			at.Total,
			at.SharePct,
			c(tm.color, trendStr),
			c(cyan, sparkline),
		)
	}

	p("")

	bf := r.BusFactor
	bfColor := green
	warning := ""
	if bf.N == 1 {
		bfColor = red
		warning = " ← CRITICAL"
	} else if bf.N <= 2 {
		bfColor = yellow
		warning = " ← at risk"
	}
	p("  %s bus factor: %s %s",
		c(dim, "⚑"),
		c(bold+bfColor, fmt.Sprintf("%d %s own %.0f%% of commits%s",
			bf.N, pluralise("person", "people", bf.N),
			bf.Threshold, warning)),
		"")
	if bf.N <= 2 {
		p("  %s", c(dim, "  → "+strings.Join(bf.Authors, ", ")))
	}
	p("")
}

func buildSparkline(weekly []WeeklySlice, displayWeeks []string) string {
	idx := map[string]int{}
	for _, s := range weekly {
		idx[s.Week] = s.Commits
	}

	maxVal := 0
	for _, wk := range displayWeeks {
		if v := idx[wk]; v > maxVal {
			maxVal = v
		}
	}

	blocks := []string{"·", "▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	var sb strings.Builder
	for _, wk := range displayWeeks {
		v := idx[wk]
		i := 0
		if maxVal > 0 {
			i = int(math.Round(float64(v) / float64(maxVal) * float64(len(blocks)-1)))
		}
		sb.WriteString(blocks[i])
	}
	return sb.String()
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

func pluralise(singular, plural string, n int) string {
	if n == 1 {
		return singular
	}
	return plural
}
