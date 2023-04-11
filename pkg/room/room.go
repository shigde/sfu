package room

type room struct {
	id      string
	streams map[string]string
}

type RoomManager struct {
	rooms       map[string]room
	closeSignal chan struct{}
}
