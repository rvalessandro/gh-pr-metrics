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
  --repo See-Dr-Pte-Ltd/reallysick-monorepo \
  --repo See-Dr-Pte-Ltd/another-repo \
  --users rvalessandro,dece88,aufaikrimaa \
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
