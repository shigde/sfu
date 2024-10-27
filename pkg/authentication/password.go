package authentication

type PasswordForgotEmail struct {
	Email string `json:"email"`
}

type PasswordForgot struct {
	Token       string `json:"token"`
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

type PasswordForgotWithToken struct {
	Token string `json:"token"`
	PasswordForgot
}
