package dto

type TaskMessage struct {
	TaskID   uint   `json:"task_id"`
	TaskType string `json:"task_type"`
}