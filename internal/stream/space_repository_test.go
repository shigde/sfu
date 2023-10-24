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

func testSpaceRepositorySetup(t *testing.T) *SpaceRepository {
	t.Helper()
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	_ = db.AutoMigrate(&Space{})
	store := &testStore{db}
	repository := NewSpaceRepository(store)
	return repository
}
func TestSpaceRepository(t *testing.T) {

	t.Run("Get not existing Space", func(t *testing.T) {
		repo := testSpaceRepositorySetup(t)
		space, err := repo.GetByIdentifier(context.Background(), "123")
		assert.Error(t, err, ErrSpaceNotFound)
		assert.Nil(t, space)
	})

	t.Run("Create Space", func(t *testing.T) {
		repo := testSpaceRepositorySetup(t)
		space, err := repo.CreateWithIdentifier(context.Background(), "456")
		assert.NoError(t, err)
		assert.NotNil(t, space)
	})

	t.Run("Create and Get Space", func(t *testing.T) {
		repo := testSpaceRepositorySetup(t)
		spaceCreated, _ := repo.CreateWithIdentifier(context.Background(), "789")
		assert.NotNil(t, spaceCreated)
		spaceGet, err := repo.GetByIdentifier(context.Background(), "789")
		assert.NoError(t, err)
		assert.NotSame(t, spaceCreated, spaceGet)
		assert.Equal(t, spaceCreated.Identifier, spaceGet.Identifier)
	})

	t.Run("Delete Space", func(t *testing.T) {
		repo := testSpaceRepositorySetup(t)
		spaceCreated, _ := repo.CreateWithIdentifier(context.Background(), "1012")
		assert.NotNil(t, spaceCreated)

		err := repo.Delete(context.Background(), "1012")
		assert.NoError(t, err)

		spaceGet, err := repo.GetByIdentifier(context.Background(), "1012")
		assert.Error(t, err)
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
				space, _ := repo.CreateWithIdentifier(context.Background(), fmt.Sprintf("id-%d", spaceId))
				assert.NotNil(t, space)
				wg.Done()
			}(i)

			if i == createOn {
				go func() {
					space, _ := repo.CreateWithIdentifier(context.Background(), id)
					assert.NotNil(t, space)
					close(created)
					wg.Done()
				}()
			}

			if i == deleteOn {
				go func() {
					<-created
					err := repo.Delete(context.Background(), id)
					assert.NoError(t, err)
					wg.Done()
				}()
			}
		}

		wg.Wait()

		_, err := repo.GetByIdentifier(context.Background(), id)
		assert.Error(t, err, ErrSpaceNotFound)
		assert.Equal(t, int64(wantedCount), repo.Len(context.Background()))
	})
}
