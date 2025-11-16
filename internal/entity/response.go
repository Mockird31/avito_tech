package entity

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type TeamResponse struct {
	Team *Team `json:"team"`
}

type UserResponse struct {
	User *User `json:"user"`
}

type PullRequestResponse struct {
	PullRequest *PullRequest `json:"pr"`
}

type PullRequestReassignResponse struct {
	PullRequest *PullRequest `json:"pr"`
	ReplacedBy  string       `json:"replaced_by"`
}

type ReviewerPullRequests struct {
	UserId       string              `json:"user_id"`
	PullRequests []*PullRequestShort `json:"pull_requests"`
}
