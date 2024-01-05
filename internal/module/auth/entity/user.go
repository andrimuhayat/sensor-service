package entity

type User struct {
	Email    string `json:"email" db:"email"`
	Password string `json:"password" db:"password"`
	Role     string `json:"roles" db:"roles"`
}
