package entity

type TeamMember struct {
	UserID   string `json:"user_id" valid:"stringlength(1|64)~user_id length 1..64"`
	Username string `json:"username" valid:"stringlength(1|128)~username length 1..128"`
	IsActive bool   `json:"is_active"`
}

type Team struct {
	TeamName string        `json:"team_name" valid:"stringlength(1|128)~team_name length 1..128"`
	Members  []*TeamMember `json:"members"`
}
