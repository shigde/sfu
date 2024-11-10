package authentication

type ForgotPasswordEmail struct {
	Email string `json:"email"`
}

type UpdatePassword struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

type ForgotPassword struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}
