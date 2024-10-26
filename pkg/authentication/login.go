package authentication

type Login struct {
	Email string `json:"email"`
	Pass  string `json:"pass"`
}
