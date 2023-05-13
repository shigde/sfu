package engine

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testSetup(t *testing.T) (*RtpStreamRepository, string) {
	t.Helper()
	repository := NewRtpStreamRepository()
	s := &RtpStream{}
	streamId := repository.Add(s)

	return repository, streamId
}
func TestRtpStream(t *testing.T) {

	assertRtpStream := func(t testing.TB, got *RtpStream, want *RtpStream) {
		t.Helper()
		if got != want {
			t.Errorf("got %s want %s", got.Id, want.Id)
		}
	}

	assertRepoLength := func(t testing.TB, repo *RtpStreamRepository, want int) {
		t.Helper()

		if len(repo.All()) != want {
			t.Fatalf("wanted %d, got %d", want, len(repo.All()))
		}
	}

	assertRepoHasStream := func(t testing.TB, repo *RtpStreamRepository, want *RtpStream) {
		t.Helper()
		want, hasStream := repo.FindById(want.Id)
		assert.True(t, hasStream)

		assertRtpStream(t, want, want)
	}

	t.Run("Add Stream", func(t *testing.T) {
		repo, _ := testSetup(t)
		stream := &RtpStream{}
		repo.Add(stream)

		assertRepoHasStream(t, repo, stream)
		assertRepoLength(t, repo, 2)
	})

	t.Run("Delete Stream", func(t *testing.T) {
		repo, _ := testSetup(t)
		stream := &RtpStream{}
		id := repo.Add(stream)

		if deleted := repo.Delete(id); !deleted {
			t.Fatalf("deleting of rtp stream fails")
		}

		assert.False(t, repo.Contains(id))
	})

	t.Run("Contains Stream", func(t *testing.T) {
		repo, streamId := testSetup(t)
		stream := &RtpStream{}
		id := repo.Add(stream)

		assert.True(t, repo.Contains(id))
		assert.True(t, repo.Contains(streamId))
		assert.False(t, repo.Contains("not_in_repo"))
	})

	t.Run("Find Stream By Id", func(t *testing.T) {
		repo, streamId := testSetup(t)
		want := &RtpStream{}
		id := repo.Add(want)

		_, find := repo.FindById(streamId)
		got, _ := repo.FindById(id)

		assert.True(t, find)
		assertRtpStream(t, got, want)
	})

	t.Run("Update Stream", func(t *testing.T) {
		repo, streamId := testSetup(t)

		want, _ := repo.FindById(streamId)
		assert.True(t, repo.Update(want))

		got, _ := repo.FindById(streamId)
		assertRtpStream(t, got, want)
	})

	t.Run("Safely Concurrently Adding and Deleting", func(t *testing.T) {
		wantedCount := 1000
		deleteOn := 500
		repo, id := testSetup(t)

		var wg sync.WaitGroup
		wg.Add(wantedCount + 1)

		for i := 0; i < wantedCount; i++ {
			go func() {
				repo.Add(&RtpStream{})
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
	})
}
