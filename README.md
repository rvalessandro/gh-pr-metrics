# gh-pr-metrics

A `gh` CLI extension that prints PR cycle-time metrics for one or more repositories, with aggregates, size distribution, and per-author rollup.

## Install

```bash
gh extension install rvalessandro/gh-pr-metrics
```

Requires `gh auth login` first.

## Usage

```bash
# current repo, last 30 days
gh pr-metrics

# explicit range, multi-repo, whitelist authors
gh pr-metrics \
  --repo systeric/xyz-monorepo \
  --repo systeric/xyz-api \
  --users mharris,tcole,jpark \
  --start 2026-03-22 --end 2026-04-22

# CSV for spreadsheets
gh pr-metrics --format csv > metrics.csv

# JSON for tooling
gh pr-metrics --format json > metrics.json

# Markdown for sharing
gh pr-metrics --format md > metrics.md
```

## Flags

| flag | default | description |
| --- | --- | --- |
| `--repo` | current repo | target repo `owner/name` (repeatable or comma-separated) |
| `--users` | — | whitelist of author logins (filters out PRs by others) |
| `--exclude-users` | — | author logins to exclude (repeatable or comma-separated) |
| `--exclude-bots` | `false` | drop PRs whose author is a GitHub App/Bot (dependabot, renovate, …) |
| `--start` | 30 days ago | start date `YYYY-MM-DD` (inclusive) |
| `--end` | today | end date `YYYY-MM-DD` (inclusive) |
| `--query` | — | extra GitHub search qualifiers to append |
| `--format` | `table` | `table`, `csv`, `json`, `md` |
| `--chunk-days` | `7` | split wide date ranges into N-day chunks to avoid GQL timeouts (`0` disables) |
| `--timeout` | `60` | per-request GraphQL timeout in seconds |

## Example output

```
$ gh pr-metrics --repo systeric/xyz-monorepo --start 2026-03-22 --end 2026-04-22

REPO                    PR    AUTHOR   SIZE  COMMITS  ADD  DEL  FILES  COMMENTS  TTFIRST  FIRST→LAST  FIRSTAPPR→MERGE  LEADTIME
systeric/xyz-monorepo   #142  mharris  M     3        187   42     8         4    2h31m          --             18m    3h14m
systeric/xyz-monorepo   #139  jpark    S     2         38   11     4         2    1h47m          --             12m     2h5m
systeric/xyz-monorepo   #137  tcole    L     8        621  203    22         7    4h12m       3h5m             25m   31h48m
systeric/xyz-monorepo   #135  mharris  XS    1          5    2     2         0      45m          --              8m     1h2m

SUMMARY  4 PRs over 31 days (0.9/week)
  total churn: +851 / -258 across 36 file-changes

STAGE                                  P50     P90      MAX  N
1. created → first review            2h9m   4h12m    4h12m  4
2. first review → first approval     3h5m    3h5m     3h5m  1
3. first approval → merge             15m     25m      25m  4
4. created → merged (E2E)           2h38m  31h48m   31h48m  4
   feature lead time (commit → merge)  2h39m  32h1m  32h1m  4

SIZE DISTRIBUTION  XS:1 S:1 M:1 L:1 XL:0 XXL:0

BY AUTHOR — output
 LOGIN  PRS  +LINES  -LINES  AVG  MED
mharris    2     192      44  118  118
  jpark    1      38      11   49   38
  tcole    1     621     203  824  824

BY AUTHOR — cycle time (p50; use --format csv for p90)
 LOGIN  PRS   TTFR  FEEDBACK  APPR→MERGE    E2E
mharris    2  1h38m        --         13m  1h51m
  jpark    1  1h47m        --         12m   2h5m
  tcole    1  4h12m      3h5m         25m  31h48m

AVG/MED    = mean / median lines changed (add+del) per PR
TTFR       = created → first review
FEEDBACK   = first review → first approval (iteration loop)
APPR→MERGE = first approval → merge
E2E        = created → merged
```

## Metrics

**PR lifecycle stages** (all reported as p50 / p90 / max per stage, top-level and per author):

1. **TTFR** — created → first non-author review
2. **Feedback** — first review → first approval (iteration loop)
3. **Appr→Merge** — first approval → merged
4. **E2E** — created → merged (total PR cycle time)
5. **Feature lead time** — earliest commit → merged (includes work done before opening the PR)

**Per PR** (in table / CSV / JSON): commits, additions, deletions, files, comments, participants, all 5 durations above, size bucket.

**Summary**: PR count, throughput/week, total churn, p50 / p90 / max of each stage.

**Size distribution** (lines changed = additions + deletions):
XS (<10) · S (<50) · M (<250) · L (<1000) · XL (<5000) · XXL (≥5000).

**Per author** rollup: PR count, churn, p50/p90 of each stage.
