package authentication

type User struct {
	// user-name@domian
	UserId string `json:"user"`
	// Instance Token
	Token string `json:"token"`
	// Not checked currently
	Client string `json:"client"`
}
