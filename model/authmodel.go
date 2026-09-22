package model

type Auth struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Age int `json:"age"`
}