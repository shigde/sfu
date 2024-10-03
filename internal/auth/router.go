package auth

import (
	"github.com/gorilla/mux"
	"github.com/shigde/sfu/internal/auth/handler"
)

func UseRoutes(router *mux.Router, accountService *AccountService) {
	router.HandleFunc("/authenticate", handler.Authentication(accountService)).Methods("POST")
	router.HandleFunc("/auth/login", handler.Login(accountService)).Methods("POST")
	router.HandleFunc("/auth/register", handler.Register(accountService)).Methods("POST")
	router.HandleFunc("/auth/forgotpassword", handler.ForgotPassword(accountService)).Methods("POST")
	router.HandleFunc("/auth/newpassword", handler.NewPassword(accountService)).Methods("POST")
	router.HandleFunc("/auth/verify", handler.Verification(accountService)).Methods("POST")
}
