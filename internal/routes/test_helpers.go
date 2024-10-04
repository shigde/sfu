package routes

import (
	"context"
	"net/url"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/shigde/sfu/internal/activitypub/models"
	"github.com/shigde/sfu/internal/auth/account"
	"github.com/shigde/sfu/internal/auth/session"
	"github.com/shigde/sfu/internal/lobby"
	"github.com/shigde/sfu/internal/routes/mocks"
	"github.com/shigde/sfu/internal/storage"
	"github.com/shigde/sfu/internal/stream"
)

func testRouterSetup(t *testing.T) (*testHelper, *stream.Space, *stream.LiveStream, *account.Account, string) {
	t.Helper()

	lobbyManager := mocks.NewLobbyManager()
	store := storage.NewTestStore()
	_ = store.GetDatabase().AutoMigrate(&stream.LiveStream{}, &stream.Space{}, &account.Account{})

	streamRepo := stream.NewLiveStreamRepository(store)
	spaceRepo := stream.NewSpaceRepository(store)
	accountRepo := account.NewAccountRepository(store)
	instanceUrl, _ := url.Parse("http://shig.de")

	liveStreamService := stream.NewLiveStreamService(streamRepo, spaceRepo)
	liveLobbyService := stream.NewLiveLobbyService(store, lobbyManager)
	accountService := account.NewAccountService(accountRepo, "test-token", instanceUrl, mocks.SecurityConfig, nil)

	account := &account.Account{}
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
	liveStream.Title = "TestStream"
	liveStream.Video = &models.Video{Name: "TestStream"}
	liveStream.User = "testUser@test.de"
	liveStream.Account = account
	liveStream.AccountId = account.ID
	liveStream.Space = space
	liveStream.SpaceId = space.ID
	liveStream.Lobby = lobby.NewLobbyEntity(liveStream.UUID, space.Identifier, "http://localhost:1234/federation/accounts/shig-test")

	_, _ = streamRepo.Add(context.Background(), liveStream)

	bearer, _ := session.CreateJWTToken(account.UUID, mocks.SecurityConfig.JWT)
	bearer = "Bearer " + bearer

	th := &testHelper{}
	th.router = NewRouter(mocks.SecurityConfig, mocks.RtpConfig, accountService, liveStreamService, liveLobbyService)
	th.liveStreamRepo = streamRepo
	return th, space, liveStream, account, bearer
}

type testHelper struct {
	router         *mux.Router
	liveStreamRepo *stream.LiveStreamRepository
}
