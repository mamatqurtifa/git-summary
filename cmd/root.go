package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/mamatqurtifa/git-summary/internal/compare"
	"github.com/mamatqurtifa/git-summary/internal/dirmap"
	"github.com/mamatqurtifa/git-summary/internal/display"
	"github.com/mamatqurtifa/git-summary/internal/export"
	"github.com/mamatqurtifa/git-summary/internal/gitlog"
	"github.com/mamatqurtifa/git-summary/internal/prompt"
	"github.com/mamatqurtifa/git-summary/internal/stats"
	"github.com/mamatqurtifa/git-summary/internal/tagdiff"
	"github.com/mamatqurtifa/git-summary/internal/trend"
	"github.com/mamatqurtifa/git-summary/internal/workpattern"
)

const version = "0.2.0"

func Execute() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "completion":
			runCompletion()
			return
		case "version", "--version", "-version":
			fmt.Printf("git-summary v%s\n", version)
			return
		case "help", "--help", "-help", "-h":
			printHelp()
			return
		case "changelog":
			runChangelog()
			return
		case "compare":
			runCompare()
			return
		}
	}

	// main flags
	fs := flag.NewFlagSet("git-summary", flag.ExitOnError)
	var (
		since    = fs.String("since", "", "Start date (e.g. '1 week ago', '2024-01-01')")
		until    = fs.String("until", "", "End date")
		author   = fs.String("author", "", "Filter by author name or email")
		branch   = fs.String("branch", "", "Branch to analyze (default: current)")
		topN     = fs.Int("top", 10, "Number of top contributors to show")
		noColor  = fs.Bool("no-color", false, "Disable colored output")
		format   = fs.String("format", "", "Export format: json, csv, md (prints to stdout)")
		output   = fs.String("output", "", "Write export to file instead of stdout")
		showTrend   = fs.Bool("trends", false, "Show contributor trends & bus factor")
		showDirmap  = fs.Bool("dirs", false, "Show directory ownership breakdown")
		showWork    = fs.Bool("workpattern", false, "Show work pattern analysis")
		workStart   = fs.Int("work-start", 9, "Work hours start (24h, for workpattern)")
		workEnd     = fs.Int("work-end", 18, "Work hours end (24h, for workpattern)")
		showAll     = fs.Bool("all", false, "Show all sections (trends + dirs + workpattern)")
	)

	fs.Usage = printHelp
	_ = fs.Parse(os.Args[1:])

	repoPath := "."
	if fs.NArg() > 0 {
		repoPath = fs.Arg(0)
	}

	if *since == "" {
		chosen, ok := prompt.PickRange()
		if !ok {
			fmt.Println("Cancelled.")
			return
		}
		*since = chosen
	}

	opts := gitlog.Options{
		RepoPath: repoPath,
		Since:    *since,
		Until:    *until,
		Author:   *author,
		Branch:   *branch,
	}

	commits, err := gitlog.Parse(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if len(commits) == 0 {
		fmt.Println("No commits found for the given filters.")
		return
	}

	summary := stats.Compute(commits, *topN)

	// Export mode
	if *format != "" {
		f, err := export.ParseFormat(*format)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		w := os.Stdout
		if *output != "" {
			fh, err := os.Create(*output)
			if err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				os.Exit(1)
			}
			defer fh.Close()
			w = fh
			fmt.Fprintf(os.Stderr, "Writing %s report to %s...\n", *format, *output)
		}
		if err := export.Write(w, summary, f); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		return
	}

	// Terminal display
	display.Render(summary, display.Options{
		NoColor: *noColor,
		TopN:    *topN,
	})

	if *showTrend || *showAll {
		tr := trend.Compute(commits, *topN)
		trend.Render(os.Stdout, tr, *noColor)
	}

	if *showDirmap || *showAll {
		dr := dirmap.Compute(commits, 2)
		dirmap.Render(os.Stdout, dr, *noColor)
	}

	if *showWork || *showAll {
		wr := workpattern.Compute(commits, workpattern.WorkHours{
			Start: *workStart,
			End:   *workEnd,
		}, *topN)
		workpattern.Render(os.Stdout, wr, *noColor)
	}
}

// Subcommand: compare

func runCompare() {
	fs := flag.NewFlagSet("git-summary compare", flag.ExitOnError)
	since   := fs.String("since", "1 month ago", "Time range")
	topN    := fs.Int("top", 10, "Top N contributors")
	noColor := fs.Bool("no-color", false, "Disable color")
	_ = fs.Parse(os.Args[2:])

	args := fs.Args()
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: git-summary compare <branchA> <branchB> [repo-path]")
		os.Exit(1)
	}
	branchA, branchB := args[0], args[1]
	repoPath := "."
	if len(args) >= 3 {
		repoPath = args[2]
	}

	res, err := compare.Run(repoPath, branchA, branchB, *since, *topN)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	compare.Render(os.Stdout, res, *noColor)
}

// Subcommand: changelog

func runChangelog() {
	fs := flag.NewFlagSet("git-summary changelog", flag.ExitOnError)
	from    := fs.String("from", "", "Start tag/ref (required)")
	to      := fs.String("to", "HEAD", "End tag/ref (default: HEAD)")
	md      := fs.Bool("md", false, "Output as markdown")
	output  := fs.String("output", "", "Write to file")
	noColor := fs.Bool("no-color", false, "Disable color")
	list    := fs.Bool("list-tags", false, "List available tags and exit")
	_ = fs.Parse(os.Args[2:])

	repoPath := "."
	if fs.NArg() > 0 {
		repoPath = fs.Arg(0)
	}

	if *list {
		tags, err := tagdiff.ListTags(repoPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		fmt.Println("Available tags:")
		for _, t := range tags {
			fmt.Println(" ", t)
		}
		return
	}

	if *from == "" {
		tags, err := tagdiff.ListTags(repoPath)
		if err != nil || len(tags) < 2 {
			fmt.Fprintln(os.Stderr, "error: --from is required (or create at least 2 git tags)")
			fmt.Fprintln(os.Stderr, "       use --list-tags to see available tags")
			os.Exit(1)
		}
		*to = tags[0]
		*from = tags[1]
		fmt.Fprintf(os.Stderr, "Auto-detected: %s → %s\n", *from, *to)
	}

	res, err := tagdiff.Run(repoPath, *from, *to)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	w := os.Stdout
	if *output != "" {
		fh, err := os.Create(*output)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		defer fh.Close()
		w = fh
		fmt.Fprintf(os.Stderr, "Writing changelog to %s...\n", *output)
	}

	if *md {
		tagdiff.RenderMarkdown(w, res)
	} else {
		tagdiff.RenderTerminal(w, res, *noColor)
	}
}

// Help

func printHelp() {
	fmt.Print(`git-summary — Git repository activity at a glance

Usage:
  git-summary [options] [path]
  git-summary compare <branchA> <branchB> [path] [options]
  git-summary changelog [options] [path]
  git-summary completion <bash|zsh|fish>
  git-summary version

Main options:
  -since string      Start date (default: interactive picker)
                     e.g. "1 week ago", "1 month ago", "2024-01-01"
  -until string      End date (default: now)
  -author string     Filter by author (partial name or email)
  -branch string     Branch to analyze (default: current)
  -top int           Top N contributors (default: 10)
  -no-color          Disable colored output

Extra sections:
  -trends            Show contributor trends & bus factor
  -dirs              Show directory ownership breakdown
  -workpattern       Show work pattern analysis (overtime, weekends)
  -all               Show all extra sections
  -work-start int    Work hours start for analysis (default: 9)
  -work-end int      Work hours end for analysis (default: 18)

Export:
  -format string     Export format: json, csv, md
  -output string     Write export to file (default: stdout)

Compare subcommand:
  git-summary compare main develop
  git-summary compare main develop --since "1 month ago"

Changelog subcommand:
  git-summary changelog --from v1.0.0 --to v1.1.0
  git-summary changelog                          # auto-detects last 2 tags
  git-summary changelog --from v1.0.0 --md       # markdown output
  git-summary changelog --list-tags              # list available tags

Shell completion:
  source <(git-summary completion bash)          # bash
  source <(git-summary completion zsh)           # zsh
  git-summary completion fish | ...              # fish

Examples:
  git-summary
  git-summary --since "1 week ago" --all
  git-summary --format json --output report.json
  git-summary --format md --output REPORT.md
  git-summary compare main develop --since "2 weeks ago"
  git-summary changelog --from v1.0 --to v2.0 --md --output CHANGELOG.md

`)
}

// Shell completion

func runCompletion() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: git-summary completion <bash|zsh|fish>")
		os.Exit(1)
	}
	switch os.Args[2] {
	case "bash":
		fmt.Print(bashCompletion)
	case "zsh":
		fmt.Print(zshCompletion)
	case "fish":
		fmt.Print(fishCompletion)
	default:
		fmt.Fprintf(os.Stderr, "unknown shell %q\n", os.Args[2])
		os.Exit(1)
	}
}

const bashCompletion = `# bash completion for git-summary
# source <(git-summary completion bash)

_git_summary_completions() {
    local cur prev words
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    local subcommands="compare changelog completion version help"
    local main_opts="-since -until -author -branch -top -no-color -format -output -trends -dirs -workpattern -all -work-start -work-end"

    # Subcommand in position 1
    if [[ ${COMP_CWORD} -eq 1 ]]; then
        COMPREPLY=( $(compgen -W "$subcommands $main_opts" -- "$cur") )
        return 0
    fi

    case "${COMP_WORDS[1]}" in
        compare)
            local branches
            branches=$(git branch 2>/dev/null | sed 's/\* //' | tr -d ' ')
            COMPREPLY=( $(compgen -W "$branches -since -no-color -top" -- "$cur") )
            return 0 ;;
        changelog)
            case "$prev" in
                --from|-from)
                    local tags; tags=$(git tag 2>/dev/null)
                    COMPREPLY=( $(compgen -W "$tags" -- "$cur") ); return 0 ;;
                --to|-to)
                    local tags; tags=$(git tag 2>/dev/null)
                    COMPREPLY=( $(compgen -W "HEAD $tags" -- "$cur") ); return 0 ;;
            esac
            COMPREPLY=( $(compgen -W "-from -to -md -output -list-tags -no-color" -- "$cur") )
            return 0 ;;
        completion)
            COMPREPLY=( $(compgen -W "bash zsh fish" -- "$cur") ); return 0 ;;
    esac

    case "$prev" in
        -since)  COMPREPLY=( $(compgen -W "today yesterday '1 week ago' '2 weeks ago' '1 month ago' '3 months ago' '6 months ago' '1 year ago'" -- "$cur") ); return 0 ;;
        -until)  COMPREPLY=( $(compgen -W "today yesterday '1 week ago' '1 month ago'" -- "$cur") ); return 0 ;;
        -branch) branches=$(git branch 2>/dev/null | sed 's/\* //' | tr -d ' '); COMPREPLY=( $(compgen -W "$branches" -- "$cur") ); return 0 ;;
        -top)    COMPREPLY=( $(compgen -W "5 10 20 50" -- "$cur") ); return 0 ;;
        -format) COMPREPLY=( $(compgen -W "json csv md" -- "$cur") ); return 0 ;;
        -work-start|-work-end) COMPREPLY=( $(compgen -W "6 7 8 9 10 17 18 19 20" -- "$cur") ); return 0 ;;
    esac

    if [[ "$cur" == -* ]]; then
        COMPREPLY=( $(compgen -W "$main_opts" -- "$cur") )
    else
        COMPREPLY=( $(compgen -d -- "$cur") )
    fi
}

complete -F _git_summary_completions git-summary
`

const zshCompletion = `#compdef git-summary
# source <(git-summary completion zsh)

_git_summary() {
    local state line
    typeset -A opt_args

    _arguments \
        '1: :->subcmd' \
        '*: :->args'

    case $state in
        subcmd)
            _describe 'subcommand' '(
                compare:"compare two branches side by side"
                changelog:"generate changelog between tags"
                completion:"print shell completion script"
                version:"print version"
            )' ;;
        args)
            case $line[1] in
                compare)
                    _arguments \
                        '-since[Time range]:date:(today "1 week ago" "1 month ago")' \
                        '-top[Top N]:number:(5 10 20)' \
                        '-no-color[Disable color]' \
                        '1:branchA:->branches' \
                        '2:branchB:->branches'
                    if [[ $state == branches ]]; then
                        local branches
                        branches=(${(f)"$(git branch 2>/dev/null | sed 's/\* //' | tr -d ' ')"})
                        _describe 'branch' branches
                    fi ;;
                changelog)
                    _arguments \
                        '-from[Start tag]:tag:($(git tag 2>/dev/null))' \
                        '-to[End tag]:tag:(HEAD $(git tag 2>/dev/null))' \
                        '-md[Markdown output]' \
                        '-output[Output file]:file:_files' \
                        '-list-tags[List tags]' ;;
                completion)
                    _describe 'shell' '(bash zsh fish)' ;;
                *)
                    _arguments \
                        '-since[Start date]:date:(today yesterday "1 week ago" "1 month ago" "3 months ago" "1 year ago")' \
                        '-until[End date]:date:(today yesterday "1 week ago")' \
                        '-author[Filter author]:author' \
                        '-branch[Branch]:branch:($(git branch 2>/dev/null | sed "s/\* //" | tr -d " "))' \
                        '-top[Top N]:number:(5 10 20 50)' \
                        '-format[Export format]:fmt:(json csv md)' \
                        '-output[Output file]:file:_files' \
                        '-trends[Show trends]' \
                        '-dirs[Show dir breakdown]' \
                        '-workpattern[Show work patterns]' \
                        '-all[Show all sections]' \
                        '-no-color[Disable color]' \
                        '1:repo:_directories' ;;
            esac ;;
    esac
}

compdef _git_summary git-summary
`

const fishCompletion = `# fish completion for git-summary
# git-summary completion fish > ~/.config/fish/completions/git-summary.fish

# Main flags
complete -c git-summary -n '__fish_use_subcommand' -l since       -d 'Start date'         -x -a "today yesterday '1 week ago' '2 weeks ago' '1 month ago' '3 months ago' '1 year ago'"
complete -c git-summary -n '__fish_use_subcommand' -l until       -d 'End date'           -x -a "today yesterday '1 week ago'"
complete -c git-summary -n '__fish_use_subcommand' -l author      -d 'Filter by author'
complete -c git-summary -n '__fish_use_subcommand' -l branch      -d 'Branch'             -x -a "(git branch 2>/dev/null | string trim | string replace -r '^\* ' '')"
complete -c git-summary -n '__fish_use_subcommand' -l top         -d 'Top N'              -x -a "5 10 20 50"
complete -c git-summary -n '__fish_use_subcommand' -l format      -d 'Export format'      -x -a "json csv md"
complete -c git-summary -n '__fish_use_subcommand' -l output      -d 'Output file'
complete -c git-summary -n '__fish_use_subcommand' -l trends      -d 'Show trends'
complete -c git-summary -n '__fish_use_subcommand' -l dirs        -d 'Show dir breakdown'
complete -c git-summary -n '__fish_use_subcommand' -l workpattern -d 'Show work patterns'
complete -c git-summary -n '__fish_use_subcommand' -l all         -d 'Show all sections'
complete -c git-summary -n '__fish_use_subcommand' -l no-color    -d 'Disable color'

# Subcommands
complete -c git-summary -f -n '__fish_use_subcommand' -a compare    -d 'Compare two branches'
complete -c git-summary -f -n '__fish_use_subcommand' -a changelog  -d 'Generate changelog'
complete -c git-summary -f -n '__fish_use_subcommand' -a completion -d 'Print completion script'
complete -c git-summary -f -n '__fish_use_subcommand' -a version    -d 'Print version'

# changelog flags
complete -c git-summary -n '__fish_seen_subcommand_from changelog' -l from      -d 'Start tag' -x -a "(git tag 2>/dev/null)"
complete -c git-summary -n '__fish_seen_subcommand_from changelog' -l to        -d 'End tag'   -x -a "HEAD (git tag 2>/dev/null)"
complete -c git-summary -n '__fish_seen_subcommand_from changelog' -l md        -d 'Markdown output'
complete -c git-summary -n '__fish_seen_subcommand_from changelog' -l output    -d 'Output file'
complete -c git-summary -n '__fish_seen_subcommand_from changelog' -l list-tags -d 'List tags'

# compare flags
complete -c git-summary -n '__fish_seen_subcommand_from compare' -l since    -d 'Time range' -x -a "'1 week ago' '1 month ago'"
complete -c git-summary -n '__fish_seen_subcommand_from compare' -l top      -d 'Top N'      -x -a "5 10 20"
complete -c git-summary -n '__fish_seen_subcommand_from compare' -l no-color -d 'Disable color'

# completion shells
complete -c git-summary -n '__fish_seen_subcommand_from completion' -f -a "bash zsh fish"
`
