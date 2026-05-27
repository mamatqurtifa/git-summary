// Package stats computes derived metrics from a list of commits.
package stats

import (
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/mamatqurtifa/git-summary/internal/gitlog"
)

// Contributor holds per-author activity.
type Contributor struct {
	Name    string
	Email   string
	Commits int
	Added   int
	Deleted int
}

// FileActivity holds activity for a single file.
type FileActivity struct {
	Path    string
	Changes int
}

// WeeklyActivity holds commit count per week bucket.
type WeeklyActivity struct {
	Label   string
	Commits int
}

// HourlyActivity holds commit count per hour of day.
type HourlyActivity [24]int

// Summary is the full computed report.
type Summary struct {
	TotalCommits    int
	TotalAdded      int
	TotalDeleted    int
	ActiveDays      int
	DateRange       string
	Contributors    []Contributor
	TopFiles        []FileActivity
	WeeklyActivity  []WeeklyActivity
	HourlyActivity  HourlyActivity
	MostActiveHour  int
	MostActiveDay   string
	AvgCommitsPerDay float64
}

// Compute derives a Summary from a slice of commits.
func Compute(commits []gitlog.Commit, topN int) Summary {
	if len(commits) == 0 {
		return Summary{}
	}

	s := Summary{}
	authorMap := map[string]*Contributor{}
	fileMap := map[string]int{}
	weekMap := map[string]int{}
	daySet := map[string]bool{}
	dayCount := map[string]int{}

	var earliest, latest time.Time
	earliest = commits[0].Timestamp
	latest = commits[0].Timestamp

	for _, c := range commits {
		s.TotalCommits++
		s.TotalAdded += c.Added
		s.TotalDeleted += c.Deleted

		// Track time range
		if c.Timestamp.Before(earliest) {
			earliest = c.Timestamp
		}
		if c.Timestamp.After(latest) {
			latest = c.Timestamp
		}

		// Contributors
		key := c.Email
		if _, ok := authorMap[key]; !ok {
			authorMap[key] = &Contributor{Name: c.Author, Email: c.Email}
		}
		authorMap[key].Commits++
		authorMap[key].Added += c.Added
		authorMap[key].Deleted += c.Deleted

		// Files
		for _, f := range c.Files {
			fileMap[f]++
		}

		// Weekly
		y, w := c.Timestamp.ISOWeek()
		wk := formatWeek(y, w)
		weekMap[wk]++

		// Hourly
		s.HourlyActivity[c.Timestamp.Hour()]++

		// Active days
		day := c.Timestamp.Format("2006-01-02")
		daySet[day] = true
		dayCount[c.Timestamp.Weekday().String()]++
	}

	s.ActiveDays = len(daySet)
	s.DateRange = earliest.Format("Jan 2, 2006") + " → " + latest.Format("Jan 2, 2006")

	if s.ActiveDays > 0 {
		s.AvgCommitsPerDay = float64(s.TotalCommits) / float64(s.ActiveDays)
	}

	// Sort contributors
	for _, v := range authorMap {
		s.Contributors = append(s.Contributors, *v)
	}
	sort.Slice(s.Contributors, func(i, j int) bool {
		return s.Contributors[i].Commits > s.Contributors[j].Commits
	})
	if len(s.Contributors) > topN {
		s.Contributors = s.Contributors[:topN]
	}

	// Top files
	type filePair struct {
		path    string
		changes int
	}
	var files []filePair
	for p, c := range fileMap {
		files = append(files, filePair{p, c})
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].changes > files[j].changes
	})
	max := 10
	if len(files) < max {
		max = len(files)
	}
	for _, f := range files[:max] {
		name := shortenPath(f.path)
		s.TopFiles = append(s.TopFiles, FileActivity{Path: name, Changes: f.changes})
	}

	// Weekly activity sorted by week label
	for wk, cnt := range weekMap {
		s.WeeklyActivity = append(s.WeeklyActivity, WeeklyActivity{Label: wk, Commits: cnt})
	}
	sort.Slice(s.WeeklyActivity, func(i, j int) bool {
		return s.WeeklyActivity[i].Label < s.WeeklyActivity[j].Label
	})

	// Most active hour
	maxH := 0
	for h, cnt := range s.HourlyActivity {
		if cnt > s.HourlyActivity[maxH] {
			maxH = h
		}
	}
	s.MostActiveHour = maxH

	// Most active weekday
	maxDay := ""
	maxDayCnt := 0
	for d, cnt := range dayCount {
		if cnt > maxDayCnt {
			maxDayCnt = cnt
			maxDay = d
		}
	}
	s.MostActiveDay = maxDay

	return s
}

func formatWeek(year, week int) string {
	return strings.Join([]string{
		padInt(year, 4), "-W", padInt(week, 2),
	}, "")
}

func padInt(n, width int) string {
	s := string(rune('0'+n%10))
	for n >= 10 {
		n /= 10
		s = string(rune('0'+n%10)) + s
	}
	for len(s) < width {
		s = "0" + s
	}
	return s
}

func shortenPath(p string) string {
	parts := strings.Split(filepath.ToSlash(p), "/")
	if len(parts) <= 3 {
		return p
	}
	return ".../" + strings.Join(parts[len(parts)-2:], "/")
}
