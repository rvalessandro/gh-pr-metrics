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

**`--format csv`**

```
$ gh pr-metrics --repo systeric/xyz-monorepo --start 2026-03-22 --end 2026-04-22 --format csv
repo,pr,author,size,commits,additions,deletions,changed_files,comments,participants,merged_at,time_to_first_review_min,first_to_last_review_min,review_to_approval_min,first_approval_to_merge_min,created_to_merged_min,feature_lead_time_min
systeric/xyz-monorepo,142,mharris,M,3,187,42,8,4,3,2026-04-18T14:32:00Z,151,,,18,194,194
systeric/xyz-monorepo,139,jpark,S,2,38,11,4,2,2,2026-04-15T11:08:00Z,107,,,12,125,125
systeric/xyz-monorepo,137,tcole,L,8,621,203,22,7,4,2026-04-11T16:45:00Z,252,185,185,25,1908,1921
systeric/xyz-monorepo,135,mharris,XS,1,5,2,2,0,2,2026-04-08T09:20:00Z,45,,,8,62,62
```

Empty duration cells mean the stage was not observed (no review, or self-merged). All duration columns are whole minutes.

**`--format json`**

Duration fields in `rows` are whole minutes (`null` when not observed). Durations in `summary` and `ByAuthor` are nanoseconds (`time.Duration`).

```json
{
  "rows": [
    {
      "repo": "systeric/xyz-monorepo",
      "number": 142,
      "author": "mharris",
      "size": "M",
      "commits": 3,
      "additions": 187,
      "deletions": 42,
      "changed_files": 8,
      "comments": 4,
      "participants": 3,
      "merged_at": "2026-04-18T14:32:00Z",
      "time_to_first_review_min": 151,
      "first_to_last_review_min": null,
      "review_to_approval_min": null,
      "first_approval_to_merge_min": 18,
      "created_to_merged_min": 194,
      "feature_lead_time_min": 194
    },
    {
      "repo": "systeric/xyz-monorepo",
      "number": 139,
      "author": "jpark",
      "size": "S",
      "commits": 2,
      "additions": 38,
      "deletions": 11,
      "changed_files": 4,
      "comments": 2,
      "participants": 2,
      "merged_at": "2026-04-15T11:08:00Z",
      "time_to_first_review_min": 107,
      "first_to_last_review_min": null,
      "review_to_approval_min": null,
      "first_approval_to_merge_min": 12,
      "created_to_merged_min": 125,
      "feature_lead_time_min": 125
    },
    {
      "repo": "systeric/xyz-monorepo",
      "number": 137,
      "author": "tcole",
      "size": "L",
      "commits": 8,
      "additions": 621,
      "deletions": 203,
      "changed_files": 22,
      "comments": 7,
      "participants": 4,
      "merged_at": "2026-04-11T16:45:00Z",
      "time_to_first_review_min": 252,
      "first_to_last_review_min": 185,
      "review_to_approval_min": 185,
      "first_approval_to_merge_min": 25,
      "created_to_merged_min": 1908,
      "feature_lead_time_min": 1921
    },
    {
      "repo": "systeric/xyz-monorepo",
      "number": 135,
      "author": "mharris",
      "size": "XS",
      "commits": 1,
      "additions": 5,
      "deletions": 2,
      "changed_files": 2,
      "comments": 0,
      "participants": 2,
      "merged_at": "2026-04-08T09:20:00Z",
      "time_to_first_review_min": 45,
      "first_to_last_review_min": null,
      "review_to_approval_min": null,
      "first_approval_to_merge_min": 8,
      "created_to_merged_min": 62,
      "feature_lead_time_min": 62
    }
  ],
  "summary": {
    "Total": 4,
    "WindowDays": 31,
    "PerWeek": 0.9032258064516129,
    "Adds": 851,
    "Dels": 258,
    "ChangedFiles": 36,
    "TimeToFirstReview":    { "P50": 6420000000000,   "P90": 9060000000000,   "Max": 15120000000000,  "N": 4 },
    "FirstToLastReview":    { "P50": 11100000000000,  "P90": 11100000000000,  "Max": 11100000000000,  "N": 1 },
    "ReviewToApproval":     { "P50": 11100000000000,  "P90": 11100000000000,  "Max": 11100000000000,  "N": 1 },
    "FirstApprovalToMerge": { "P50": 720000000000,    "P90": 1080000000000,   "Max": 1500000000000,   "N": 4 },
    "CreatedToMerged":      { "P50": 7500000000000,   "P90": 11640000000000,  "Max": 114480000000000, "N": 4 },
    "FeatureLeadTime":      { "P50": 7500000000000,   "P90": 11640000000000,  "Max": 115260000000000, "N": 4 },
    "Sizes": { "XS": 1, "S": 1, "M": 1, "L": 1, "XL": 0, "XXL": 0 },
    "ByAuthor": [
      {
        "Login": "mharris", "PRs": 2, "Adds": 192, "Dels": 44, "AvgLines": 118, "MedLines": 229,
        "TTFR":        { "P50": 2700000000000, "P90": 2700000000000, "Max": 9060000000000,   "N": 2 },
        "Feedback":    { "P50": 0,             "P90": 0,             "Max": 0,               "N": 0 },
        "ApprToMerge": { "P50": 480000000000,  "P90": 480000000000,  "Max": 1080000000000,   "N": 2 },
        "E2E":         { "P50": 3720000000000, "P90": 3720000000000, "Max": 11640000000000,  "N": 2 }
      },
      {
        "Login": "jpark", "PRs": 1, "Adds": 38, "Dels": 11, "AvgLines": 49, "MedLines": 49,
        "TTFR":        { "P50": 6420000000000, "P90": 6420000000000, "Max": 6420000000000, "N": 1 },
        "Feedback":    { "P50": 0,             "P90": 0,             "Max": 0,             "N": 0 },
        "ApprToMerge": { "P50": 720000000000,  "P90": 720000000000,  "Max": 720000000000,  "N": 1 },
        "E2E":         { "P50": 7500000000000, "P90": 7500000000000, "Max": 7500000000000, "N": 1 }
      },
      {
        "Login": "tcole", "PRs": 1, "Adds": 621, "Dels": 203, "AvgLines": 824, "MedLines": 824,
        "TTFR":        { "P50": 15120000000000,  "P90": 15120000000000,  "Max": 15120000000000,  "N": 1 },
        "Feedback":    { "P50": 11100000000000,  "P90": 11100000000000,  "Max": 11100000000000,  "N": 1 },
        "ApprToMerge": { "P50": 1500000000000,   "P90": 1500000000000,   "Max": 1500000000000,   "N": 1 },
        "E2E":         { "P50": 114480000000000, "P90": 114480000000000, "Max": 114480000000000, "N": 1 }
      }
    ]
  }
}
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
