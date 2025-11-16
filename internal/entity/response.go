package entity

type ErrorResponse struct {
	Error *Error `json:"error"`
}

type Error struct {
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

type AssignmentStatsResponse struct {
	Statistics []*UserAssignmentCount `json:"statistics"`
}

type DeactivateUsersResponse struct {
	DeactivateUsers *DeactivateUsers `json:"deactivate_users"`
}
