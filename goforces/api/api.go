package api

import (
	"net/http"
	"oj/goforces/internal/controllers"
	"oj/goforces/internal/middlewares"
)

func SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/register", controllers.Register)
	mux.HandleFunc("/login", controllers.Login)

	mux.Handle("/profile/{username}", middlewares.AuthMiddleware(http.HandlerFunc(controllers.GetProfile)))
	mux.Handle("/profile/update", middlewares.AuthMiddleware(http.HandlerFunc(controllers.UpdateProfile)))

	mux.Handle("/admin/user", middlewares.AuthMiddleware(http.HandlerFunc(controllers.GetUserProfile)))
	mux.Handle("/admin/user/role", middlewares.AuthMiddleware(http.HandlerFunc(controllers.UpdateUserRole)))

	// TODO why problems has two handlers??
	mux.Handle("/problems", middlewares.AuthMiddleware(http.HandlerFunc(controllers.ProblemsHandler)))
	mux.Handle("/problems/{id}", middlewares.AuthMiddleware(http.HandlerFunc(controllers.GetProblemByID)))
	mux.Handle("/problems/mine", middlewares.AuthMiddleware(http.HandlerFunc(controllers.GetMyProblems)))

	mux.Handle("/admin/problems", middlewares.AuthMiddleware(http.HandlerFunc(controllers.AdminGetAllProblems)))
	mux.Handle("/admin/problems/status", middlewares.AuthMiddleware(http.HandlerFunc(controllers.AdminUpdateProblemStatus)))

	mux.Handle("/submit", middlewares.AuthMiddleware(http.HandlerFunc(controllers.CreateSubmission)))
	mux.Handle("/submissions", middlewares.AuthMiddleware(http.HandlerFunc(controllers.GetMySubmissions)))

	return mux
}
