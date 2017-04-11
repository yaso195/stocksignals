package model

type User struct {
	ID       int    `json:"id" db:"id"`
	Email    string `json:"email" db:"email"`
	Password string `json:"password" db:"password"`
}
