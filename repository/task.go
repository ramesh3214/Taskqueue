package repository

import (
	"context"

	"github.com/ramesh3214/taskflow/model"
	"gorm.io/gorm"
)

type Taskrepo struct {
	db *gorm.DB
}

func CreateNewTask(db *gorm.DB) *Taskrepo {
	return &Taskrepo{
		db: db,
	}
}

func (t *Taskrepo) CreateTask(ctx context.Context, task *model.Task) error {

	err := t.db.WithContext(ctx).Create(task).Error

	if err != nil {
		return err
	}

	return nil

}

func (repo *Taskrepo) GetTaskByID(
	ctx context.Context,
	id uint,
) (*model.Task, error) {
	var task model.Task

	err := repo.db.WithContext(ctx).First(&task, id).Error
	if err != nil {
		return nil, err
	}

	return &task, nil
}

func (repo *Taskrepo) UpdateTaskByID(
	ctx context.Context,
	id uint,
	status string,
) error {

	var task model.Task

	err := repo.db.WithContext(ctx).
		Model(&task).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status": status,
		}).Error

	if err != nil {
		return err
	}

	return nil
}


func (repo *Taskrepo) IncrementRetryCount(
    ctx context.Context,
    id uint,
) (*model.Task, error) {

    var task model.Task

    err := repo.db.WithContext(ctx).
        First(&task, id).Error

    if err != nil {
        return nil, err
    }

    task.RetryCount++

    err = repo.db.WithContext(ctx).
        Save(&task).Error

    if err != nil {
        return nil, err
    }

    return &task, nil
}
