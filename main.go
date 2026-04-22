package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	gh "github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	graphql "github.com/cli/shurcooL-graphql"
)

type stringList []string

func (s *stringList) String() string { return strings.Join(*s, ",") }
func (s *stringList) Set(v string) error {
	for _, p := range strings.Split(v, ",") {
		if p = strings.TrimSpace(p); p != "" {
			*s = append(*s, p)
		}
	}
	return nil
}

func main() {
	var repos, users, excludeUsers stringList
	flag.Var(&repos, "repo", "target repo in owner/name form (repeatable or comma-separated)")
	flag.Var(&users, "users", "whitelist of author logins (repeatable or comma-separated); PRs by others are excluded")
	flag.Var(&excludeUsers, "exclude-users", "author logins to exclude (repeatable or comma-separated)")
	excludeBots := flag.Bool("exclude-bots", false, "exclude PRs whose author is a GitHub App/Bot (dependabot, renovate, etc.)")
	start := flag.String("start", "", "start date YYYY-MM-DD (default: 30 days ago)")
	end := flag.String("end", "", "end date YYYY-MM-DD (default: today)")
	extraQuery := flag.String("query", "", "extra GitHub search qualifiers appended to the merged PR query")
	format := flag.String("format", "table", "output format: table | csv | json | md")
	chunkDays := flag.Int("chunk-days", 7, "chunk the date range into N-day windows to avoid timeouts (0 = no chunking)")
	timeoutSec := flag.Int("timeout", 60, "per-request GraphQL timeout (seconds)")
	flag.Parse()

	today := time.Now().UTC().Format("2006-01-02")
	if *end == "" {
		*end = today
	}
	if *start == "" {
		*start = time.Now().UTC().AddDate(0, 0, -30).Format("2006-01-02")
	}

	if len(repos) == 0 {
		// Fallback to current repo if invoked inside a git checkout.
		repo, err := gh.CurrentRepository()
		if err != nil {
			log.Fatal("no --repo given and not in a repo checkout")
		}
		repos = stringList{fmt.Sprintf("%s/%s", repo.Owner(), repo.Name())}
	}

	client, err := gh.GQLClient(&api.ClientOptions{
		EnableCache: true,
		CacheTTL:    15 * time.Minute,
		Timeout:     time.Duration(*timeoutSec) * time.Second,
	})
	if err != nil {
		log.Fatal("run `gh auth login` first: ", err)
	}

	startT, err := time.Parse("2006-01-02", *start)
	must(err)
	endT, err := time.Parse("2006-01-02", *end)
	must(err)
	if endT.Before(startT) {
		log.Fatal("--end must be on/after --start")
	}
	windowDays := int(endT.Sub(startT).Hours()/24) + 1

	rows := collect(client, repos, users, excludeUsers, *excludeBots, startT, endT, *chunkDays, *extraQuery)

	s := summarizeAll(rows, windowDays)

	switch *format {
	case "table":
		writeTable(os.Stdout, rows, s)
	case "csv":
		must(writeCSV(os.Stdout, rows))
	case "json":
		must(writeJSON(os.Stdout, rows, s))
	case "md", "markdown":
		writeMarkdown(os.Stdout, rows, s)
	default:
		log.Fatalf("unknown --format %q", *format)
	}
}

func collect(client api.GQLClient, repos, users, excludeUsers stringList, excludeBots bool, startT, endT time.Time, chunkDays int, extraQuery string) []prRow {
	chunks := chunkRange(startT, endT, chunkDays)
	exclude := map[string]bool{}
	for _, u := range excludeUsers {
		exclude[u] = true
	}
	seen := map[string]bool{}
	var rows []prRow
	for _, repo := range repos {
		for _, ch := range chunks {
			prs := queryChunk(client, repo, users, ch.from, ch.to, extraQuery)
			for _, p := range prs {
				if excludeBots && p.Author.Typename == "Bot" {
					continue
				}
				if exclude[p.Author.Login] {
					continue
				}
				key := fmt.Sprintf("%s#%d", p.Repository.NameWithOwner, p.Number)
				if seen[key] {
					continue
				}
				seen[key] = true
				rows = append(rows, rowFromPR(p))
			}
		}
	}
	return rows
}

type dateChunk struct{ from, to time.Time }

func chunkRange(start, end time.Time, chunkDays int) []dateChunk {
	if chunkDays <= 0 {
		return []dateChunk{{start, end}}
	}
	var out []dateChunk
	cur := start
	for !cur.After(end) {
		next := cur.AddDate(0, 0, chunkDays-1)
		if next.After(end) {
			next = end
		}
		out = append(out, dateChunk{cur, next})
		cur = next.AddDate(0, 0, 1)
	}
	return out
}

func queryChunk(client api.GQLClient, repo string, users stringList, from, to time.Time, extra string) []pullRequestNode {
	q := fmt.Sprintf("repo:%s type:pr is:merged merged:%s..%s",
		repo, from.Format("2006-01-02"), to.Format("2006-01-02"))
	for _, u := range users {
		q += " author:" + u
	}
	if extra != "" {
		q += " " + extra
	}

	var out []pullRequestNode
	var cursor *graphql.String
	for {
		var qr metricsQuery
		vars := map[string]interface{}{
			"query":       graphql.String(q),
			"resultCount": graphql.Int(100),
			"afterCursor": cursor,
		}
		if err := client.Query("PRMetrics", &qr, vars); err != nil {
			log.Fatalf("query failed for %s %s..%s: %v", repo, from.Format("2006-01-02"), to.Format("2006-01-02"), err)
		}
		for _, n := range qr.Search.Nodes {
			if n.PullRequest.Number != 0 {
				out = append(out, n.PullRequest)
			}
		}
		if !qr.Search.PageInfo.HasNextPage {
			break
		}
		c := graphql.String(qr.Search.PageInfo.EndCursor)
		cursor = &c
	}
	return out
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
