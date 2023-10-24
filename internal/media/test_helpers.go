package media

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/shigde/sfu/internal/auth"
	"github.com/shigde/sfu/internal/lobby"
	"github.com/shigde/sfu/internal/storage"
	"github.com/shigde/sfu/internal/stream"
)

func testRouterSetup(t *testing.T) (*testHelper, *stream.Space, *stream.LiveStream, *auth.Account, string) {
	t.Helper()

	lobbyManager := newTestLobbyManager()
	store := storage.NewTestStore()
	_ = store.GetDatabase().AutoMigrate(&stream.LiveStream{}, &stream.Space{}, &auth.Account{})

	streamRepo := stream.NewLiveStreamRepository(store)
	spaceRepo := stream.NewSpaceRepository(store)
	accountRepo := auth.NewAccountRepository(store)

	liveStreamService := stream.NewLiveStreamService(streamRepo, spaceRepo)
	liveLobbyService := stream.NewLiveLobbyService(store, lobbyManager)
	accountService := auth.NewAccountService(accountRepo, "test-token", securityConfig)

	account := &auth.Account{}
	account.UUID = uuid.NewString()
	account.User = "testUser@test.de"
	_, _ = accountRepo.Add(context.Background(), account)

	space := &stream.Space{}
	space.Account = account
	space.AccountId = account.ID
	space.Identifier = uuid.NewString()
	_, _ = spaceRepo.Add(context.Background(), space)

	liveStream := &stream.LiveStream{}
	liveStream.UUID = uuid.New()
	liveStream.User = "testUser@test.de"
	liveStream.Account = account
	liveStream.AccountId = account.ID
	liveStream.Space = space
	liveStream.SpaceId = space.ID
	liveStream.Lobby = lobby.NewLobbyEntity(liveStream.UUID)

	_, _ = streamRepo.Add(context.Background(), liveStream)

	bearer, _ := auth.CreateJWTToken(account.UUID, securityConfig.JWT)
	bearer = "Bearer " + bearer

	th := &testHelper{}
	th.router = NewRouter(securityConfig, rtpConfig, accountService, liveStreamService, liveLobbyService)
	th.liveStreamRepo = streamRepo
	return th, space, liveStream, account, bearer
}

type testHelper struct {
	router         *mux.Router
	liveStreamRepo *stream.LiveStreamRepository
}
