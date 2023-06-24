package lobby

type Role struct {
	name string
}

func NewGuest() *Role {
	return &Role{"guest"}
}

func NewOwner() *Role {
	return &Role{"owner"}
}

func (r *Role) isOwner() bool {
	return r.name == "owner"
}
