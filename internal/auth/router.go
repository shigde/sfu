package auth

import (
	"github.com/gorilla/mux"
	"github.com/shigde/sfu/internal/auth/account"
	"github.com/shigde/sfu/internal/auth/handler"
	"github.com/shigde/sfu/internal/auth/session"
)

func UseRoutes(router *mux.Router, accountService *account.AccountService) {
	router.HandleFunc("/authenticate", handler.Authentication(accountService)).Methods("POST")
	router.HandleFunc("/auth/login", handler.Login(accountService)).Methods("POST")
	router.HandleFunc("/auth/register", handler.Register(accountService)).Methods("POST")
	router.HandleFunc("/auth/forgotPassword", handler.ForgotPassword(accountService)).Methods("POST")
	router.HandleFunc("/auth/newPassword", handler.NewPasswordByForgot(accountService)).Methods("PUT")
	router.HandleFunc("/auth/updatePassword", session.HttpMiddleware(accountService.GetConfig(), handler.NewPasswordByLogin(accountService))).Methods("PUT")
	router.HandleFunc("/auth/deleteAccount", session.HttpMiddleware(accountService.GetConfig(), handler.DeleteAccount(accountService))).Methods("POST")
	router.HandleFunc("/auth/user", session.HttpMiddleware(accountService.GetConfig(), handler.GetUser(accountService))).Methods("GET")
	router.HandleFunc("/auth/verify", handler.Verification(accountService)).Methods("PUT")
}
