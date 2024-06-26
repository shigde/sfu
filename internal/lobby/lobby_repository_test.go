package lobby

import (
	"context"
	"net/url"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/lobby/sessions"
	"github.com/shigde/sfu/internal/storage"
	"github.com/stretchr/testify/assert"
)

func testLobbyRepositorySetup(t *testing.T) *lobbyRepository {
	t.Helper()

	// When a random lobby is created, there is no entity in the store.
	// With this, We ensure that the entity URL is the same as the lobby URL, and no connector is started.
	// @TODO build test lobby with entity
	homeActorIri, _ := url.Parse("")
	store := storage.NewTestStore()
	_ = store.GetDatabase().AutoMigrate(&LobbyEntity{Host: homeActorIri.String()})

	var engine sessions.RtpEngine
	repository := newLobbyRepository(store, engine, homeActorIri, "test-key")

	return repository
}
func TestLobbyRepository(t *testing.T) {

	t.Run("Get not existing lobby", func(t *testing.T) {
		repo := testLobbyRepositorySetup(t)
		lobby, ok := repo.getLobby(uuid.New())
		assert.False(t, ok)
		assert.Nil(t, lobby)
	})

	t.Run("Create lobby", func(t *testing.T) {
		repo := testLobbyRepositorySetup(t)
		lobby, _ := repo.getOrCreateLobby(context.Background(), uuid.New(), make(chan lobbyItem))
		assert.NotNil(t, lobby)
		lobby.stop()
	})

	t.Run("Create and Get lobby", func(t *testing.T) {
		repo := testLobbyRepositorySetup(t)
		id := uuid.New()
		lobbyCreated, _ := repo.getOrCreateLobby(context.Background(), id, make(chan lobbyItem))

		assert.NotNil(t, lobbyCreated)
		lobbyGet, ok := repo.getLobby(id)
		assert.True(t, ok)
		assert.Same(t, lobbyCreated, lobbyGet)
		lobbyCreated.stop()
	})

	t.Run("Delete lobby", func(t *testing.T) {
		repo := testLobbyRepositorySetup(t)
		id := uuid.New()
		created, _ := repo.getOrCreateLobby(context.Background(), id, make(chan lobbyItem))
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
		repo := testLobbyRepositorySetup(t)

		var wg sync.WaitGroup
		wg.Add(wantedCount + 2)
		created := make(chan struct{})

		for i := 0; i < wantedCount; i++ {
			go func(id int) {
				lobby, _ := repo.getOrCreateLobby(context.Background(), uuid.New(), make(chan lobbyItem))
				assert.NotNil(t, lobby)
				wg.Done()
			}(i)

			if i == createOn {
				go func() {
					lobby, _ := repo.getOrCreateLobby(context.Background(), id, make(chan lobbyItem))
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
