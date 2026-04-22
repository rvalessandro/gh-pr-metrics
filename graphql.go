package main

type pageInfo struct {
	HasNextPage bool
	EndCursor   string
}

type author struct {
	Login    string
	Typename string `graphql:"__typename"`
}

type reviewNode struct {
	Author    author
	CreatedAt string
	State     string
}

type reviews struct {
	TotalCount int
	Nodes      []reviewNode
}

type commit struct {
	CommittedDate string
}

type commitNode struct {
	Commit commit
}

type commits struct {
	TotalCount int
	Nodes      []commitNode
}

type readyForReviewEvent struct {
	CreatedAt string
}

type timelineItemNode struct {
	ReadyForReviewEvent readyForReviewEvent `graphql:"... on ReadyForReviewEvent"`
}

type timelineItems struct {
	Nodes []timelineItemNode
}

type participants struct{ TotalCount int }
type comments struct{ TotalCount int }

type pullRequestNode struct {
	Author        author
	Additions     int
	Deletions     int
	Number        int
	CreatedAt     string
	ChangedFiles  int
	IsDraft       bool
	MergedAt      string
	Repository    struct{ NameWithOwner string }
	Participants  participants
	Comments      comments
	Reviews       reviews       `graphql:"reviews(first: 100, states: [APPROVED, CHANGES_REQUESTED, COMMENTED])"`
	Commits       commits       `graphql:"commits(first: 100)"`
	TimelineItems timelineItems `graphql:"timelineItems(first: 1, itemTypes: [READY_FOR_REVIEW_EVENT])"`
}

type searchNode struct {
	PullRequest pullRequestNode `graphql:"... on PullRequest"`
}

type metricsQuery struct {
	Search struct {
		PageInfo pageInfo
		Nodes    []searchNode
	} `graphql:"search(query: $query, type: ISSUE, first: $resultCount, after: $afterCursor)"`
}
