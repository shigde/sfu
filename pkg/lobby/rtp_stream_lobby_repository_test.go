package lobby

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testRtpStreamLobbyRepositorySetup(t *testing.T) *RtpStreamLobbyRepository {
	t.Helper()
	repository := newRtpStreamLobbyRepository()

	return repository
}
func TestSpaceRepository(t *testing.T) {

	t.Run("Get not existing Lobby", func(t *testing.T) {
		repo := testRtpStreamLobbyRepositorySetup(t)
		space, ok := repo.GetLobby("123")
		assert.False(t, ok)
		assert.Nil(t, space)
	})

	t.Run("Create Lobby", func(t *testing.T) {
		repo := testRtpStreamLobbyRepositorySetup(t)
		lobby := repo.GetOrCreateLobby("456")
		assert.NotNil(t, lobby)
	})

	t.Run("Create and Get Lobby", func(t *testing.T) {
		repo := testRtpStreamLobbyRepositorySetup(t)
		lobbyCreated := repo.GetOrCreateLobby("789")
		assert.NotNil(t, lobbyCreated)
		lobbyGet, ok := repo.GetLobby("789")
		assert.True(t, ok)
		assert.Same(t, lobbyCreated, lobbyGet)
	})

	t.Run("Delete Lobby", func(t *testing.T) {
		repo := testRtpStreamLobbyRepositorySetup(t)
		created := repo.GetOrCreateLobby("1012")
		assert.NotNil(t, created)

		deleted := repo.Delete("1012")
		assert.True(t, deleted)

		get, ok := repo.GetLobby("1012")
		assert.False(t, ok)
		assert.Nil(t, get)
	})

	t.Run("Safely Concurrently Adding and Deleting", func(t *testing.T) {
		wantedCount := 1000
		createOn := 200
		deleteOn := 500
		id := "abc"
		repo := testRtpStreamLobbyRepositorySetup(t)

		var wg sync.WaitGroup
		wg.Add(wantedCount + 2)
		created := make(chan struct{})

		for i := 0; i < wantedCount; i++ {
			go func(id int) {
				lobby := repo.GetOrCreateLobby(fmt.Sprintf("id-%d", id))
				assert.NotNil(t, lobby)
				wg.Done()
			}(i)

			if i == createOn {
				go func() {
					lobby := repo.GetOrCreateLobby(id)
					assert.NotNil(t, lobby)
					close(created)
					wg.Done()
				}()
			}

			if i == deleteOn {
				go func() {
					<-created
					deleted := repo.Delete(id)
					assert.True(t, deleted)
					wg.Done()
				}()
			}
		}

		wg.Wait()

		_, ok := repo.GetLobby(id)
		assert.False(t, ok)
		assert.Equal(t, wantedCount, repo.Len())
	})
}
