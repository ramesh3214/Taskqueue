package route

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ramesh3214/taskflow/controller"
	"github.com/ramesh3214/taskflow/middleware"
)

func AuthRoute(
	router *mux.Router,
	authController *controller.Authcontroller,
) {

	router.HandleFunc(
		"/auth/login",
		authController.Login,
	).Methods(http.MethodPost)

	router.HandleFunc(
		"/auth/signup",
		authController.Signup,
	).Methods(http.MethodPost)

	router.Handle(
		"/auth/delete",
		middleware.Authmiddleware(
			http.HandlerFunc(authController.DeleteAccount),
		),
	).Methods(http.MethodDelete)

	router.Handle("/auth/me", middleware.Authmiddleware(http.HandlerFunc(authController.GetProfile))).Methods(http.MethodGet)

	router.Handle("/auth/me", middleware.Authmiddleware(http.HandlerFunc(authController.UpdateProfile))).Methods(http.MethodPost)
}
