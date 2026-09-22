package validator

import (
	"fmt"

	"github.com/ramesh3214/taskflow/dto"
)

func SigninValidator(s dto.Login) (error) {

	if s.Email == "" {
		return  fmt.Errorf("email is required")
	}

	if s.Password == "" {
		return  fmt.Errorf("password is required")
	}

	return nil
}

func SignupValidator(s dto.Signup) ( error) {

	if s.Name == "" {
		return  fmt.Errorf("name is required")
	}

	if s.Email == "" {
		return  fmt.Errorf("email is required")
	}

	if s.Password == "" {
		return  fmt.Errorf("password is required")
	}

	return nil
}
