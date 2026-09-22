package util

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claim struct {
	UserID uint `json:"userID"`
	jwt.RegisteredClaims
}

func GenerateJwtToken(
	ctx context.Context,
	id uint,
) (string, error) {

	jwtsecret := os.Getenv("JWT_TOKEN")

	if jwtsecret == "" {
		return "", fmt.Errorf("JWT_TOKEN is not set")
	}

	claims := Claim{
		UserID: id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(
				time.Now().Add(15 * time.Minute),
			),
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	return token.SignedString([]byte(jwtsecret))
}