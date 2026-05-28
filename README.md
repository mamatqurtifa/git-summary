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
| `completion <shell>` | Shell tab completion (bash/zsh/fish) |
| `version` | Print version |

### Export
Flag `-format json|csv|md` exports any summary to a file or stdout — useful in CI pipelines or for sharing with stakeholders.

---

## Installation

### Windows

**Step 1 — Download**

Go to the [Releases](https://github.com/mamatqurtifa/git-summary/releases/latest) page and download the correct file for your machine:

| Your machine | File to download |
|---|---|
| Most laptops & desktops (64-bit Intel/AMD) | `git-summary-windows-amd64.exe` |
| ARM-based Windows (Surface Pro X, etc.) | `git-summary-windows-arm64.exe` |

Not sure which one? Open PowerShell and run:
```powershell
echo $env:PROCESSOR_ARCHITECTURE
# AMD64 → download amd64
# ARM64 → download arm64
```

**Step 2 — Rename**

Rename the downloaded file to `git-summary.exe` (remove the `-windows-amd64` part).

**Step 3 — Move to a folder in your PATH**

Create a dedicated folder for CLI tools and move the file there:

```powershell
# Create the folder (skip if it already exists)
mkdir "$env:USERPROFILE\bin"

# Move the renamed file
Move-Item "$env:USERPROFILE\Downloads\git-summary.exe" "$env:USERPROFILE\bin\git-summary.exe"
```

**Step 4 — Add the folder to PATH**

1. Open **Start Menu**, search for **"Edit the system environment variables"**, and open it
2. Click **"Environment Variables..."**
3. Under **"User variables"**, select **Path** and click **"Edit"**
4. Click **"New"** and paste: `C:\Users\<your-username>\bin`
5. Click **OK** on all dialogs

**Step 5 — Verify**

Open a **new** PowerShell window (important — existing windows won't have the new PATH) and run:

```powershell
git-summary --help
```

> **SmartScreen warning:** The first time you run the file, Windows may show _"Windows protected your PC"_. Click **"More info"** → **"Run anyway"**. This is normal for open-source binaries without a paid code-signing certificate.

---

### macOS

```bash
# Apple Silicon — M1, M2, M3 (arm64)
curl -sSL https://github.com/mamatqurtifa/git-summary/releases/latest/download/git-summary-darwin-arm64 \
  -o git-summary
chmod +x git-summary
sudo mv git-summary /usr/local/bin/

# Intel Mac (amd64)
curl -sSL https://github.com/mamatqurtifa/git-summary/releases/latest/download/git-summary-darwin-amd64 \
  -o git-summary
chmod +x git-summary
sudo mv git-summary /usr/local/bin/
```

> **Gatekeeper warning:** If you see _"cannot be opened because the developer cannot be verified"_, run:
> ```bash
> xattr -d com.apple.quarantine /usr/local/bin/git-summary
> ```

---

### Linux

```bash
# amd64 (most servers and desktops)
curl -sSL https://github.com/mamatqurtifa/git-summary/releases/latest/download/git-summary-linux-amd64 \
  -o git-summary
chmod +x git-summary
sudo mv git-summary /usr/local/bin/

# arm64 (Raspberry Pi, AWS Graviton, etc.)
curl -sSL https://github.com/mamatqurtifa/git-summary/releases/latest/download/git-summary-linux-arm64 \
  -o git-summary
chmod +x git-summary
sudo mv git-summary /usr/local/bin/
```

---

### Install with Go (all platforms)

```bash
go install github.com/mamatqurtifa/git-summary@latest
```

### Build from source

```bash
git clone https://github.com/mamatqurtifa/git-summary.git
cd git-summary
./install.sh        # Linux / macOS — builds + installs + sets up shell completion
```

**Requirements:** Go 1.21+, Git 2.x

---

## Command reference

### `git-summary` — main summary

Run inside any git repository. Without `-since`, an interactive time-range picker appears.

```
git-summary [options] [path]
```

**Options:**

| Flag | Default | Description |
|------|---------|-------------|
| `-since` | interactive picker | Start date. Accepts natural language or ISO date |
| `-until` | now | End date |
| `-author` | all authors | Filter by author name or email (partial match) |
| `-branch` | current branch | Branch to analyse |
| `-top` | 10 | Number of top contributors to show |
| `-no-color` | false | Disable ANSI colors (useful for piping) |
| `-trends` | false | Show contributor trends & bus factor section |
| `-dirs` | false | Show directory ownership breakdown section |
| `-workpattern` | false | Show work pattern analysis section |
| `-all` | false | Show all extra sections at once |
| `-work-start` | 9 | Work day start hour (for workpattern analysis) |
| `-work-end` | 18 | Work day end hour (for workpattern analysis) |
| `-format` | — | Export format: `json`, `csv`, or `md` |
| `-output` | stdout | Write export output to a file |

**Examples:**

```bash
# Interactive picker (no arguments)
git-summary

# Last week, all sections
git-summary --since "1 week ago" --all

# Last month, filter by author
git-summary --since "1 month ago" --author "alice"

# Specific date range
git-summary --since 2025-01-01 --until 2025-06-30

# Analyse a specific branch, show top 5 only
git-summary --branch main --top 5

# Analyse a different repo
git-summary /path/to/other/repo

# No color output (for scripts or piping)
git-summary --since "1 month ago" --no-color

# Adjust work hours for workpattern analysis
git-summary --workpattern --work-start 8 --work-end 17
```

**Accepted `-since` / `-until` values:**

```
"24 hours ago"      "1 week ago"       "2 weeks ago"
"1 month ago"       "3 months ago"     "6 months ago"
"1 year ago"        2025-01-01         yesterday
today               (empty = all time)
```

---

### `git-summary compare` — branch comparison

Shows a side-by-side diff of two branches: commits, active days, lines changed, top contributors, who is unique to each branch, and files changed in both branches.

```
git-summary compare <branchA> <branchB> [options]
```

**Options:**

| Flag | Default | Description |
|------|---------|-------------|
| `-since` | 1 month ago | Time range to compare |
| `-top` | 10 | Top N contributors per branch |
| `-no-color` | false | Disable colors |

**Examples:**

```bash
# Compare main vs develop (last month)
git-summary compare main develop

# Compare with custom time range — flags can go anywhere
git-summary compare main develop --since "3 months ago"
git-summary compare --since "1 week ago" main develop

# Compare in a different repo
git-summary compare main develop /path/to/repo
```

**Sample output:**

```
  Branch Comparison
  ─────────────────────────────────────────────────────────────
  main                           │ develop
  ─────────────────────────────────────────────────────────────
  Commits               45       │ 62
  Active days           18       │ 24
  Lines added          +2301     │ +4120
  Lines deleted        -820      │ -1540
  Contributors            3      │ 4

  Top contributors
  Alice Chen    32    │ Alice Chen    41
  Bob Smith     13    │ Bob Smith     15
                      │ carol99        6

  contributors only in develop: carol99

  Files changed in both branches
  ⚠ src/api/routes.ts
  ⚠ package.json
```

---

### `git-summary changelog` — release diff

Generates a categorised changelog between two git tags or refs. Commits are grouped by conventional commit prefix (`feat`, `fix`, `docs`, `chore`, `refactor`, `perf`, `test`).

```
git-summary changelog [options] [path]
```

**Options:**

| Flag | Default | Description |
|------|---------|-------------|
| `-from` | auto (second-latest tag) | Start tag or ref |
| `-to` | HEAD | End tag or ref |
| `-md` | false | Output as Markdown instead of terminal format |
| `-output` | stdout | Write to file |
| `-list-tags` | false | List all available tags and exit |
| `-no-color` | false | Disable colors |

**Examples:**

```bash
# Auto-detect last two tags
git-summary changelog

# Between specific tags
git-summary changelog --from v1.0.0 --to v1.1.0

# From a tag to current HEAD
git-summary changelog --from v1.0.0

# Save as Markdown (great for CHANGELOG.md or GitHub release notes)
git-summary changelog --from v1.0.0 --to v1.1.0 --md --output CHANGELOG.md

# See what tags are available
git-summary changelog --list-tags
```

**Sample output:**

```
  Changelog: v1.0.0 → v1.1.0
  ──────────────────────────────────────────────────
  8 total commits

  ✨ New features (2)
  a1b2c3d  feat: add --dirs flag for directory breakdown  — Alice
  d4e5f6a  feat: export to CSV format  — Bob

  🐛 Bug fixes (3)
  b2c3d4e  fix: prevent panic on long file paths  — Alice
  c3d4e5f  fix: workpattern percentages rounding  — Bob
  e5f6a7b  fix: compare branch flag parsing  — Alice

  📝 Documentation (1)
  f6a7b8c  docs: update Windows installation guide  — Alice

  🔧 Chores (2)
  7b8c9d0  chore: update go.mod module path  — Bob
  8c9d0e1  chore: add CI workflow  — Alice
```

---

### Export (`-format`)

Export the summary data to JSON, CSV, or Markdown. Can be combined with any filter flag.

```bash
git-summary [filter options] -format <json|csv|md> [-output <file>]
```

**Examples:**

```bash
# Print JSON to terminal
git-summary --since "1 month ago" --format json

# Save JSON to file (for CI, dashboards, or further processing)
git-summary --since "1 month ago" --format json --output report.json

# Save Markdown report (ready to share with team or stakeholders)
git-summary --since "1 month ago" --format md --output REPORT.md

# Save CSV (open in Excel or Google Sheets)
git-summary --format csv --output data.csv

# Pipe JSON into jq for custom queries
git-summary --format json | jq '.contributors[0]'
git-summary --format json | jq '.total_commits'
git-summary --format json | jq '[.contributors[] | {name, commits}]'
```

**JSON structure:**

```json
{
  "generated_at": "2025-05-27T10:00:00Z",
  "date_range": "Apr 27, 2025 → May 27, 2025",
  "total_commits": 142,
  "total_added": 8423,
  "total_deleted": 3201,
  "active_days": 89,
  "avg_commits_per_day": 1.6,
  "most_active_hour": 10,
  "most_active_day": "Tuesday",
  "contributors": [
    { "name": "Alice", "email": "alice@example.com", "commits": 87, "added": 5420, "deleted": 1823 }
  ],
  "top_files": [
    { "path": "internal/parser.go", "changes": 38 }
  ],
  "weekly_activity": [
    { "week": "2025-W17", "commits": 12 }
  ],
  "hourly_activity": [0, 0, 0, 0, 0, 1, 3, 8, 12, 18, 21, ...]
}
```

---

### Shell completion

After setup, `git-summary [TAB][TAB]` completes flags, subcommands, branch names, tag names, and date values automatically.

```bash
# bash — add to ~/.bashrc
source <(git-summary completion bash)

# zsh — add to ~/.zshrc
source <(git-summary completion zsh)

# fish — one-time install, auto-loads on next session
git-summary completion fish > ~/.config/fish/completions/git-summary.fish
```

**What gets completed:**

```
git-summary [TAB]
→ compare  changelog  completion  version  -since  -until  -author ...

git-summary -since [TAB]
→ today  yesterday  "1 week ago"  "2 weeks ago"  "1 month ago" ...

git-summary -branch [TAB]
→ main  develop  feature/xyz   (pulled live from git branch)

git-summary changelog --from [TAB]
→ v1.0.0  v1.0.1  v1.1.0       (pulled live from git tag)

git-summary -format [TAB]
→ json  csv  md
```

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
  carol99                   14   9.9%   ✗ gone      ▄▃▁▁·····

  ⚑ bus factor: 2 people own 80% of commits ← at risk
      → Alice Chen, Bob Smith
```

**Trend labels:**

| Label | Meaning |
|-------|---------|
| `↑ rising` | Activity increased significantly in recent weeks |
| `↓ falling` | Activity decreased significantly in recent weeks |
| `→ stable` | Consistent commit frequency |
| `★ new` | No prior activity, recently started committing |
| `✗ gone` | Was active before, no recent commits |

**Bus factor** is the minimum number of contributors who collectively hold 80% of the commit history. A bus factor of 1 means the project depends critically on a single person.

---

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
    ├  Bob                                             71%
    ├  Alice                                           29%
```

Directories with a sole owner (≥90%) are highlighted in yellow — indicating a knowledge silo and potential bus factor risk at the folder level.

---

### `-workpattern` — Work pattern analysis

```
  Work Pattern Analysis
  ─────────────────────────────────────────────────────────────────
  team →  weekend: 31%   after-hours: 55%
  ⚠ Team: 31% of all commits happen on weekends — review workload

  author               commits  weekend  after-hrs   peak  signals
  ─────────────────────────────────────────────────────────────────
  Alice Chen                87      8%        22%  10:00 Mon
  Bob Smith                 41     41%        68%  23:00 Sat  ⚠ 41% weekend commits
```

**Signals detected:**

| Signal | Threshold |
|--------|-----------|
| Weekend commits | ≥30% of commits on Sat/Sun |
| After-hours commits | ≥50% outside configured work hours |
| Night owl | ≥25% of commits after 22:00 |
| Early bird | ≥20% of commits before 07:00 |

Work hours default to 09:00–18:00. Adjust with `-work-start` and `-work-end`:

```bash
git-summary --workpattern --work-start 8 --work-end 17
```

---

## Project structure

```
git-summary/
├── main.go                          Entry point
├── go.mod                           Module definition (zero external deps)
├── install.sh                       Build + install + completion setup (Linux/macOS)
├── .github/
│   └── workflows/
│       ├── ci.yml                   Build check on every push
│       └── release.yml              Cross-platform build on every tag push
├── cmd/
│   └── root.go                      CLI flags, subcommand routing, shell completion scripts
└── internal/
    ├── gitlog/
    │   └── parser.go                Shells out to git log, parses into []Commit
    ├── stats/
    │   └── compute.go               Core metrics — contributors, files, heatmap, weekly chart
    ├── display/
    │   └── render.go                ANSI terminal rendering
    ├── export/
    │   └── export.go                JSON / CSV / Markdown serialisation
    ├── compare/
    │   └── compare.go               Branch-vs-branch diff logic and renderer
    ├── tagdiff/
    │   └── tagdiff.go               Tag/ref range commit parser and changelog renderer
    ├── trend/
    │   └── trend.go                 Per-author sparklines, trend detection, bus factor
    ├── dirmap/
    │   └── dirmap.go                Directory-level file grouping and ownership percentages
    ├── workpattern/
    │   └── workpattern.go           Timestamp pattern analysis and burnout signal detection
    └── prompt/
        └── picker.go                Interactive time-range picker for terminal
```

---

## Contributing

Contributions are very welcome — whether it's a bug fix, new feature, documentation improvement, or just a typo correction.

**Before submitting a PR for a large change, please open an issue first** so we can discuss the approach. For small fixes, feel free to open a PR directly.

```bash
# Clone the repo
git clone https://github.com/mamatqurtifa/git-summary.git
cd git-summary

# Build
go build -ldflags="-s -w" -o git-summary .

# Run tests
go test ./...

# Vet
go vet ./...
```

### Ideas for contribution

If you're looking for something to work on, here are areas that would be great additions:

- **Phase 3** — Full TUI interactive mode (`git-summary tui`) using [bubbletea](https://github.com/charmbracelet/bubbletea)
- **Phase 3** — HTML report export (`--html report.html`) with interactive charts
- **Phase 4** — `git-summary serve` — live dashboard in the browser
- **Phase 4** — GitHub Action for automated weekly Slack/Discord reports
- More export formats (e.g. XML, NDJSON)
- Windows shell completion (PowerShell)
- Tests for existing packages

---

## Reporting bugs & getting help

Found a bug? Something not working as expected? Have a question?

**→ [Open an issue on GitHub](https://github.com/mamatqurtifa/git-summary/issues/new)**

When reporting a bug, please include:
- Your OS and architecture (e.g. Windows amd64, macOS arm64)
- The `git-summary version` output
- The exact command you ran
- The full error message or unexpected output

You can also reach me directly at **[@mamatqurtifa](https://github.com/mamatqurtifa)** on GitHub.

