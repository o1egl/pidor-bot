package domain

type User struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	IsActive  bool   `json:"is_active"`
}

func (u User) Mention() string {
	if u.Username != "" {
		return "@" + u.Username
	}
	return u.FirstName + " " + u.LastName
}
