package lobby

import (
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func testSessionRepositorySetup(t *testing.T) (*sessionRepository, *session) {
	t.Helper()
	repository := newSessionRepository()
	s := &session{Id: uuid.New(), user: uuid.New()}
	repository.Add(s)

	return repository, s
}
func TestSessionRepository(t *testing.T) {

	assertSession := func(t testing.TB, got *session, want *session) {
		t.Helper()
		if got != want {
			t.Errorf("got %s want %s", got.Id, want.Id)
		}
	}

	assertRepoLength := func(t testing.TB, repo *sessionRepository, want int) {
		t.Helper()

		if len(repo.All()) != want {
			t.Fatalf("wanted %d, got %d", want, len(repo.All()))
		}
	}

	assertRepoHasSession := func(t testing.TB, repo *sessionRepository, want *session) {
		t.Helper()
		want, session := repo.FindById(want.Id)
		assert.True(t, session)

		assertSession(t, want, want)
	}

	t.Run("Add Session", func(t *testing.T) {
		repo, _ := testSessionRepositorySetup(t)
		s := &session{Id: uuid.New()}
		repo.Add(s)
		assertRepoHasSession(t, repo, s)
		assertRepoLength(t, repo, 2)
	})

	t.Run("Delete Session", func(t *testing.T) {
		repo, _ := testSessionRepositorySetup(t)
		s := &session{Id: uuid.New()}
		repo.Add(s)

		if deleted := repo.Delete(s.Id); !deleted {
			t.Fatalf("deleting of rtp session fails")
		}

		assert.False(t, repo.Contains(s.Id))
	})

	t.Run("Contains Session", func(t *testing.T) {
		repo, firstSession := testSessionRepositorySetup(t)
		s := &session{Id: uuid.New()}
		repo.Add(s)

		assert.True(t, repo.Contains(s.Id))
		assert.True(t, repo.Contains(firstSession.Id))
		assert.False(t, repo.Contains(uuid.New()))
	})

	t.Run("Find Session By Id", func(t *testing.T) {
		repo, firstSession := testSessionRepositorySetup(t)
		want := &session{Id: uuid.New()}
		repo.Add(want)

		_, find := repo.FindById(firstSession.Id)
		got, _ := repo.FindById(want.Id)

		assert.True(t, find)
		assertSession(t, got, want)
	})

	t.Run("Find Session By User Id", func(t *testing.T) {
		repo, firstSession := testSessionRepositorySetup(t)
		want := &session{Id: uuid.New(), user: uuid.New()}
		repo.Add(want)

		_, find := repo.FindByUserId(firstSession.user)
		got, _ := repo.FindByUserId(want.user)

		assert.True(t, find)
		assertSession(t, got, want)
	})

	t.Run("Update Session", func(t *testing.T) {
		repo, firstSession := testSessionRepositorySetup(t)

		want, _ := repo.FindById(firstSession.Id)
		assert.True(t, repo.Update(want))

		got, _ := repo.FindById(firstSession.Id)
		assertSession(t, got, want)
	})
	t.Run("Safely Concurrently Adding and Deleting", func(t *testing.T) {
		wantedCount := 1000
		createOn := 200
		deleteOn := 500
		id := uuid.New()
		repo, _ := testSessionRepositorySetup(t)

		var wg sync.WaitGroup
		wg.Add(wantedCount + 2)
		created := make(chan struct{})

		for i := 0; i < wantedCount; i++ {
			go func() {
				repo.Add(&session{Id: uuid.New()})
				wg.Done()
			}()

			if i == createOn {
				go func() {
					repo.Add(&session{Id: id})
					assert.True(t, repo.Contains(id))
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

		// wanted + first session
		assert.Equal(t, wantedCount+1, repo.Len())
		assertRepoLength(t, repo, wantedCount+1)
		assert.False(t, repo.Contains(id))
	})

}
