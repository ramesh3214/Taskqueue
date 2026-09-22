package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/ramesh3214/taskflow/dto"
	"github.com/ramesh3214/taskflow/model"
	"gorm.io/gorm"
)

type AuthRepo struct {
	db *gorm.DB
}

func NewAuthRepo(db *gorm.DB) *AuthRepo {
	return &AuthRepo{
		db: db,
	}
}


func (repo *AuthRepo) CreateNewUser(
	ctx context.Context,
	newUser dto.Signup,
) (dto.Profiledata, error) {

	var user model.Auth

	err := repo.db.WithContext(ctx).
		Where("email = ?", newUser.Email).
		First(&user).Error

	if err == nil {
		return dto.Profiledata{}, fmt.Errorf("user already exists")
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.Profiledata{}, fmt.Errorf(
			"failed to check existing user: %w",
			err,
		)
	}

	user.Age = newUser.Age
	user.Email = newUser.Email
	user.Name = newUser.Name
	user.Password = newUser.Password

	if err := repo.db.WithContext(ctx).Create(&user).Error; err != nil {
		return dto.Profiledata{}, fmt.Errorf(
			"failed to create user: %w",
			err,
		)
	}

	return dto.Profiledata{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Age:   user.Age,
	}, nil
}


func (repo *AuthRepo) GetByID(
	ctx context.Context,
	id uint,
) (dto.Profiledata, error) {

	var user model.Auth

	err := repo.db.WithContext(ctx).
		First(&user, id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.Profiledata{}, fmt.Errorf("user not found")
	}

	if err != nil {
		return dto.Profiledata{}, fmt.Errorf(
			"failed to get user: %w",
			err,
		)
	}

	return dto.Profiledata{
		ID:user.ID,
		Name:  user.Name,
		Email: user.Email,
		Age:   user.Age,
	}, nil
}


func (repo *AuthRepo) UpdateData(
	ctx context.Context,
	id uint,
	update dto.Updateprofile,
) (dto.Profiledata, error) {

	var user model.Auth

	
	err := repo.db.WithContext(ctx).
		First(&user, id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.Profiledata{}, fmt.Errorf("user not found")
	}

	if err != nil {
		return dto.Profiledata{}, fmt.Errorf(
			"failed to get user: %w",
			err,
		)
	}

	// Update only allowed fields.
	user.Name = update.Name
	user.Age = update.Age

	//Save changes
	err = repo.db.WithContext(ctx).
		Save(&user).Error

	if err != nil {
		return dto.Profiledata{}, fmt.Errorf(
			"failed to update user: %w",
			err,
		)
	}

	return dto.Profiledata{
		ID:user.ID,
		Name:  user.Name,
		Email: user.Email,
		Age:   user.Age,
	}, nil
}

func (repo *AuthRepo) DeleteByID(ctx context.Context, id uint) (string, error) {
	var user model.Auth

	result := repo.db.WithContext(ctx).Delete(&user, id)

	if result.Error != nil {
		return "", fmt.Errorf("failed to delete the account: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return "", fmt.Errorf("account not found")
	}

	return "account has been deleted successfully", nil
}


func (repo *AuthRepo) Getuserbyemail(
	ctx context.Context,
	email string,
) (*model.Auth, error) {

	var user model.Auth

	err := repo.db.
		WithContext(ctx).
		Where("email = ?", email).
		First(&user).Error

	if err != nil {
		return nil, err
	}

	return &user, nil
}