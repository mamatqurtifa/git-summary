// Package prompt provides interactive terminal prompts.
package prompt

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Option struct {
	Label string
	Value string
}

var timeRanges = []Option{
	{"Last 24 hours",  "24 hours ago"},
	{"Last week",      "1 week ago"},
	{"Last 2 weeks",   "2 weeks ago"},
	{"Last month",     "1 month ago"},
	{"Last 3 months",  "3 months ago"},
	{"Last 6 months",  "6 months ago"},
	{"Last year",      "1 year ago"},
	{"All time",       ""},
}

const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	cyan   = "\033[36m"
	yellow = "\033[33m"
	dim    = "\033[2m"
	green  = "\033[32m"
	up     = "\033[%dA"
	clrLine = "\033[2K"
)

func c(code, s string) string { return code + s + reset }

// PickRange shows an interactive time range selector.
// Returns the chosen "since" value and true, or "", false if cancelled.
func PickRange() (string, bool) {
	if !isTerminal() {
		return "1 month ago", true
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Println(c(bold+cyan, "  git-summary") + c(dim, " — select time range"))
	fmt.Println(c(dim, "  Use number keys or arrow keys, then Enter"))
	fmt.Println()

	selected := 1 // default: "Last week"

	renderMenu(selected)

	fmt.Print(c(dim, "  Enter number (default 2): "))
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)

	if line == "" {
	} else if line == "q" || line == "Q" {
		return "", false
	} else {
		n, err := strconv.Atoi(line)
		if err != nil || n < 1 || n > len(timeRanges) {
			fmt.Fprintf(os.Stderr, "  Invalid choice, using default (Last week)\n")
		} else {
			selected = n
		}
	}

	chosen := timeRanges[selected-1]
	fmt.Printf(c(green, "  ✓ Showing: %s\n"), chosen.Label)
	fmt.Println()

	return chosen.Value, true
}

func renderMenu(selected int) {
	for i, opt := range timeRanges {
		num := fmt.Sprintf("  %d.", i+1)
		if i+1 == selected {
			fmt.Printf("%s %s\n",
				c(dim, num),
				c(bold+yellow, "▶ "+opt.Label),
			)
		} else {
			fmt.Printf("%s %s\n",
				c(dim, num),
				c(dim, "  "+opt.Label),
			)
		}
	}
	fmt.Println()
}

func isTerminal() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
