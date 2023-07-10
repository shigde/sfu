package lobby

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testRtpStreamRepositorySetup(t *testing.T) (*rtpStreamRepository, string) {
	t.Helper()
	repository := newRtpStreamRepository()
	s := &rtpStream{}
	streamId := repository.Add(s)

	return repository, streamId
}
func TestRtpStreamRepository(t *testing.T) {

	assertRtpStream := func(t testing.TB, got *rtpStream, want *rtpStream) {
		t.Helper()
		if got != want {
			t.Errorf("got %s want %s", got.Id, want.Id)
		}
	}

	assertRepoLength := func(t testing.TB, repo *rtpStreamRepository, want int) {
		t.Helper()

		if len(repo.All()) != want {
			t.Fatalf("wanted %d, got %d", want, len(repo.All()))
		}
	}

	assertRepoHasStream := func(t testing.TB, repo *rtpStreamRepository, want *rtpStream) {
		t.Helper()
		want, hasStream := repo.FindById(want.Id)
		assert.True(t, hasStream)

		assertRtpStream(t, want, want)
	}

	t.Run("Add Stream", func(t *testing.T) {
		repo, _ := testRtpStreamRepositorySetup(t)
		stream := &rtpStream{}
		repo.Add(stream)

		assertRepoHasStream(t, repo, stream)
		assertRepoLength(t, repo, 2)
	})

	t.Run("Delete Stream", func(t *testing.T) {
		repo, _ := testRtpStreamRepositorySetup(t)
		stream := &rtpStream{}
		id := repo.Add(stream)

		if deleted := repo.Delete(id); !deleted {
			t.Fatalf("deleting of rtp stream fails")
		}

		assert.False(t, repo.Contains(id))
	})

	t.Run("Contains Stream", func(t *testing.T) {
		repo, streamId := testRtpStreamRepositorySetup(t)
		stream := &rtpStream{}
		id := repo.Add(stream)

		assert.True(t, repo.Contains(id))
		assert.True(t, repo.Contains(streamId))
		assert.False(t, repo.Contains("not_in_repo"))
	})

	t.Run("Find Stream By Id", func(t *testing.T) {
		repo, streamId := testRtpStreamRepositorySetup(t)
		want := &rtpStream{}
		id := repo.Add(want)

		_, find := repo.FindById(streamId)
		got, _ := repo.FindById(id)

		assert.True(t, find)
		assertRtpStream(t, got, want)
	})

	t.Run("Update Stream", func(t *testing.T) {
		repo, streamId := testRtpStreamRepositorySetup(t)

		want, _ := repo.FindById(streamId)
		assert.True(t, repo.Update(want))

		got, _ := repo.FindById(streamId)
		assertRtpStream(t, got, want)
	})

	t.Run("Safely Concurrently Adding and Deleting", func(t *testing.T) {
		wantedCount := 1000
		deleteOn := 500
		repo, id := testRtpStreamRepositorySetup(t)

		var wg sync.WaitGroup
		wg.Add(wantedCount + 1)

		for i := 0; i < wantedCount; i++ {
			go func() {
				repo.Add(&rtpStream{})
				wg.Done()
			}()

			if i == deleteOn {
				go func() {
					repo.Delete(id)
					wg.Done()
				}()
			}
		}

		wg.Wait()

		assertRepoLength(t, repo, wantedCount)
		assert.False(t, repo.Contains(id))
		assert.Equal(t, wantedCount, repo.Len())
	})
}
