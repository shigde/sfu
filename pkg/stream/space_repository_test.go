package stream

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testSpaceRepositorySetup(t *testing.T) *SpaceRepository {
	t.Helper()
	var lobby lobbyAccessor
	repository := newSpaceRepository(lobby)

	return repository
}
func TestSpaceRepository(t *testing.T) {

	t.Run("Get not existing Space", func(t *testing.T) {
		repo := testSpaceRepositorySetup(t)
		space, ok := repo.GetSpace("123")
		assert.False(t, ok)
		assert.Nil(t, space)
	})

	t.Run("Create Space", func(t *testing.T) {
		repo := testSpaceRepositorySetup(t)
		space := repo.GetOrCreateSpace("456")
		assert.NotNil(t, space)
	})

	t.Run("Create and Get Space", func(t *testing.T) {
		repo := testSpaceRepositorySetup(t)
		spaceCreated := repo.GetOrCreateSpace("789")
		assert.NotNil(t, spaceCreated)
		spaceGet, ok := repo.GetSpace("789")
		assert.True(t, ok)
		assert.Same(t, spaceCreated, spaceGet)
	})

	t.Run("Delete Space", func(t *testing.T) {
		repo := testSpaceRepositorySetup(t)
		spaceCreated := repo.GetOrCreateSpace("1012")
		assert.NotNil(t, spaceCreated)

		deleted := repo.Delete("1012")
		assert.True(t, deleted)

		spaceGet, ok := repo.GetSpace("1012")
		assert.False(t, ok)
		assert.Nil(t, spaceGet)
	})

	t.Run("Safely Concurrently Adding and Deleting", func(t *testing.T) {
		wantedCount := 1000
		createOn := 200
		deleteOn := 500
		id := "abc"
		repo := testSpaceRepositorySetup(t)

		var wg sync.WaitGroup
		wg.Add(wantedCount + 2)

		for i := 0; i < wantedCount; i++ {
			go func(id int) {
				space := repo.GetOrCreateSpace(fmt.Sprintf("id-%d", id))
				assert.NotNil(t, space)
				wg.Done()
			}(i)

			if i == createOn {
				go func() {
					space := repo.GetOrCreateSpace(id)
					assert.NotNil(t, space)
					wg.Done()
				}()
			}

			if i == deleteOn {
				go func() {
					deleted := repo.Delete(id)
					assert.True(t, deleted)
					wg.Done()
				}()
			}
		}

		wg.Wait()

		_, ok := repo.GetSpace(id)
		assert.False(t, ok)
		assert.Equal(t, wantedCount, repo.Len())
	})
}
