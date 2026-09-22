package model

type Task struct {
	ID       uint   `gorm:"primaryKey"`
	UserID   uint
	Email    string
	TaskType string
	RetryCount  int
	Status   string
}