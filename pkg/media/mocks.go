package media

import (
	"github.com/shigde/sfu/pkg/lobby"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const spaceId = "abc123"

type testStore struct {
	db *gorm.DB
}

func newTestStore() *testStore {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	return &testStore{db}
}

func (s *testStore) GetDatabase() *gorm.DB {
	return s.db
}

const testOffer = ""
const testAnswer = "{\"type\":\"answer\",\"sdp\":\"\"}\n"

type testLobbyManager struct {
}

func newTestLobbyManager() *testLobbyManager {
	return &testLobbyManager{}
}

func (l *testLobbyManager) AccessLobby(id string) (*lobby.RtpStreamLobby, error) {
	return &lobby.RtpStreamLobby{Id: id}, nil
}
