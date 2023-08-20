package lobby

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/logging"
	"github.com/stretchr/testify/assert"
)

func testLobbyManagerSetup(t *testing.T) (*LobbyManager, *lobby, uuid.UUID) {
	t.Helper()
	logging.SetupDebugLogger()
	engine := mockRtpEngineForOffer(mockedAnswer)

	manager := NewLobbyManager(engine)
	lobby := manager.lobbies.getOrCreateLobby(uuid.New())
	user := uuid.New()
	session := newSession(user, lobby.hub, engine, onQuitSessionInternallyStub)
	session.sender = newSenderHandler(session.Id, user, nil)
	session.sender.endpoint = mockConnection(mockedAnswer)
	lobby.sessions.Add(session)

	return manager, lobby, user
}

func TestLobbyManager(t *testing.T) {
	t.Run("Access a Lobby with timeout", func(t *testing.T) {
		manager, lobby, _ := testLobbyManagerSetup(t)
		userId := uuid.New()
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // trigger cancel for time out
		_, err := manager.AccessLobby(ctx, lobby.Id, userId, mockedOffer)
		assert.ErrorIs(t, err, errLobbyRequestTimeout)
	})

	t.Run("Access a new Lobby", func(t *testing.T) {
		manager, lobby, _ := testLobbyManagerSetup(t)
		userId := uuid.New()
		data, err := manager.AccessLobby(context.Background(), lobby.Id, userId, mockedOffer)

		assert.NoError(t, err)
		assert.Equal(t, mockedAnswer, data.Answer)
		assert.False(t, uuid.Nil == data.RtpSessionId)
		assert.False(t, uuid.Nil == data.Resource)
	})

	t.Run("Access a already started Lobby", func(t *testing.T) {
		manager, lobby, _ := testLobbyManagerSetup(t)

		_, err := manager.AccessLobby(context.Background(), lobby.Id, uuid.New(), mockedOffer)
		assert.NoError(t, err)

		data, err := manager.AccessLobby(context.Background(), lobby.Id, uuid.New(), mockedOffer)
		assert.NoError(t, err)

		assert.Equal(t, mockedAnswer, data.Answer)
		assert.False(t, uuid.Nil == data.RtpSessionId)
		assert.False(t, uuid.Nil == data.Resource)
	})

	t.Run("Start listen to a Lobby with timeout", func(t *testing.T) {
		manager, lobby, user := testLobbyManagerSetup(t)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // trigger cancel for time out
		_, err := manager.StartListenLobby(ctx, lobby.Id, user)
		assert.ErrorIs(t, err, errLobbyRequestTimeout)
	})

	t.Run("Start listen to a Lobby, but receiver has no messenger", func(t *testing.T) {
		manager, lobby, _ := testLobbyManagerSetup(t)
		user := uuid.New()
		session := newSession(user, lobby.hub, mockRtpEngineForOffer(mockedAnswer), onQuitSessionInternallyStub)
		session.receiver = newReceiverHandler(session.Id, session.user, nil)
		lobby.sessions.Add(session)

		_, err := manager.StartListenLobby(context.Background(), lobby.Id, user)
		assert.ErrorIs(t, err, errReceiverInSessionHasNoMessenger)
	})

	t.Run("Start listen to a Lobby, but no session exists", func(t *testing.T) {
		manager, lobby, _ := testLobbyManagerSetup(t)

		_, err := manager.StartListenLobby(context.Background(), lobby.Id, uuid.New())
		assert.ErrorIs(t, err, errNoSession)
	})

	t.Run("Start listen to a Lobby", func(t *testing.T) {
		manager, lobby, _ := testLobbyManagerSetup(t)
		user := uuid.New()
		session := newSession(user, lobby.hub, mockRtpEngineForOffer(mockedAnswer), onQuitSessionInternallyStub)
		session.receiver = newReceiverHandler(session.Id, session.user, nil)
		session.receiver.messenger = newMessenger(newSendMock(t))
		lobby.sessions.Add(session)

		data, err := manager.StartListenLobby(context.Background(), lobby.Id, user)
		assert.NoError(t, err)
		assert.Equal(t, mockedAnswer, data.Offer)
		assert.False(t, uuid.Nil == data.RtpSessionId)
	})

	t.Run("Start listen to a Lobby but session already listen", func(t *testing.T) {
		manager, lobby, user := testLobbyManagerSetup(t)

		_, err := manager.StartListenLobby(context.Background(), lobby.Id, user)
		assert.ErrorIs(t, err, errSenderInSessionAlreadyExists)
	})

	t.Run("listen to a Lobby with timeout", func(t *testing.T) {
		manager, lobby, user := testLobbyManagerSetup(t)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // trigger cancel for time out
		_, err := manager.ListenLobby(ctx, lobby.Id, user, mockedAnswer)
		assert.ErrorIs(t, err, errLobbyRequestTimeout)
	})

	t.Run("listen to a Lobby", func(t *testing.T) {
		manager, lobby, user := testLobbyManagerSetup(t)
		data, err := manager.ListenLobby(context.Background(), lobby.Id, user, mockedAnswer)
		assert.NoError(t, err)
		assert.True(t, data.Active)
		assert.False(t, uuid.Nil == data.RtpSessionId)
	})

	t.Run("listen to a Lobby but no session", func(t *testing.T) {
		manager, lobby, _ := testLobbyManagerSetup(t)
		_, err := manager.ListenLobby(context.Background(), lobby.Id, uuid.New(), mockedAnswer)
		assert.ErrorIs(t, err, errNoSession)
	})

	t.Run("leave a Lobby but no session", func(t *testing.T) {
		manager, lobby, _ := testLobbyManagerSetup(t)
		_, err := manager.LeaveLobby(context.Background(), lobby.Id, uuid.New())
		assert.ErrorIs(t, err, errNoSession)
	})

	t.Run("Leave a Lobby", func(t *testing.T) {
		manager, lobby, user := testLobbyManagerSetup(t)
		success, err := manager.LeaveLobby(context.Background(), lobby.Id, user)
		assert.NoError(t, err)
		assert.True(t, success)
	})
}
