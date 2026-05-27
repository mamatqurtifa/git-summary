// Package workpattern analyses commit timestamps to surface work habit signals.
package workpattern

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/mamatqurtifa/git-summary/internal/gitlog"
)

// WorkHours defines the "normal" working window (24-hour).
type WorkHours struct {
	Start int
	End   int
}

var defaultHours = WorkHours{Start: 9, End: 18}

// AuthorPattern holds work pattern signals for one contributor.
type AuthorPattern struct {
	Name           string
	TotalCommits   int
	WeekendPct     float64
	AfterHoursPct  float64
	EarlyBirdPct   float64
	NightOwlPct    float64
	PeakHour       int
	PeakDay        string
	Signals        []string
}

// Result is the full work pattern report.
type Result struct {
	Authors      []AuthorPattern
	TeamWeekend  float64
	TeamAfterHrs float64
	TeamSignals  []string
}

// Compute derives work patterns from commits.
func Compute(commits []gitlog.Commit, hours WorkHours, topN int) Result {
	if hours.Start == 0 && hours.End == 0 {
		hours = defaultHours
	}

	type authorKey = string
	type commitList = []gitlog.Commit
	byAuthor := map[authorKey]commitList{}
	for _, c := range commits {
		byAuthor[c.Email] = append(byAuthor[c.Email], c)
	}

	var patterns []AuthorPattern
	for _, cs := range byAuthor {
		if len(cs) == 0 {
			continue
		}
		ap := analyseAuthor(cs[0].Author, cs, hours)
		patterns = append(patterns, ap)
	}

	sort.Slice(patterns, func(i, j int) bool {
		return patterns[i].TotalCommits > patterns[j].TotalCommits
	})
	if len(patterns) > topN {
		patterns = patterns[:topN]
	}

	// Team-level stats
	totalCommits := len(commits)
	weekendCount, afterHrsCount := 0, 0
	for _, c := range commits {
		wd := c.Timestamp.Weekday()
		if wd == time.Saturday || wd == time.Sunday {
			weekendCount++
		}
		h := c.Timestamp.Hour()
		if h < hours.Start || h >= hours.End {
			afterHrsCount++
		}
	}

	res := Result{Authors: patterns}
	if totalCommits > 0 {
		res.TeamWeekend = float64(weekendCount) / float64(totalCommits) * 100
		res.TeamAfterHrs = float64(afterHrsCount) / float64(totalCommits) * 100
	}
	res.TeamSignals = teamSignals(res.TeamWeekend, res.TeamAfterHrs)
	return res
}

func analyseAuthor(name string, commits []gitlog.Commit, hours WorkHours) AuthorPattern {
	total := len(commits)
	weekend, afterHrs, earlyBird, nightOwl := 0, 0, 0, 0
	hourCount := [24]int{}
	dayCount := map[string]int{}

	for _, c := range commits {
		h := c.Timestamp.Hour()
		wd := c.Timestamp.Weekday()
		hourCount[h]++
		dayCount[wd.String()]++

		if wd == time.Saturday || wd == time.Sunday {
			weekend++
		}
		if h < hours.Start || h >= hours.End {
			afterHrs++
		}
		if h < 7 {
			earlyBird++
		}
		if h >= 22 {
			nightOwl++
		}
	}

	// Peak hour
	peakH := 0
	for h, v := range hourCount {
		if v > hourCount[peakH] {
			peakH = h
		}
	}
	// Peak day
	peakDay := ""
	peakDayCnt := 0
	for d, v := range dayCount {
		if v > peakDayCnt {
			peakDayCnt = v
			peakDay = d
		}
	}

	ap := AuthorPattern{
		Name:          name,
		TotalCommits:  total,
		WeekendPct:    pct(weekend, total),
		AfterHoursPct: pct(afterHrs, total),
		EarlyBirdPct:  pct(earlyBird, total),
		NightOwlPct:   pct(nightOwl, total),
		PeakHour:      peakH,
		PeakDay:       peakDay,
	}
	ap.Signals = authorSignals(ap)
	return ap
}

func pct(n, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(n) / float64(total) * 100
}

func authorSignals(ap AuthorPattern) []string {
	var s []string
	if ap.WeekendPct >= 30 {
		s = append(s, fmt.Sprintf("%.0f%% weekend commits — possible burnout", ap.WeekendPct))
	}
	if ap.NightOwlPct >= 25 {
		s = append(s, fmt.Sprintf("%.0f%% commits after 22:00 — late-night pattern", ap.NightOwlPct))
	}
	if ap.AfterHoursPct >= 50 {
		s = append(s, fmt.Sprintf("%.0f%% outside working hours", ap.AfterHoursPct))
	}
	if ap.EarlyBirdPct >= 20 {
		s = append(s, fmt.Sprintf("%.0f%% commits before 07:00 — early bird", ap.EarlyBirdPct))
	}
	return s
}

func teamSignals(weekendPct, afterHrsPct float64) []string {
	var s []string
	if weekendPct >= 25 {
		s = append(s, fmt.Sprintf("Team: %.0f%% of all commits happen on weekends — review workload", weekendPct))
	}
	if afterHrsPct >= 40 {
		s = append(s, fmt.Sprintf("Team: %.0f%% of commits outside work hours — timezone spread or overtime", afterHrsPct))
	}
	return s
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

// Render prints the work pattern report to w.
func Render(w io.Writer, r Result, noColor bool) {
	p := func(f string, a ...any) { fmt.Fprintf(w, f+"\n", a...) }
	c := func(code, s string) string {
		if noColor {
			return s
		}
		return col(code, s)
	}

	p("")
	p(c(bold+cyan, "  Work Pattern Analysis"))
	p(c(dim, "  "+strings.Repeat("─", 65)))

	// Team-level overview
	p("  %s  weekend: %s   after-hours: %s",
		c(dim, "team →"),
		colorPct(r.TeamWeekend, 15, 30, noColor),
		colorPct(r.TeamAfterHrs, 30, 50, noColor),
	)
	for _, sig := range r.TeamSignals {
		p("  %s %s", c(yellow, "⚠"), c(yellow, sig))
	}
	p("")

	// Per-author table
	p("  %-20s %6s %8s %10s %9s  %s",
		c(dim, "author"),
		c(dim, "commits"),
		c(dim, "weekend"),
		c(dim, "after-hrs"),
		c(dim, "peak"),
		c(dim, "signals"),
	)
	p(c(dim, "  "+strings.Repeat("─", 65)))

	for _, ap := range r.Authors {
		sigStr := ""
		if len(ap.Signals) > 0 {
			sigStr = "⚠ " + ap.Signals[0]
		}

		p("  %-20s %6d %7.0f%% %9.0f%%  %02d:00 %-3s  %s",
			c(bold, truncate(ap.Name, 20)),
			ap.TotalCommits,
			ap.WeekendPct,
			ap.AfterHoursPct,
			ap.PeakHour,
			shortDay(ap.PeakDay),
			c(yellow, sigStr),
		)

		// Extra signals
		for i, sig := range ap.Signals {
			if i == 0 {
				continue
			}
			p("  %s %s", strings.Repeat(" ", 50), c(dim, "⚠ "+sig))
		}
	}
	p("")
}

func colorPct(v, warnAt, critAt float64, noColor bool) string {
	s := fmt.Sprintf("%.0f%%", v)
	if noColor {
		return s
	}
	if v >= critAt {
		return col(red, s)
	}
	if v >= warnAt {
		return col(yellow, s)
	}
	return col(green, s)
}

func shortDay(day string) string {
	if len(day) >= 3 {
		return day[:3]
	}
	return day
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
