package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"text/tabwriter"
	"time"
)

func fmtDur(d time.Duration) string {
	if d < 0 {
		return "--"
	}
	d = d.Round(time.Minute)
	h := int(d / time.Hour)
	m := int((d % time.Hour) / time.Minute)
	if h == 0 && m == 0 {
		return "<1m"
	}
	if h == 0 {
		return fmt.Sprintf("%dm", m)
	}
	return fmt.Sprintf("%dh%dm", h, m)
}

func sortRows(rows []prRow) {
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Repo != rows[j].Repo {
			return rows[i].Repo < rows[j].Repo
		}
		return rows[i].Number > rows[j].Number
	})
}

func writeTable(w io.Writer, rows []prRow, s summary) {
	sortRows(rows)
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "REPO\tPR\tAUTHOR\tSIZE\tCOMMITS\tADD\tDEL\tFILES\tCOMMENTS\tTTFIRST\tFIRST→LAST\tFIRSTAPPR→MERGE\tLEADTIME")
	for _, r := range rows {
		fmt.Fprintf(tw, "%s\t#%d\t%s\t%s\t%d\t%d\t%d\t%d\t%d\t%s\t%s\t%s\t%s\n",
			r.Repo, r.Number, r.Author, r.SizeBucket, r.Commits,
			r.Additions, r.Deletions, r.ChangedFiles, r.Comments,
			fmtDur(r.TimeToFirstReview),
			fmtDur(r.FirstToLastReview),
			fmtDur(r.FirstApprovalToMerge),
			fmtDur(r.FeatureLeadTime),
		)
	}
	tw.Flush()

	fmt.Fprintf(w, "\nSUMMARY  %d PRs over %d days (%.1f/week)\n", s.Total, s.WindowDays, s.PerWeek)
	fmt.Fprintf(w, "  total churn: +%d / -%d across %d file-changes\n", s.Adds, s.Dels, s.ChangedFiles)
	fmt.Fprintln(w, "")
	stw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(stw, "METRIC\tP50\tP90\tMAX\tN")
	fmt.Fprintf(stw, "time to first review\t%s\t%s\t%s\t%d\n",
		fmtDur(s.TimeToFirstReview.P50), fmtDur(s.TimeToFirstReview.P90),
		fmtDur(s.TimeToFirstReview.Max), s.TimeToFirstReview.N)
	fmt.Fprintf(stw, "first→last review\t%s\t%s\t%s\t%d\n",
		fmtDur(s.FirstToLastReview.P50), fmtDur(s.FirstToLastReview.P90),
		fmtDur(s.FirstToLastReview.Max), s.FirstToLastReview.N)
	fmt.Fprintf(stw, "first approval→merge\t%s\t%s\t%s\t%d\n",
		fmtDur(s.FirstApprovalToMerge.P50), fmtDur(s.FirstApprovalToMerge.P90),
		fmtDur(s.FirstApprovalToMerge.Max), s.FirstApprovalToMerge.N)
	fmt.Fprintf(stw, "feature lead time\t%s\t%s\t%s\t%d\n",
		fmtDur(s.FeatureLeadTime.P50), fmtDur(s.FeatureLeadTime.P90),
		fmtDur(s.FeatureLeadTime.Max), s.FeatureLeadTime.N)
	stw.Flush()

	fmt.Fprintf(w, "\nSIZE DISTRIBUTION  XS:%d S:%d M:%d L:%d XL:%d XXL:%d\n",
		s.Sizes.XS, s.Sizes.S, s.Sizes.M, s.Sizes.L, s.Sizes.XL, s.Sizes.XXL)

	if len(s.ByAuthor) > 0 {
		fmt.Fprintln(w, "\nBY AUTHOR")
		atw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		fmt.Fprintln(atw, "LOGIN\tPRS\tADD\tDEL\tTTFIRST(P50/P90)\tLEADTIME(P50/P90)")
		for _, a := range s.ByAuthor {
			fmt.Fprintf(atw, "%s\t%d\t%d\t%d\t%s / %s\t%s / %s\n",
				a.Login, a.PRs, a.Adds, a.Dels,
				fmtDur(a.FirstRev.P50), fmtDur(a.FirstRev.P90),
				fmtDur(a.LeadTime.P50), fmtDur(a.LeadTime.P90),
			)
		}
		atw.Flush()
	}
}

func writeCSV(w io.Writer, rows []prRow) error {
	sortRows(rows)
	c := csv.NewWriter(w)
	defer c.Flush()
	if err := c.Write([]string{
		"repo", "pr", "author", "size", "commits", "additions", "deletions",
		"changed_files", "comments", "participants", "merged_at",
		"time_to_first_review_min", "first_to_last_review_min",
		"first_approval_to_merge_min", "feature_lead_time_min",
	}); err != nil {
		return err
	}
	mins := func(d time.Duration) string {
		if d < 0 {
			return ""
		}
		return strconv.FormatInt(int64(d/time.Minute), 10)
	}
	for _, r := range rows {
		if err := c.Write([]string{
			r.Repo, strconv.Itoa(r.Number), r.Author, r.SizeBucket,
			strconv.Itoa(r.Commits), strconv.Itoa(r.Additions), strconv.Itoa(r.Deletions),
			strconv.Itoa(r.ChangedFiles), strconv.Itoa(r.Comments), strconv.Itoa(r.Participants),
			r.MergedAt.Format(time.RFC3339),
			mins(r.TimeToFirstReview), mins(r.FirstToLastReview),
			mins(r.FirstApprovalToMerge), mins(r.FeatureLeadTime),
		}); err != nil {
			return err
		}
	}
	return nil
}

func writeJSON(w io.Writer, rows []prRow, s summary) error {
	sortRows(rows)
	type rowJSON struct {
		Repo                     string `json:"repo"`
		Number                   int    `json:"number"`
		Author                   string `json:"author"`
		Size                     string `json:"size"`
		Commits                  int    `json:"commits"`
		Additions                int    `json:"additions"`
		Deletions                int    `json:"deletions"`
		ChangedFiles             int    `json:"changed_files"`
		Comments                 int    `json:"comments"`
		Participants             int    `json:"participants"`
		MergedAt                 string `json:"merged_at"`
		TimeToFirstReviewMin     *int64 `json:"time_to_first_review_min"`
		FirstToLastReviewMin     *int64 `json:"first_to_last_review_min"`
		FirstApprovalToMergeMin  *int64 `json:"first_approval_to_merge_min"`
		FeatureLeadTimeMin       *int64 `json:"feature_lead_time_min"`
	}
	mins := func(d time.Duration) *int64 {
		if d < 0 {
			return nil
		}
		v := int64(d / time.Minute)
		return &v
	}
	out := struct {
		Rows    []rowJSON `json:"rows"`
		Summary summary   `json:"summary"`
	}{Summary: s}
	for _, r := range rows {
		out.Rows = append(out.Rows, rowJSON{
			Repo: r.Repo, Number: r.Number, Author: r.Author, Size: r.SizeBucket,
			Commits: r.Commits, Additions: r.Additions, Deletions: r.Deletions,
			ChangedFiles: r.ChangedFiles, Comments: r.Comments, Participants: r.Participants,
			MergedAt:                r.MergedAt.Format(time.RFC3339),
			TimeToFirstReviewMin:    mins(r.TimeToFirstReview),
			FirstToLastReviewMin:    mins(r.FirstToLastReview),
			FirstApprovalToMergeMin: mins(r.FirstApprovalToMerge),
			FeatureLeadTimeMin:      mins(r.FeatureLeadTime),
		})
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func writeMarkdown(w io.Writer, rows []prRow, s summary) {
	sortRows(rows)
	fmt.Fprintf(w, "# PR Metrics — %d PRs over %d days (%.1f/week)\n\n", s.Total, s.WindowDays, s.PerWeek)
	fmt.Fprintf(w, "Total churn: **+%d / −%d** across %d file-changes\n\n", s.Adds, s.Dels, s.ChangedFiles)
	fmt.Fprintln(w, "| metric | p50 | p90 | max | N |")
	fmt.Fprintln(w, "| --- | --- | --- | --- | --- |")
	dr := func(name string, d durSummary) {
		fmt.Fprintf(w, "| %s | %s | %s | %s | %d |\n", name, fmtDur(d.P50), fmtDur(d.P90), fmtDur(d.Max), d.N)
	}
	dr("time to first review", s.TimeToFirstReview)
	dr("first→last review", s.FirstToLastReview)
	dr("first approval→merge", s.FirstApprovalToMerge)
	dr("feature lead time", s.FeatureLeadTime)
	fmt.Fprintf(w, "\n**Size distribution**: XS:%d S:%d M:%d L:%d XL:%d XXL:%d\n\n",
		s.Sizes.XS, s.Sizes.S, s.Sizes.M, s.Sizes.L, s.Sizes.XL, s.Sizes.XXL)
	if len(s.ByAuthor) > 0 {
		fmt.Fprintln(w, "## By author\n")
		fmt.Fprintln(w, "| login | PRs | add | del | ttfirst p50 | leadtime p50 |")
		fmt.Fprintln(w, "| --- | --- | --- | --- | --- | --- |")
		for _, a := range s.ByAuthor {
			fmt.Fprintf(w, "| %s | %d | %d | %d | %s | %s |\n",
				a.Login, a.PRs, a.Adds, a.Dels, fmtDur(a.FirstRev.P50), fmtDur(a.LeadTime.P50))
		}
	}
	fmt.Fprintln(w, "\n## Per-PR rows")
	fmt.Fprintln(w, "| repo | pr | author | size | add | del | files | ttfirst | leadtime |")
	fmt.Fprintln(w, "| --- | --- | --- | --- | --- | --- | --- | --- | --- |")
	for _, r := range rows {
		fmt.Fprintf(w, "| %s | #%d | %s | %s | %d | %d | %d | %s | %s |\n",
			r.Repo, r.Number, r.Author, r.SizeBucket,
			r.Additions, r.Deletions, r.ChangedFiles,
			fmtDur(r.TimeToFirstReview), fmtDur(r.FeatureLeadTime))
	}
}
