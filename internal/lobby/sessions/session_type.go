package sessions

type SessionType int

const (
	// UserSession represents the connection of a user directly connected to this instance
	UserSession SessionType = iota + 1
	// InstanceSession represents the connection to the Shig instance hosting the live stream.
	// When the livestream belongs to another Shig instance, this Shig instance connects to that remote instance.
	InstanceSession
	// RemoteInstanceSession represents the connection of another Shig instance.
	RemoteInstanceSession
)
