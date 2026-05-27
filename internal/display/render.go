// Package display renders a stats.Summary to the terminal.
package display

import (
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/yourusername/git-summary/internal/stats"
)

// Options controls rendering behavior.
type Options struct {
	NoColor bool
	TopN    int
}

// ANSI color codes
const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	dim    = "\033[2m"
	cyan   = "\033[36m"
	green  = "\033[32m"
	yellow = "\033[33m"
	blue   = "\033[34m"
	purple = "\033[35m"
	red    = "\033[31m"
	white  = "\033[37m"
)

var noColor bool

func color(code, text string) string {
	if noColor {
		return text
	}
	return code + text + reset
}

// Render prints the full summary report.
func Render(s stats.Summary, opts Options) {
	noColor = opts.NoColor || !isTerminal()

	printHeader(s)
	printOverview(s)
	printContributors(s)
	printTopFiles(s)
	printHeatmap(s)
	printWeeklyChart(s)
}

func printHeader(s stats.Summary) {
	width := 60
	line := strings.Repeat("─", width)
	fmt.Println(color(bold+cyan, "╭"+line+"╮"))
	title := "  git-summary"
	pad := width - len(title) - 1
	fmt.Println(color(bold+cyan, "│") + color(bold+white, title) + strings.Repeat(" ", pad) + color(bold+cyan, "│"))
	sub := "  " + s.DateRange
	pad = width - len(sub) - 1
	if pad < 0 {
		pad = 0
	}
	fmt.Println(color(bold+cyan, "│") + color(dim, sub) + strings.Repeat(" ", pad) + color(bold+cyan, "│"))
	fmt.Println(color(bold+cyan, "╰"+line+"╯"))
	fmt.Println()
}

func printOverview(s stats.Summary) {
	fmt.Println(color(bold+yellow, "  Overview"))
	fmt.Println(color(dim, "  "+strings.Repeat("─", 40)))

	printKV("  Total commits", fmt.Sprintf("%d", s.TotalCommits), green)
	printKV("  Active days", fmt.Sprintf("%d", s.ActiveDays), green)
	printKV("  Avg commits/day", fmt.Sprintf("%.1f", s.AvgCommitsPerDay), green)
	printKV("  Lines added", fmt.Sprintf("+%d", s.TotalAdded), green)
	printKV("  Lines deleted", fmt.Sprintf("-%d", s.TotalDeleted), red)
	printKV("  Most active hour", fmt.Sprintf("%02d:00", s.MostActiveHour), yellow)
	printKV("  Most active day", s.MostActiveDay, yellow)
	fmt.Println()
}

func printKV(label, value, valueColor string) {
	lPad := 26 - len(label)
	if lPad < 1 {
		lPad = 1
	}
	fmt.Printf("%s%s%s\n",
		color(dim, label),
		strings.Repeat(" ", lPad),
		color(valueColor, value),
	)
}

func printContributors(s stats.Summary) {
	if len(s.Contributors) == 0 {
		return
	}
	fmt.Println(color(bold+yellow, "  Top Contributors"))
	fmt.Println(color(dim, "  "+strings.Repeat("─", 40)))

	maxCommits := s.Contributors[0].Commits
	barWidth := 20

	for i, c := range s.Contributors {
		rank := fmt.Sprintf("  %2d.", i+1)
		name := truncate(c.Name, 20)
		namePad := 20 - len(name)

		barLen := 0
		if maxCommits > 0 {
			barLen = int(float64(c.Commits) / float64(maxCommits) * float64(barWidth))
		}
		bar := strings.Repeat("█", barLen) + strings.Repeat("░", barWidth-barLen)

		commits := fmt.Sprintf("%4d commits", c.Commits)
		diff := fmt.Sprintf("+%d/-%d", c.Added, c.Deleted)

		fmt.Printf("%s %s%s %s %s %s\n",
			color(dim, rank),
			color(bold, name),
			strings.Repeat(" ", namePad),
			color(blue, bar),
			color(green, commits),
			color(dim, diff),
		)
	}
	fmt.Println()
}

func printTopFiles(s stats.Summary) {
	if len(s.TopFiles) == 0 {
		return
	}
	fmt.Println(color(bold+yellow, "  Most Changed Files"))
	fmt.Println(color(dim, "  "+strings.Repeat("─", 40)))

	maxChanges := s.TopFiles[0].Changes
	barWidth := 15

	for _, f := range s.TopFiles {
		barLen := 0
		if maxChanges > 0 {
			barLen = int(float64(f.Changes) / float64(maxChanges) * float64(barWidth))
		}
		bar := strings.Repeat("▪", barLen) + strings.Repeat("·", barWidth-barLen)
		name := truncate(f.Path, 36)
		namePad := 36 - len(name)

		fmt.Printf("  %s%s %s %s\n",
			color(white, name),
			strings.Repeat(" ", namePad),
			color(purple, bar),
			color(dim, fmt.Sprintf("%d changes", f.Changes)),
		)
	}
	fmt.Println()
}

func printHeatmap(s stats.Summary) {
	fmt.Println(color(bold+yellow, "  Commit Activity by Hour"))
	fmt.Println(color(dim, "  "+strings.Repeat("─", 40)))

	maxVal := 0
	for _, v := range s.HourlyActivity {
		if v > maxVal {
			maxVal = v
		}
	}

	blocks := []string{"░", "▒", "▓", "█"}

	fmt.Print("  ")
	for h := 0; h < 24; h++ {
		v := s.HourlyActivity[h]
		idx := 0
		if maxVal > 0 {
			idx = int(math.Round(float64(v) / float64(maxVal) * float64(len(blocks)-1)))
		}
		if v == 0 {
			fmt.Print(color(dim, "·"))
		} else if h == s.MostActiveHour {
			fmt.Print(color(yellow, blocks[idx]))
		} else {
			fmt.Print(color(cyan, blocks[idx]))
		}
	}
	fmt.Println()

	// Hour labels
	fmt.Print("  ")
	for h := 0; h < 24; h += 6 {
		label := fmt.Sprintf("%-6s", fmt.Sprintf("%02d:00", h))
		fmt.Print(color(dim, label))
	}
	fmt.Println()
	fmt.Println()
}

func printWeeklyChart(s stats.Summary) {
	if len(s.WeeklyActivity) == 0 {
		return
	}

	// Show last 12 weeks max
	data := s.WeeklyActivity
	if len(data) > 12 {
		data = data[len(data)-12:]
	}

	maxVal := 0
	for _, w := range data {
		if w.Commits > maxVal {
			maxVal = w.Commits
		}
	}

	fmt.Println(color(bold+yellow, "  Weekly Activity"))
	fmt.Println(color(dim, "  "+strings.Repeat("─", 40)))

	chartHeight := 5
	for row := chartHeight; row >= 1; row-- {
		threshold := float64(maxVal) * float64(row) / float64(chartHeight)
		fmt.Print("  ")
		for _, w := range data {
			if float64(w.Commits) >= threshold {
				fmt.Print(color(cyan, "▆ "))
			} else {
				fmt.Print(color(dim, "  "))
			}
		}
		fmt.Println()
	}

	// Week labels (show every other)
	fmt.Print("  ")
	for i, w := range data {
		label := w.Label[6:]
		if i%2 == 0 {
			fmt.Print(color(dim, label+" "))
		} else {
			fmt.Print("    ")
		}
	}
	fmt.Println()
	fmt.Println()
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
