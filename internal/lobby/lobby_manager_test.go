package lobby

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/logging"
	"github.com/shigde/sfu/internal/storage"
	"github.com/stretchr/testify/assert"
)

func testLobbyManagerSetup(t *testing.T) (*LobbyManager, *lobby, uuid.UUID) {
	t.Helper()
	logging.SetupDebugLogger()
	store := storage.NewTestStore()
	_ = store.GetDatabase().AutoMigrate(&LobbyEntity{})

	engine := mockRtpEngineForOffer(mockedAnswer)

	manager := NewLobbyManager(store, engine)
	lobby, _ := manager.lobbies.getOrCreateLobby(context.Background(), uuid.New(), make(chan uuid.UUID))
	user := uuid.New()
	session := newSession(user, lobby.hub, engine, nil)
	session.signal.messenger = newMockedMessenger(t)
	session.egress = mockConnection(mockedAnswer)
	session.signal.egressEndpoint = session.egress

	lobby.sessions.Add(session)

	return manager, lobby, user
}

func TestLobbyManager(t *testing.T) {
	t.Run("Create lobby ingestion endpoint with timeout", func(t *testing.T) {
		manager, lobby, _ := testLobbyManagerSetup(t)
		userId := uuid.New()
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // trigger cancel for time out
		_, err := manager.CreateLobbyIngressEndpoint(ctx, lobby.Id, userId, mockedOffer)
		assert.ErrorIs(t, err, errLobbyRequestTimeout)
	})

	t.Run("Create new lobby with ingestion endpoint", func(t *testing.T) {
		manager, lobby, _ := testLobbyManagerSetup(t)
		userId := uuid.New()
		data, err := manager.CreateLobbyIngressEndpoint(context.Background(), lobby.Id, userId, mockedOffer)

		assert.NoError(t, err)
		assert.Equal(t, mockedAnswer, data.Answer)
		assert.False(t, uuid.Nil == data.RtpSessionId)
		assert.False(t, uuid.Nil == data.Resource)
	})

	t.Run("Create ingestion endpoint when lobby already started", func(t *testing.T) {
		manager, lobby, _ := testLobbyManagerSetup(t)

		_, err := manager.CreateLobbyIngressEndpoint(context.Background(), lobby.Id, uuid.New(), mockedOffer)
		assert.NoError(t, err)

		data, err := manager.CreateLobbyIngressEndpoint(context.Background(), lobby.Id, uuid.New(), mockedOffer)
		assert.NoError(t, err)

		assert.Equal(t, mockedAnswer, data.Answer)
		assert.False(t, uuid.Nil == data.RtpSessionId)
		assert.False(t, uuid.Nil == data.Resource)
	})

	t.Run("Start listen to a lobby with timeout", func(t *testing.T) {
		manager, lobby, user := testLobbyManagerSetup(t)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // trigger cancel for time out
		_, err := manager.InitLobbyEgressEndpoint(ctx, lobby.Id, user)
		assert.ErrorIs(t, err, errLobbyRequestTimeout)
	})

	t.Run("Start listen to a lobby, but receiver has no messenger", func(t *testing.T) {
		manager, lobby, _ := testLobbyManagerSetup(t)
		user := uuid.New()
		session := newSession(user, lobby.hub, mockRtpEngineForOffer(mockedAnswer), nil)
		session.ingress = mockConnection(nil)
		session.signal.messenger = newMockedMessenger(t)
		lobby.sessions.Add(session)

		oldTimeOut := waitingTimeOut
		waitingTimeOut = 0
		_, err := manager.InitLobbyEgressEndpoint(context.Background(), lobby.Id, user)
		assert.ErrorIs(t, err, errTimeoutByWaitingForMessenger)
		waitingTimeOut = oldTimeOut
	})

	t.Run("Start listen to a lobby, but no session exists", func(t *testing.T) {
		manager, lobby, _ := testLobbyManagerSetup(t)

		_, err := manager.InitLobbyEgressEndpoint(context.Background(), lobby.Id, uuid.New())
		assert.ErrorIs(t, err, errNoSession)
	})

	t.Run("Start listen to a lobby", func(t *testing.T) {
		manager, lobby, _ := testLobbyManagerSetup(t)
		user := uuid.New()
		session := newSession(user, lobby.hub, mockRtpEngineForOffer(mockedAnswer), nil)
		session.ingress = mockConnection(nil)
		session.signal.messenger = newMockedMessenger(t)
		session.signal.stopWaitingForMessenger()
		lobby.sessions.Add(session)

		data, err := manager.InitLobbyEgressEndpoint(context.Background(), lobby.Id, user)
		assert.NoError(t, err)
		assert.Equal(t, mockedAnswer, data.Offer)
		assert.False(t, uuid.Nil == data.RtpSessionId)
	})

	t.Run("Start listen to a lobby but session already listen", func(t *testing.T) {
		manager, lobby, user := testLobbyManagerSetup(t)

		_, err := manager.InitLobbyEgressEndpoint(context.Background(), lobby.Id, user)
		assert.ErrorIs(t, err, errSenderInSessionAlreadyExists)
	})

	t.Run("listen to a lobby with timeout", func(t *testing.T) {
		manager, lobby, user := testLobbyManagerSetup(t)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // trigger cancel for time out
		_, err := manager.FinalCreateLobbyEgressEndpoint(ctx, lobby.Id, user, mockedAnswer)
		assert.ErrorIs(t, err, errLobbyRequestTimeout)
	})

	t.Run("listen to a lobby", func(t *testing.T) {
		manager, lobby, user := testLobbyManagerSetup(t)
		data, err := manager.FinalCreateLobbyEgressEndpoint(context.Background(), lobby.Id, user, mockedAnswer)
		assert.NoError(t, err)
		assert.True(t, data.Active)
		assert.False(t, uuid.Nil == data.RtpSessionId)
	})

	t.Run("listen to a lobby but no session", func(t *testing.T) {
		manager, lobby, _ := testLobbyManagerSetup(t)
		_, err := manager.FinalCreateLobbyEgressEndpoint(context.Background(), lobby.Id, uuid.New(), mockedAnswer)
		assert.ErrorIs(t, err, errNoSession)
	})

	t.Run("leave a lobby but no session", func(t *testing.T) {
		manager, lobby, _ := testLobbyManagerSetup(t)
		_, err := manager.LeaveLobby(context.Background(), lobby.Id, uuid.New())
		assert.ErrorIs(t, err, errNoSession)
	})

	t.Run("Leave a lobby", func(t *testing.T) {
		manager, lobby, user := testLobbyManagerSetup(t)
		success, err := manager.LeaveLobby(context.Background(), lobby.Id, user)
		assert.NoError(t, err)
		assert.True(t, success)
	})
}
