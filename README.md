# git-summary

> Git repository activity at a glance — contributors, hot files, commit heatmap, weekly trends, branch comparison, changelog generation, directory ownership, contributor trends & bus factor, and work pattern analysis. All in your terminal.

```
╭────────────────────────────────────────────────────────────╮
│  git-summary                                               │
│  Jan 1, 2024 → May 27, 2025                                │
╰────────────────────────────────────────────────────────────╯

  Overview          Contributors        Most Changed Files
  ─────────────     ──────────────────  ────────────────────
  Commits    142     1. Alice  ████  87  internal/parser.go 38
  Active days 89     2. Bob    ███   41  cmd/root.go        24
  +8423 / -3201      3. Carol  █     14  README.md          12
```

## Features

### Core summary (`git-summary`)
- **Overview stats** — commits, active days, avg/day, lines added/deleted
- **Top contributors** with visual bar charts
- **Most changed files** with activity bars
- **Hourly heatmap** — when does your team actually commit?
- **Weekly activity chart** — visualise project momentum

### Extra sections (flags)
| Flag | What it shows |
|------|--------------|
| `-trends` | Per-contributor sparklines, rising/falling/gone signals, **bus factor** |
| `-dirs` | Directory ownership map — who owns what folder |
| `-workpattern` | Weekend commits, after-hours %, overtime & burnout signals |
| `-all` | All of the above at once |

### Subcommands
| Command | What it does |
|---------|-------------|
| `compare <A> <B>` | Side-by-side branch comparison |
| `changelog` | Auto-categorised changelog between two tags/refs |
| `completion <shell>` | Shell completion (bash/zsh/fish) |
| `version` | Print version |

### Export
Flag `-format json|csv|md` exports any summary to a file or stdout — useful in CI pipelines or for sharing with stakeholders.

---

## Installation

### Linux

```bash
# amd64 (most servers and desktops)
curl -sSL https://github.com/yourusername/git-summary/releases/latest/download/git-summary-linux-amd64 \
  -o git-summary
chmod +x git-summary
sudo mv git-summary /usr/local/bin/

# arm64 (Raspberry Pi, AWS Graviton, etc.)
curl -sSL https://github.com/yourusername/git-summary/releases/latest/download/git-summary-linux-arm64 \
  -o git-summary
chmod +x git-summary
sudo mv git-summary /usr/local/bin/
```

### macOS

```bash
# Apple Silicon — M1, M2, M3 (arm64)
curl -sSL https://github.com/yourusername/git-summary/releases/latest/download/git-summary-darwin-arm64 \
  -o git-summary
chmod +x git-summary
sudo mv git-summary /usr/local/bin/

# Intel Mac (amd64)
curl -sSL https://github.com/yourusername/git-summary/releases/latest/download/git-summary-darwin-amd64 \
  -o git-summary
chmod +x git-summary
sudo mv git-summary /usr/local/bin/
```

> **macOS note:** if you see _"cannot be opened because the developer cannot be verified"_, run:
> `xattr -d com.apple.quarantine /usr/local/bin/git-summary`

### Windows

1. Go to the [Releases](https://github.com/yourusername/git-summary/releases/latest) page
2. Download `git-summary-windows-amd64.exe`
3. Rename it to `git-summary.exe`
4. Move it to a folder that's in your `PATH`, for example `C:\Users\<you>\bin\`

Or with PowerShell:

```powershell
Invoke-WebRequest `
  -Uri "https://github.com/yourusername/git-summary/releases/latest/download/git-summary-windows-amd64.exe" `
  -OutFile "$env:USERPROFILE\bin\git-summary.exe"
```

> Make sure `$env:USERPROFILE\bin` is in your `PATH`. Search "environment variables" in the Start menu to add it.

### Install with Go (all platforms)

```bash
go install github.com/yourusername/git-summary@latest
```

### Build from source

```bash
git clone https://github.com/yourusername/git-summary.git
cd git-summary
./install.sh        # Linux / macOS — builds + installs + sets up shell completion
```

**Requirements:** Go 1.21+, Git 2.x

---

## Usage

```bash
# Run inside any git repo (shows interactive time-range picker)
git-summary

# With flags
git-summary --since "1 week ago"
git-summary --since "1 month ago" --all
git-summary --author alice --top 5
git-summary --branch main --no-color
```

### All flags

```
Main options:
  -since string      Start date (default: interactive picker)
                     e.g. "1 week ago", "1 month ago", "2024-01-01"
  -until string      End date (default: now)
  -author string     Filter by author (partial name or email)
  -branch string     Branch to analyze (default: current)
  -top int           Top N contributors (default: 10)
  -no-color          Disable colored output

Extra sections:
  -trends            Contributor trends & bus factor
  -dirs              Directory ownership breakdown
  -workpattern       Work pattern analysis
  -all               Show all extra sections
  -work-start int    Work day start hour (default: 9)
  -work-end int      Work day end hour   (default: 18)

Export:
  -format string     json | csv | md
  -output string     Write to file (default: stdout)
```

---

## Subcommands

### `compare` — branch comparison

```bash
git-summary compare main develop
git-summary compare main develop --since "2 weeks ago"
git-summary compare main develop /path/to/repo
```

Shows side-by-side: commits, active days, lines changed, top contributors per branch, contributors unique to each branch, and files changed in both (potential merge conflicts).

### `changelog` — release diff

```bash
# Auto-detect last two tags
git-summary changelog

# Specific range
git-summary changelog --from v1.0.0 --to v1.1.0

# Markdown output (great for CHANGELOG.md)
git-summary changelog --from v1.0.0 --md --output CHANGELOG.md

# List available tags
git-summary changelog --list-tags
```

Commits are auto-categorised by conventional commit prefix (`feat`, `fix`, `docs`, `chore`, etc.). Unknown prefixes go into "Other changes".

---

## Extra sections in detail

### `-trends` — Contributor trends & bus factor

```
  Contributor Trends
  ────────────────────────────────────────────────────────────
  author               commits  share   trend       last 8 weeks →
  ────────────────────────────────────────────────────────────
  Alice Chen                87  61.3%   ↑ rising    ▁▂▃▄▆▇█▇
  Bob Smith                 41  28.9%   → stable    ▃▃▄▃▄▃▄▃
  carol99                   14   9.9%   ✗ gone      ▄▃▁▁··· ·

  ⚑ bus factor: 2 people own 80% of commits ← at risk
      → Alice Chen, Bob Smith
```

**Bus factor** is the minimum number of people who hold 80% of commit history. A bus factor of 1 is a critical single point of failure.

### `-dirs` — Directory ownership

```
  Directory Breakdown
  ────────────────────────────────────────────────────────────
  directory                    changes  top owner           ownership bar
  ────────────────────────────────────────────────────────────
  internal/stats                    42  Alice          94%  ████████████████░░
    ├  Alice                                           94%
    ├  Bob                                              6%
  cmd                               28  Bob            71%  ████████████░░░░░░
```

Directories highlighted in yellow have a sole owner (≥90%) — a potential bus factor risk at the folder level.

### `-workpattern` — Work pattern analysis

```
  Work Pattern Analysis
  ─────────────────────────────────────────────────────────────────
  team →  weekend: 31%   after-hours: 55%
  ⚠ Team: 31% of all commits happen on weekends — review workload

  author               commits  weekend  after-hrs   peak  signals
  ─────────────────────────────────────────────────────────────────
  Alice Chen                87      8%        22%  10:00 Mon
  Bob Smith                 41     41%        68%  23:00 Sat  ⚠ 41% weekend commits — possible burnout
```

Work hours default to 09:00–18:00. Override with `-work-start` and `-work-end`.

---

## Export examples

```bash
# JSON for CI or dashboards
git-summary --since "1 month ago" --format json --output report.json

# Markdown report for stakeholders
git-summary --since "1 month ago" --format md --output REPORT.md

# CSV for spreadsheets
git-summary --format csv --output data.csv

# Pipe-friendly
git-summary --format json | jq '.contributors[0]'
```

---

## Shell completion

```bash
# bash — add to ~/.bashrc
source <(git-summary completion bash)

# zsh — add to ~/.zshrc
source <(git-summary completion zsh)

# fish — one-time install
git-summary completion fish > ~/.config/fish/completions/git-summary.fish
```

After setup, `git-summary [TAB][TAB]` completes flags, subcommands, branch names, tag names, and date values.

---

## Project structure

```
git-summary/
├── main.go                          Entry point
├── install.sh                       Build + install + completion setup
├── cmd/
│   └── root.go                      CLI flags, subcommand routing
└── internal/
    ├── gitlog/
    │   └── parser.go                git log → []Commit
    ├── stats/
    │   └── compute.go               Core metrics (contributors, files, heatmap)
    ├── display/
    │   └── render.go                Terminal rendering + ANSI colors
    ├── export/
    │   └── export.go                JSON / CSV / Markdown export
    ├── compare/
    │   └── compare.go               Branch-vs-branch comparison
    ├── tagdiff/
    │   └── tagdiff.go               Tag/ref changelog generator
    ├── trend/
    │   └── trend.go                 Contributor sparklines & bus factor
    ├── dirmap/
    │   └── dirmap.go                Directory ownership breakdown
    ├── workpattern/
    │   └── workpattern.go           Work hour & burnout signal analysis
    └── prompt/
        └── picker.go                Interactive time-range picker
```

| Package | Responsibility |
|---------|---------------|
| `cmd` | Parses flags, routes to subcommands, wires everything |
| `internal/gitlog` | Shells out to `git log --numstat`, parses into `[]Commit` |
| `internal/stats` | Pure computation — no I/O, easily testable |
| `internal/display` | ANSI terminal rendering |
| `internal/export` | Serialises to JSON / CSV / Markdown |
| `internal/compare` | Branch diff logic + renderer |
| `internal/tagdiff` | Tag-range commit parsing + changelog renderer |
| `internal/trend` | Time-series per author, bus factor detection |
| `internal/dirmap` | File-path grouping, ownership percentages |
| `internal/workpattern` | Timestamp pattern analysis, burnout signals |
| `internal/prompt` | Interactive terminal picker (no external deps) |

---

## Contributing

PRs welcome. Please open an issue first for large changes.

```bash
go test ./...
go build -ldflags="-s -w" -o git-summary .
```

## License

MIT
