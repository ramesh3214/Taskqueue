package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ramesh3214/taskflow/dto"
	"github.com/ramesh3214/taskflow/model"
	"github.com/ramesh3214/taskflow/queue"
	"github.com/ramesh3214/taskflow/repository"
	"github.com/ramesh3214/taskflow/util"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	authRepo *repository.AuthRepo
	redis    *redis.Client
	taskRepo *repository.Taskrepo

	// Redis queue removed.
	// RabbitMQ is now responsible for background tasks.
	rabbitMQ *queue.RabbitMQ
}

func NewAuthService(
	redis *redis.Client,
	authRepo *repository.AuthRepo,
	taskRepo *repository.Taskrepo,
	rabbitMQ *queue.RabbitMQ,
) *AuthService {
	return &AuthService{
		authRepo: authRepo,
		redis:    redis,
		taskRepo: taskRepo,
		rabbitMQ: rabbitMQ,
	}
}

func (s *AuthService) Login(
	ctx context.Context,
	email string,
	password string,
) (dto.Loginres, error) {

	user, err := s.authRepo.Getuserbyemail(ctx, email)
	if err != nil {
		return dto.Loginres{}, errors.New(
			"invalid username or password",
		)
	}

	fmt.Println("Password from request:", password)
	fmt.Println("Password from DB:", user.Password)

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(password),
	)

	if err != nil {
		fmt.Println("bcrypt error:", err)

		return dto.Loginres{}, errors.New(
			"invalid username or password",
		)
	}

	token, err := util.GenerateJwtToken(ctx, user.ID)
	if err != nil {
		return dto.Loginres{}, fmt.Errorf(
			"failed to generate token: %w",
			err,
		)
	}

	return dto.Loginres{
		Id:       user.ID,
		Name:     user.Name,
		Email:    user.Email,
		Age:      user.Age,
		JwtToken: token,
	}, nil
}

func (s *AuthService) Signup(
	ctx context.Context,
	newSignup dto.Signup,
) (dto.Profiledata, error) {

	// -------------------------
	// Hash password
	// -------------------------

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(newSignup.Password),
		bcrypt.DefaultCost,
	)

	if err != nil {
		return dto.Profiledata{}, fmt.Errorf(
			"failed to hash password: %w",
			err,
		)
	}

	newSignup.Password = string(hashedPassword)

	// -------------------------
	// Create user
	// -------------------------

	result, err := s.authRepo.CreateNewUser(
		ctx,
		newSignup,
	)

	if err != nil {
		return dto.Profiledata{}, fmt.Errorf(
			"failed to create user: %w",
			err,
		)
	}

	// -------------------------
	// Create background task
	// -------------------------

	task := &model.Task{
		UserID:     result.ID,
		Email:      result.Email,
		TaskType:   "WELCOME_EMAIL",
		Status:     "PENDING",
		RetryCount: 0,
	}

	err = s.taskRepo.CreateTask(
		ctx,
		task,
	)

	if err != nil {
		return dto.Profiledata{}, fmt.Errorf(
			"failed to create email task: %w",
			err,
		)
	}

	taskMessage := dto.TaskMessage{
		TaskID:   task.ID,
		TaskType: task.TaskType,
	}

	message, err := json.Marshal(taskMessage)

	if err != nil {
		return dto.Profiledata{}, fmt.Errorf(
			"failed to create task message: %w",
			err,
		)
	}

	err = s.rabbitMQ.Publish(
		"task_exchange",
		"task.created",
		string(message),
	)

	if err != nil {
		return dto.Profiledata{}, fmt.Errorf(
			"failed to publish task to RabbitMQ: %w",
			err,
		)
	}

	fmt.Printf(
		"Task %d published to RabbitMQ\n",
		task.ID,
	)

	return result, nil
}

func (s *AuthService) UpdateProfile(
	ctx context.Context,
	id uint,
	updateProfileData dto.Updateprofile,
) (dto.Profiledata, error) {

	result, err := s.authRepo.UpdateData(
		ctx,
		id,
		updateProfileData,
	)

	if err != nil {
		return dto.Profiledata{}, fmt.Errorf(
			"failed to update profile: %w",
			err,
		)
	}

	// Redis is still used for profile cache.
	key := fmt.Sprintf(
		"user:profile:%d",
		id,
	)

	err = s.redis.Del(
		ctx,
		key,
	).Err()

	if err != nil {
		fmt.Printf(
			"failed to delete profile cache: %v\n",
			err,
		)
	}

	return result, nil
}

func (s *AuthService) DeleteProfile(
	ctx context.Context,
	id uint,
) (string, error) {

	result, err := s.authRepo.DeleteByID(
		ctx,
		id,
	)

	if err != nil {
		return "", fmt.Errorf(
			"failed to delete profile: %w",
			err,
		)
	}

	// Redis is still used for profile cache.
	key := fmt.Sprintf(
		"user:profile:%d",
		id,
	)

	err = s.redis.Del(
		ctx,
		key,
	).Err()

	if err != nil {
		fmt.Printf(
			"failed to delete profile cache: %v\n",
			err,
		)
	}

	return result, nil
}

func (s *AuthService) GetProfile(
	ctx context.Context,
	id uint,
) (dto.Profiledata, error) {

	// -------------------------
	// Check Redis cache
	// -------------------------

	key := fmt.Sprintf(
		"user:profile:%d",
		id,
	)

	cachedData, err := s.redis.Get(
		ctx,
		key,
	).Result()

	if err == nil {

		var user dto.Profiledata

		err := json.Unmarshal(
			[]byte(cachedData),
			&user,
		)

		if err != nil {
			return dto.Profiledata{}, fmt.Errorf(
				"failed to decode cached profile: %w",
				err,
			)
		}

		return user, nil
	}

	if err != redis.Nil {
		fmt.Printf(
			"redis error: %v\n",
			err,
		)
	}

	// -------------------------
	// Cache MISS → PostgreSQL
	// -------------------------

	result, err := s.authRepo.GetByID(
		ctx,
		id,
	)

	if err != nil {
		return dto.Profiledata{}, fmt.Errorf(
			"failed to get profile: %w",
			err,
		)
	}

	// -------------------------
	// Save profile in Redis
	// -------------------------

	data, err := json.Marshal(result)

	if err != nil {
		return dto.Profiledata{}, fmt.Errorf(
			"failed to encode profile: %w",
			err,
		)
	}

	err = s.redis.Set(
		ctx,
		key,
		data,
		10*time.Minute,
	).Err()

	if err != nil {
		fmt.Printf(
			"failed to save profile in redis: %v\n",
			err,
		)
	}

	return result, nil
}
