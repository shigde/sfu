package authentication

type PasswordForget struct {
	Token       string `json:"token"`
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

type PasswordForgetWithToken struct {
	Token string `json:"token"`
	PasswordForget
}
