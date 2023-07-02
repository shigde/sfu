package stream

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type testStore struct {
	db *gorm.DB
}

func (s *testStore) GetDatabase() *gorm.DB {
	return s.db
}

func testSpaceRepositorySetup(t *testing.T) *SpaceRepository {
	t.Helper()
	var lobby lobbyAccessor
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	store := &testStore{db}
	repository, _ := newSpaceRepository(lobby, store)
	return repository
}
func TestSpaceRepository(t *testing.T) {

	t.Run("Get not existing Space", func(t *testing.T) {
		repo := testSpaceRepositorySetup(t)
		space, ok := repo.GetSpace(context.Background(), "123")
		assert.False(t, ok)
		assert.Nil(t, space)
	})

	t.Run("Create Space", func(t *testing.T) {
		repo := testSpaceRepositorySetup(t)
		space := repo.GetOrCreateSpace(context.Background(), "456")
		assert.NotNil(t, space)
	})

	t.Run("Create and Get Space", func(t *testing.T) {
		repo := testSpaceRepositorySetup(t)
		spaceCreated := repo.GetOrCreateSpace(context.Background(), "789")
		assert.NotNil(t, spaceCreated)
		spaceGet, ok := repo.GetSpace(context.Background(), "789")
		assert.True(t, ok)
		assert.NotSame(t, spaceCreated, spaceGet)
		assert.Equal(t, spaceCreated, spaceGet)
	})

	t.Run("Delete Space", func(t *testing.T) {
		repo := testSpaceRepositorySetup(t)
		spaceCreated := repo.GetOrCreateSpace(context.Background(), "1012")
		assert.NotNil(t, spaceCreated)

		deleted := repo.Delete(context.Background(), "1012")
		assert.True(t, deleted)

		spaceGet, ok := repo.GetSpace(context.Background(), "1012")
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
		created := make(chan struct{})

		for i := 0; i < wantedCount; i++ {
			go func(spaceId int) {
				space := repo.GetOrCreateSpace(context.Background(), fmt.Sprintf("id-%d", spaceId))
				assert.NotNil(t, space)
				wg.Done()
			}(i)

			if i == createOn {
				go func() {
					space := repo.GetOrCreateSpace(context.Background(), id)
					assert.NotNil(t, space)
					close(created)
					wg.Done()
				}()
			}

			if i == deleteOn {
				go func() {
					<-created
					deleted := repo.Delete(context.Background(), id)
					assert.True(t, deleted)
					wg.Done()
				}()
			}
		}

		wg.Wait()

		_, ok := repo.GetSpace(context.Background(), id)
		assert.False(t, ok)
		assert.Equal(t, int64(wantedCount), repo.Len(context.Background()))
	})
}
