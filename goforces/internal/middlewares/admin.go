package middlewares

import (
	"net/http"
	"oj/goforces/internal/services"
)

func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			http.Error(w, "User not authenticated", http.StatusUnauthorized)
			return
		}
		user, err := services.GetUserByID(userID)
		if err != nil || user.Role != "admin" {
			http.Error(w, "Admin access required", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
