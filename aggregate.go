package main

import (
	"sort"
	"time"
)

type durSummary struct {
	P50 time.Duration
	P90 time.Duration
	Max time.Duration
	N   int
}

type sizeDist struct {
	XS, S, M, L, XL, XXL int
}

type authorRow struct {
	Login    string
	PRs      int
	Adds     int
	Dels     int
	FirstRev durSummary
	LeadTime durSummary
}

type summary struct {
	Total                int
	WindowDays           int
	PerWeek              float64
	Adds                 int
	Dels                 int
	ChangedFiles         int
	TimeToFirstReview    durSummary
	FirstToLastReview    durSummary
	FirstApprovalToMerge durSummary
	FeatureLeadTime      durSummary
	Sizes                sizeDist
	ByAuthor             []authorRow
}

func percentile(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	if len(sorted) == 1 {
		return sorted[0]
	}
	idx := int(p * float64(len(sorted)-1))
	return sorted[idx]
}

func summarize(ds []time.Duration) durSummary {
	filtered := make([]time.Duration, 0, len(ds))
	for _, d := range ds {
		if d >= 0 {
			filtered = append(filtered, d)
		}
	}
	if len(filtered) == 0 {
		return durSummary{}
	}
	sort.Slice(filtered, func(i, j int) bool { return filtered[i] < filtered[j] })
	return durSummary{
		P50: percentile(filtered, 0.50),
		P90: percentile(filtered, 0.90),
		Max: filtered[len(filtered)-1],
		N:   len(filtered),
	}
}

func bucketize(rows []prRow) sizeDist {
	var d sizeDist
	for _, r := range rows {
		switch r.SizeBucket {
		case "XS":
			d.XS++
		case "S":
			d.S++
		case "M":
			d.M++
		case "L":
			d.L++
		case "XL":
			d.XL++
		case "XXL":
			d.XXL++
		}
	}
	return d
}

func authorRollup(rows []prRow) []authorRow {
	byLogin := map[string]*authorRow{}
	firstRevs := map[string][]time.Duration{}
	leadTimes := map[string][]time.Duration{}

	for _, r := range rows {
		a, ok := byLogin[r.Author]
		if !ok {
			a = &authorRow{Login: r.Author}
			byLogin[r.Author] = a
		}
		a.PRs++
		a.Adds += r.Additions
		a.Dels += r.Deletions
		firstRevs[r.Author] = append(firstRevs[r.Author], r.TimeToFirstReview)
		leadTimes[r.Author] = append(leadTimes[r.Author], r.FeatureLeadTime)
	}

	out := make([]authorRow, 0, len(byLogin))
	for login, a := range byLogin {
		a.FirstRev = summarize(firstRevs[login])
		a.LeadTime = summarize(leadTimes[login])
		out = append(out, *a)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].PRs > out[j].PRs })
	return out
}

func summarizeAll(rows []prRow, windowDays int) summary {
	s := summary{Total: len(rows), WindowDays: windowDays}
	if windowDays > 0 {
		s.PerWeek = float64(len(rows)) / (float64(windowDays) / 7.0)
	}
	firstRevs := make([]time.Duration, 0, len(rows))
	firstToLast := make([]time.Duration, 0, len(rows))
	firstApprToMerge := make([]time.Duration, 0, len(rows))
	leadTimes := make([]time.Duration, 0, len(rows))
	for _, r := range rows {
		s.Adds += r.Additions
		s.Dels += r.Deletions
		s.ChangedFiles += r.ChangedFiles
		firstRevs = append(firstRevs, r.TimeToFirstReview)
		firstToLast = append(firstToLast, r.FirstToLastReview)
		firstApprToMerge = append(firstApprToMerge, r.FirstApprovalToMerge)
		leadTimes = append(leadTimes, r.FeatureLeadTime)
	}
	s.TimeToFirstReview = summarize(firstRevs)
	s.FirstToLastReview = summarize(firstToLast)
	s.FirstApprovalToMerge = summarize(firstApprToMerge)
	s.FeatureLeadTime = summarize(leadTimes)
	s.Sizes = bucketize(rows)
	s.ByAuthor = authorRollup(rows)
	return s
}
