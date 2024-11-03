package authentication

type User struct {
	Name   string `json:"name"`
	Domain string `json:"domain"`
	Role   string `json:"role"`
}
