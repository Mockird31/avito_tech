package entity

import "time"

type StatusPr int

const (
	StatusOpen StatusPr = iota
	StatusMerged
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
	Id                   string     `json:"pull_request_id"`
	PrName               string     `json:"pull_request_name"`
	AuthorId             string     `json:"author_id"`
	Status               string     `json:"status"`
	AssignedReviewersIds []string   `json:"assigned_reviewers"`
	MergedAt             *time.Time `json:"mergedAt,omitempty"`
}
