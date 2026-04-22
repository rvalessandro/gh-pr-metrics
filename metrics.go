package main

import (
	"time"
)

type prRow struct {
	Repo                 string
	Number               int
	Author               string
	Commits              int
	Additions            int
	Deletions            int
	ChangedFiles         int
	Comments             int
	Participants         int
	MergedAt             time.Time
	CreatedAt            time.Time
	TimeToFirstReview    time.Duration // -1 == n/a
	FirstToLastReview    time.Duration
	FirstApprovalToMerge time.Duration
	FeatureLeadTime      time.Duration
	SizeBucket           string
}

const naDuration = time.Duration(-1)

func parse(t string) time.Time {
	v, _ := time.Parse(time.RFC3339, t)
	return v
}

// readyOrCreated returns the ready-for-review timestamp if present,
// otherwise the PR creation time. Safe against GitHub's quirk where a
// connection totalCount can ignore itemTypes filtering.
func readyOrCreated(created time.Time, ti timelineItems) time.Time {
	if len(ti.Nodes) == 0 {
		return created
	}
	if ti.Nodes[0].ReadyForReviewEvent.CreatedAt == "" {
		return created
	}
	return parse(ti.Nodes[0].ReadyForReviewEvent.CreatedAt)
}

func timeToFirstReview(p pullRequestNode) time.Duration {
	if len(p.Reviews.Nodes) == 0 {
		return naDuration
	}
	ready := readyOrCreated(parse(p.CreatedAt), p.TimelineItems)
	for _, r := range p.Reviews.Nodes {
		if r.Author.Login == p.Author.Login {
			continue
		}
		first := parse(r.CreatedAt)
		return first.Sub(ready)
	}
	return naDuration
}

func firstToLastReview(p pullRequestNode) time.Duration {
	var first, last time.Time
	foundFirst := false
	for _, r := range p.Reviews.Nodes {
		if r.Author.Login == p.Author.Login {
			continue
		}
		t := parse(r.CreatedAt)
		if !foundFirst {
			first = t
			foundFirst = true
		}
		if r.State == "APPROVED" {
			last = t
		}
	}
	if !foundFirst || last.IsZero() {
		return naDuration
	}
	return last.Sub(first)
}

func firstApprovalToMerge(p pullRequestNode) time.Duration {
	if p.MergedAt == "" {
		return naDuration
	}
	merged := parse(p.MergedAt)
	for _, r := range p.Reviews.Nodes {
		if r.Author.Login == p.Author.Login {
			continue
		}
		if r.State != "APPROVED" {
			continue
		}
		return merged.Sub(parse(r.CreatedAt))
	}
	return naDuration
}

func featureLeadTime(p pullRequestNode) time.Duration {
	if p.MergedAt == "" || len(p.Commits.Nodes) == 0 {
		return naDuration
	}
	merged := parse(p.MergedAt)
	var earliest time.Time
	for i, c := range p.Commits.Nodes {
		t := parse(c.Commit.CommittedDate)
		if i == 0 || t.Before(earliest) {
			earliest = t
		}
	}
	return merged.Sub(earliest)
}

func sizeBucket(additions, deletions int) string {
	total := additions + deletions
	switch {
	case total < 10:
		return "XS"
	case total < 50:
		return "S"
	case total < 250:
		return "M"
	case total < 1000:
		return "L"
	case total < 5000:
		return "XL"
	default:
		return "XXL"
	}
}

func rowFromPR(p pullRequestNode) prRow {
	return prRow{
		Repo:                 p.Repository.NameWithOwner,
		Number:               p.Number,
		Author:               p.Author.Login,
		Commits:              p.Commits.TotalCount,
		Additions:            p.Additions,
		Deletions:            p.Deletions,
		ChangedFiles:         p.ChangedFiles,
		Comments:             p.Comments.TotalCount,
		Participants:         p.Participants.TotalCount,
		MergedAt:             parse(p.MergedAt),
		CreatedAt:            parse(p.CreatedAt),
		TimeToFirstReview:    timeToFirstReview(p),
		FirstToLastReview:    firstToLastReview(p),
		FirstApprovalToMerge: firstApprovalToMerge(p),
		FeatureLeadTime:      featureLeadTime(p),
		SizeBucket:           sizeBucket(p.Additions, p.Deletions),
	}
}
