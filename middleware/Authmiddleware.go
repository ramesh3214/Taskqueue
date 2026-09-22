package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ramesh3214/taskflow/util"
)

func Authmiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			http.Error(
				w,
				"unauthorized",
				http.StatusUnauthorized,
			)
			return
		}

		tokenString := strings.TrimPrefix(
			authHeader,
			"Bearer ",
		)

		claims := &util.Claim{}

		jwtsecret := os.Getenv("JWT_TOKEN")

		if jwtsecret == "" {
			http.Error(
				w,
				"JWT secret is not configured",
				http.StatusInternalServerError,
			)
			return
		}

		token, err := jwt.ParseWithClaims(
			tokenString,
			claims,
			func(token *jwt.Token) (interface{}, error) {

				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf(
						"unexpected signing method: %v",
						token.Header["alg"],
					)
				}

				return []byte(jwtsecret), nil
			},
		)

		if err != nil || !token.Valid {
			fmt.Println("JWT error:", err)

			http.Error(
				w,
				"unauthorized",
				http.StatusUnauthorized,
			)
			return
		}

		ctx := context.WithValue(
			r.Context(),
			"userID",
			claims.UserID,
		)

		next.ServeHTTP(
			w,
			r.WithContext(ctx),
		)
	})
}