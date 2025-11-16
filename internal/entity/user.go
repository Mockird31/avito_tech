package entity

type UserUpdateActive struct {
	UserId   string `json:"user_id" valid:"stringlength(1|64)~user_id length 1..64"`
	IsActive bool   `json:"is_active"`
}

type User struct {
	UserId   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

type DeactivateUsers struct {
	TeamName string   `json:"team_name" valid:"stringlength(1|128)~team_name length 1..128"`
	UserIds  []string `json:"users_ids"`
}
