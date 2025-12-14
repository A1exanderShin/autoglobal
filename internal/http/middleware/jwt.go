package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

type contextKey string

const userIDKey contextKey = "user_id"

func JWTAuth(jwtKey []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method")
				}
				return jwtKey, nil
			})

			if err != nil || !token.Valid {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			userIDFloat, ok := claims["user_id"].(float64)
			if !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			userID := int64(userIDFloat)

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
