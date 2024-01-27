package lobby

import (
	"context"
	"net/url"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/storage"
	"github.com/stretchr/testify/assert"
)

func testRtpStreamLobbyRepositorySetup(t *testing.T) *lobbyRepository {
	t.Helper()
	store := storage.NewTestStore()
	_ = store.GetDatabase().AutoMigrate(&LobbyEntity{})
	var engine rtpEngine
	host, _ := url.Parse("http://localhost:1234/federation/accounts/shig-test")
	repository := newLobbyRepository(store, engine, host)

	return repository
}
func TestStreamLobbyRepository(t *testing.T) {

	t.Run("Get not existing lobby", func(t *testing.T) {
		repo := testRtpStreamLobbyRepositorySetup(t)
		lobby, ok := repo.getLobby(uuid.New())
		assert.False(t, ok)
		assert.Nil(t, lobby)
	})

	t.Run("Create lobby", func(t *testing.T) {
		repo := testRtpStreamLobbyRepositorySetup(t)
		lobby, _ := repo.getOrCreateLobby(context.Background(), uuid.New(), make(chan uuid.UUID))
		assert.NotNil(t, lobby)
		lobby.stop()
	})

	t.Run("Create and Get lobby", func(t *testing.T) {
		repo := testRtpStreamLobbyRepositorySetup(t)
		id := uuid.New()
		lobbyCreated, _ := repo.getOrCreateLobby(context.Background(), id, make(chan uuid.UUID))

		assert.NotNil(t, lobbyCreated)
		lobbyGet, ok := repo.getLobby(id)
		assert.True(t, ok)
		assert.Same(t, lobbyCreated, lobbyGet)
		lobbyCreated.stop()
	})

	t.Run("Delete lobby", func(t *testing.T) {
		repo := testRtpStreamLobbyRepositorySetup(t)
		id := uuid.New()
		created, _ := repo.getOrCreateLobby(context.Background(), id, make(chan uuid.UUID))
		assert.NotNil(t, created)

		deleted := repo.delete(context.Background(), id)
		assert.True(t, deleted)

		get, ok := repo.getLobby(id)
		assert.False(t, ok)
		assert.Nil(t, get)
	})

	t.Run("Safely Concurrently Adding and Deleting", func(t *testing.T) {
		wantedCount := 1000
		createOn := 200
		deleteOn := 500
		id := uuid.New()
		repo := testRtpStreamLobbyRepositorySetup(t)

		var wg sync.WaitGroup
		wg.Add(wantedCount + 2)
		created := make(chan struct{})

		for i := 0; i < wantedCount; i++ {
			go func(id int) {
				lobby, _ := repo.getOrCreateLobby(context.Background(), uuid.New(), make(chan uuid.UUID))
				assert.NotNil(t, lobby)
				wg.Done()
			}(i)

			if i == createOn {
				go func() {
					lobby, _ := repo.getOrCreateLobby(context.Background(), id, make(chan uuid.UUID))
					assert.NotNil(t, lobby)
					close(created)
					wg.Done()
				}()
			}

			if i == deleteOn {
				go func() {
					<-created
					deleted := repo.delete(context.Background(), id)
					assert.True(t, deleted)
					wg.Done()
				}()
			}
		}

		wg.Wait()

		_, ok := repo.getLobby(id)
		assert.False(t, ok)
		assert.Equal(t, wantedCount, repo.Len())

		for _, savedLobby := range repo.lobbies {
			savedLobby.stop()
		}
	})
}
