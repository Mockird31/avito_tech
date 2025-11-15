package entity

type UserUpdateActive struct {
	UserId   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

type User struct {
	UserId   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}
