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
	router.HandleFunc("/auth/newPassword", handler.NewPassword(accountService)).Methods("POST")
	router.HandleFunc("/auth/deleteAccount", session.HttpMiddleware(accountService.GetConfig(), handler.DeleteAccount(accountService))).Methods("POST")
	router.HandleFunc("/auth/getAccount", session.HttpMiddleware(accountService.GetConfig(), handler.GetAccount(accountService))).Methods("GET")
	router.HandleFunc("/auth/verify", handler.Verification(accountService)).Methods("PUT")
}
