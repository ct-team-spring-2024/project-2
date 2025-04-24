package middlewares

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/sirupsen/logrus"
)

type contextKey string

const (
	userContextKey = contextKey("userId")
	jwtKey         = "your_secret_key"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		logrus.Info("Came here1")
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtKey), nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		userIdStr, ok := claims["sub"].(string)
		if !ok {
			http.Error(w, "Invalid token subject", http.StatusUnauthorized)

			return
		}

		userId, err := strconv.Atoi(userIdStr)
		if err != nil {
			http.Error(w, "Invalid user id in token", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), userContextKey, userId)
		logrus.Info(userId)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserIDFromContext(ctx context.Context) (int, bool) {
	userId, ok := ctx.Value(userContextKey).(int)
	return userId, ok
}
