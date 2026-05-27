// Package export writes a stats.Summary to JSON, CSV, or Markdown.
package export

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/yourusername/git-summary/internal/stats"
)

// Format represents an output format.
type Format string

const (
	FormatJSON     Format = "json"
	FormatCSV      Format = "csv"
	FormatMarkdown Format = "md"
)

// ParseFormat converts a string to a Format, returning an error if unknown.
func ParseFormat(s string) (Format, error) {
	switch strings.ToLower(s) {
	case "json":
		return FormatJSON, nil
	case "csv":
		return FormatCSV, nil
	case "md", "markdown":
		return FormatMarkdown, nil
	}
	return "", fmt.Errorf("unknown format %q — use json, csv, or md", s)
}

// Write serialises s into w using the given format.
func Write(w io.Writer, s stats.Summary, fmt Format) error {
	switch fmt {
	case FormatJSON:
		return writeJSON(w, s)
	case FormatCSV:
		return writeCSV(w, s)
	case FormatMarkdown:
		return writeMarkdown(w, s)
	default:
		return writeJSON(w, s)
	}
}

// JSON

type jsonSummary struct {
	GeneratedAt     string              `json:"generated_at"`
	DateRange       string              `json:"date_range"`
	TotalCommits    int                 `json:"total_commits"`
	TotalAdded      int                 `json:"total_added"`
	TotalDeleted    int                 `json:"total_deleted"`
	ActiveDays      int                 `json:"active_days"`
	AvgPerDay       float64             `json:"avg_commits_per_day"`
	MostActiveHour  int                 `json:"most_active_hour"`
	MostActiveDay   string              `json:"most_active_day"`
	Contributors    []jsonContributor   `json:"contributors"`
	TopFiles        []jsonFile          `json:"top_files"`
	WeeklyActivity  []jsonWeek          `json:"weekly_activity"`
	HourlyActivity  [24]int             `json:"hourly_activity"`
}

type jsonContributor struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Commits int    `json:"commits"`
	Added   int    `json:"added"`
	Deleted int    `json:"deleted"`
}

type jsonFile struct {
	Path    string `json:"path"`
	Changes int    `json:"changes"`
}

type jsonWeek struct {
	Week    string `json:"week"`
	Commits int    `json:"commits"`
}

func writeJSON(w io.Writer, s stats.Summary) error {
	js := jsonSummary{
		GeneratedAt:    time.Now().UTC().Format(time.RFC3339),
		DateRange:      s.DateRange,
		TotalCommits:   s.TotalCommits,
		TotalAdded:     s.TotalAdded,
		TotalDeleted:   s.TotalDeleted,
		ActiveDays:     s.ActiveDays,
		AvgPerDay:      s.AvgCommitsPerDay,
		MostActiveHour: s.MostActiveHour,
		MostActiveDay:  s.MostActiveDay,
		HourlyActivity: s.HourlyActivity,
	}
	for _, c := range s.Contributors {
		js.Contributors = append(js.Contributors, jsonContributor{
			Name: c.Name, Email: c.Email,
			Commits: c.Commits, Added: c.Added, Deleted: c.Deleted,
		})
	}
	for _, f := range s.TopFiles {
		js.TopFiles = append(js.TopFiles, jsonFile{Path: f.Path, Changes: f.Changes})
	}
	for _, wk := range s.WeeklyActivity {
		js.WeeklyActivity = append(js.WeeklyActivity, jsonWeek{Week: wk.Label, Commits: wk.Commits})
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(js)
}

// CSV

func writeCSV(w io.Writer, s stats.Summary) error {
	cw := csv.NewWriter(w)

	// Overview section
	_ = cw.Write([]string{"section", "key", "value"})
	_ = cw.Write([]string{"overview", "date_range", s.DateRange})
	_ = cw.Write([]string{"overview", "total_commits", itoa(s.TotalCommits)})
	_ = cw.Write([]string{"overview", "total_added", itoa(s.TotalAdded)})
	_ = cw.Write([]string{"overview", "total_deleted", itoa(s.TotalDeleted)})
	_ = cw.Write([]string{"overview", "active_days", itoa(s.ActiveDays)})
	_ = cw.Write([]string{"overview", "avg_commits_per_day", fmt.Sprintf("%.2f", s.AvgCommitsPerDay)})
	_ = cw.Write([]string{"overview", "most_active_hour", itoa(s.MostActiveHour)})
	_ = cw.Write([]string{"overview", "most_active_day", s.MostActiveDay})
	_ = cw.Write([]string{"", "", ""})

	// Contributors
	_ = cw.Write([]string{"contributor", "name", "email", "commits", "added", "deleted"})
	for _, c := range s.Contributors {
		_ = cw.Write([]string{
			"contributor", c.Name, c.Email,
			itoa(c.Commits), itoa(c.Added), itoa(c.Deleted),
		})
	}
	_ = cw.Write([]string{"", "", ""})

	// Top files
	_ = cw.Write([]string{"file", "path", "changes"})
	for _, f := range s.TopFiles {
		_ = cw.Write([]string{"file", f.Path, itoa(f.Changes)})
	}
	_ = cw.Write([]string{"", "", ""})

	// Weekly
	_ = cw.Write([]string{"week", "label", "commits"})
	for _, wk := range s.WeeklyActivity {
		_ = cw.Write([]string{"week", wk.Label, itoa(wk.Commits)})
	}

	cw.Flush()
	return cw.Error()
}

// Markdown

func writeMarkdown(w io.Writer, s stats.Summary) error {
	p := func(format string, a ...any) {
		fmt.Fprintf(w, format+"\n", a...)
	}

	p("# Git Summary Report")
	p("")
	p("**Period:** %s  ", s.DateRange)
	p("**Generated:** %s", time.Now().Format("Jan 2, 2006 15:04"))
	p("")
	p("## Overview")
	p("")
	p("| Metric | Value |")
	p("|--------|-------|")
	p("| Total commits | %d |", s.TotalCommits)
	p("| Active days | %d |", s.ActiveDays)
	p("| Avg commits/day | %.1f |", s.AvgCommitsPerDay)
	p("| Lines added | +%d |", s.TotalAdded)
	p("| Lines deleted | -%d |", s.TotalDeleted)
	p("| Most active hour | %02d:00 |", s.MostActiveHour)
	p("| Most active day | %s |", s.MostActiveDay)
	p("")
	p("## Top Contributors")
	p("")
	p("| # | Name | Commits | Added | Deleted |")
	p("|---|------|---------|-------|---------|")
	for i, c := range s.Contributors {
		p("| %d | %s | %d | +%d | -%d |", i+1, c.Name, c.Commits, c.Added, c.Deleted)
	}
	p("")
	p("## Most Changed Files")
	p("")
	p("| File | Changes |")
	p("|------|---------|")
	for _, f := range s.TopFiles {
		p("| `%s` | %d |", f.Path, f.Changes)
	}
	p("")
	p("## Weekly Activity")
	p("")
	p("| Week | Commits |")
	p("|------|---------|")
	for _, wk := range s.WeeklyActivity {
		p("| %s | %d |", wk.Label, wk.Commits)
	}
	p("")
	p("---")
	p("_Generated by [git-summary](https://github.com/yourusername/git-summary)_")
	return nil
}

func itoa(n int) string { return fmt.Sprintf("%d", n) }
