package authentication

type ClientUser struct {
	// user-name@domian
	UserId string `json:"user"`
	// Instance Token
	Token string `json:"token"`
	// Not checked currently
	Client string `json:"client"`
}
