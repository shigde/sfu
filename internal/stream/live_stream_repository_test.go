package stream

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testLiveStreamRepositorySetup(t *testing.T) (*LiveStreamRepository, string) {
	t.Helper()
	store := newTestStore()
	repository := NewLiveStreamRepository(store)
	_ = store.db.AutoMigrate(&LiveStream{})
	s := &LiveStream{}
	streamId, _ := repository.Add(context.Background(), s)

	return repository, streamId
}
func TestLiveStreamRepository(t *testing.T) {

	assertRtpStream := func(t testing.TB, got *LiveStream, want *LiveStream) {
		t.Helper()
		assert.NotNil(t, got)
		assert.Equal(t, want.ID, got.ID)
		assert.Equal(t, want.UUID, got.UUID)
	}

	assertRepoLength := func(t testing.TB, repo *LiveStreamRepository, want int64) {
		t.Helper()

		if repo.Len(context.Background()) != want {
			t.Fatalf("wanted %d, got %d", want, repo.Len(context.Background()))
		}
	}

	assertRepoHasStream := func(t testing.TB, repo *LiveStreamRepository, want *LiveStream) {
		t.Helper()
		got, err := repo.FindByUuid(context.Background(), want.UUID.String())
		assert.NoError(t, err)
		assertRtpStream(t, want, got)
	}

	t.Run("Get not existing Stream", func(t *testing.T) {
		repo, _ := testLiveStreamRepositorySetup(t)
		stream, err := repo.FindByUuid(context.Background(), "123")
		assert.Error(t, err, ErrStreamNotFound)
		assert.Nil(t, stream)
	})

	t.Run("Add Stream", func(t *testing.T) {
		repo, _ := testLiveStreamRepositorySetup(t)
		stream := &LiveStream{}
		_, _ = repo.Add(context.Background(), stream)

		assertRepoHasStream(t, repo, stream)
		assertRepoLength(t, repo, 2)
	})

	t.Run("Delete Stream", func(t *testing.T) {
		repo, _ := testLiveStreamRepositorySetup(t)
		stream := &LiveStream{}
		id, _ := repo.Add(context.Background(), stream)

		err := repo.Delete(context.Background(), id)
		assert.NoError(t, err)
		assert.False(t, repo.Contains(context.Background(), id))
	})

	t.Run("Contains Stream", func(t *testing.T) {
		repo, streamId := testLiveStreamRepositorySetup(t)
		stream := &LiveStream{}
		id, _ := repo.Add(context.Background(), stream)

		assert.True(t, repo.Contains(context.Background(), id))
		assert.True(t, repo.Contains(context.Background(), streamId))
		assert.False(t, repo.Contains(context.Background(), "not_in_repo"))
	})

	t.Run("Find Stream By Id", func(t *testing.T) {
		repo, streamId := testLiveStreamRepositorySetup(t)
		want := &LiveStream{}
		id, _ := repo.Add(context.Background(), want)

		_, err := repo.FindByUuid(context.Background(), streamId)
		got, _ := repo.FindByUuid(context.Background(), id)

		assert.NoError(t, err)
		assertRtpStream(t, got, want)
	})

	t.Run("Update Stream", func(t *testing.T) {
		repo, streamId := testLiveStreamRepositorySetup(t)

		want, _ := repo.FindByUuid(context.Background(), streamId)
		assert.NoError(t, repo.Update(context.Background(), want))

		got, _ := repo.FindByUuid(context.Background(), streamId)
		assertRtpStream(t, got, want)
	})

	t.Run("Safely Concurrently Adding and Deleting", func(t *testing.T) {
		wantedCount := 1000
		deleteOn := 500
		repo, id := testLiveStreamRepositorySetup(t)

		var wg sync.WaitGroup
		wg.Add(wantedCount + 1)

		for i := 0; i < wantedCount; i++ {
			go func() {
				_, _ = repo.Add(context.Background(), &LiveStream{})
				wg.Done()
			}()

			if i == deleteOn {
				go func() {
					_ = repo.Delete(context.Background(), id)
					wg.Done()
				}()
			}
		}

		wg.Wait()

		assertRepoLength(t, repo, int64(wantedCount))
		assert.False(t, repo.Contains(context.Background(), id))
		assert.Equal(t, int64(wantedCount), repo.Len(context.Background()))
	})
}
