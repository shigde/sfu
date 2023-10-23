package authentication

type User struct {
	UserId string `json:"user"`
	Token  string `json:"token"`
	Client string `json:"client"`
}
