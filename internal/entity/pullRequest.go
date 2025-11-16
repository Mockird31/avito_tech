package entity

import "time"

type StatusPr string

const (
	StatusOpen   StatusPr = "OPEN"
	StatusMerged StatusPr = "MERGED"
)

func (sp StatusPr) String() string {
	switch sp {
	case StatusOpen:
		return "OPEN"
	case StatusMerged:
		return "MERGED"
	}
	return ""
}

type PullRequest struct {
	Id                   string     `json:"pull_request_id" valid:"stringlength(1|64)~id length 1..64"`
	PrName               string     `json:"pull_request_name" valid:"stringlength(1|256)~name length 1..256"`
	AuthorId             string     `json:"author_id" valid:"stringlength(1|64)~author_id length 1..64"`
	Status               string     `json:"status" valid:"in(OPEN|MERGED)~invalid status"`
	AssignedReviewersIds []string   `json:"assigned_reviewers"`
	MergedAt             *time.Time `json:"mergedAt,omitempty"`
}

type PullRequestShort struct {
	Id       string     `json:"pull_request_id"`
	PrName   string     `json:"pull_request_name"`
	AuthorId string     `json:"author_id"`
	Status   string     `json:"status"`
	MergedAt *time.Time `json:"mergedAt,omitempty"`
}

type PullRequestReassignRequest struct {
    Id            string `json:"pull_request_id" valid:"stringlength(1|64)~pull_request_id length 1..64"`
    OldReviewerId string `json:"old_reviewer_id" valid:"stringlength(1|64)~old_reviewer_id length 1..64"`
}

type ReviewerPullRequests struct {
	UserId       string              `json:"user_id"`
	PullRequests []*PullRequestShort `json:"pull_requests"`
}
