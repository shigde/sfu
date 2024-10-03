package auth

import (
	"github.com/gorilla/mux"
	"github.com/shigde/sfu/internal/auth/handler"
)

func UseRoutes(router *mux.Router, accountService *AccountService) {
	router.HandleFunc("/authenticate", handler.Authentication(accountService)).Methods("POST")
	router.HandleFunc("/auth/login", handler.Login(accountService)).Methods("POST")
	router.HandleFunc("/auth/register", handler.Register(accountService)).Methods("POST")
	router.HandleFunc("/auth/forgotPassword", handler.ForgotPassword(accountService)).Methods("POST")
	router.HandleFunc("/auth/newPassword", handler.NewPassword(accountService)).Methods("POST")
	router.HandleFunc("/auth/deleteAccount", handler.DeleteAccount(accountService)).Methods("POST")
	router.HandleFunc("/auth/verify/{token}", handler.Verification(accountService)).Methods("GET")
}
