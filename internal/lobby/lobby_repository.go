package lobby

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/metric"
	"github.com/shigde/sfu/internal/storage"
	"golang.org/x/exp/slog"
	"gorm.io/gorm"
)

var errLobbyNotFound = errors.New("lobby not found")

type lobbyRepository struct {
	locker    *sync.RWMutex
	lobbies   map[uuid.UUID]*lobby
	store     storage.Storage
	rtpEngine rtpEngine
}

func newLobbyRepository(store storage.Storage, rtpEngine rtpEngine) *lobbyRepository {
	lobbies := make(map[uuid.UUID]*lobby)
	return &lobbyRepository{
		&sync.RWMutex{},
		lobbies,
		store,
		rtpEngine,
	}
}

func (r *lobbyRepository) getOrCreateLobby(ctx context.Context, lobbyId uuid.UUID, lobbyGarbageCollector chan<- uuid.UUID) (*lobby, error) {
	r.locker.Lock()
	defer r.locker.Unlock()
	currentLobby, ok := r.lobbies[lobbyId]
	if !ok {
		lobby := newLobby(lobbyId, r.rtpEngine, lobbyGarbageCollector)
		entity, err := r.queryLobbyEntity(ctx, lobbyId.String())
		if err != nil {
			return nil, fmt.Errorf("fetching lobby entity: %w", err)
		}

		entity.IsRunning = true
		if entity, err = r.updateLobbyEntity(ctx, entity); err != nil {
			return nil, fmt.Errorf("updating lobby entity as running: %w", err)
		}

		lobby.entity = entity
		r.lobbies[lobbyId] = lobby

		metric.RunningLobbyInc(lobby.entity.LiveStreamId.String(), lobbyId.String())
		return lobby, nil
	}
	return currentLobby, nil
}

func (r *lobbyRepository) getLobby(id uuid.UUID) (*lobby, bool) {
	r.locker.Lock()
	defer r.locker.Unlock()
	currentLobby, ok := r.lobbies[id]
	return currentLobby, ok
}

func (r *lobbyRepository) setLobbyLive(ctx context.Context, id uuid.UUID, isLive bool) bool {
	r.locker.Lock()
	defer r.locker.Unlock()
	if currentLobby, ok := r.lobbies[id]; ok {
		currentLobby.entity.IsLive = isLive
		_, err := r.updateLobbyEntity(ctx, currentLobby.entity)
		return err != nil
	}
	return false
}

func (r *lobbyRepository) delete(ctx context.Context, id uuid.UUID) bool {
	r.locker.Lock()
	defer r.locker.Unlock()
	if lobby, ok := r.lobbies[id]; ok {
		// Dealing with races.
		// If the lobby still in use then we can not delete
		if lobby.sessions.Len() > 0 {
			return false
		}
		lobby.entity.IsRunning = false
		if _, err := r.updateLobbyEntity(ctx, lobby.entity); err != nil {
			slog.Error("can not update lobby entity on delete lobby", "err", err, "lobby", id)
			return false
		}
		delete(r.lobbies, id)
		lobby.stop()
		metric.RunningLobbyDec(lobby.entity.LiveStreamId.String(), id.String())
		return true
	}
	return false
}

func (r *lobbyRepository) Len() int {
	r.locker.RLock()
	defer r.locker.RUnlock()
	return len(r.lobbies)
}

func (r *lobbyRepository) queryLobbyEntity(ctx context.Context, lobbyId string) (*LobbyEntity, error) {
	tx, cancel := r.store.GetDatabaseWithContext(ctx)
	defer cancel()

	lobby := &LobbyEntity{}
	result := tx.Where("uuid=?", lobbyId).Find(lobby)
	if result.Error != nil {
		err := fmt.Errorf("finding lobby for stream %s: %w", lobbyId, result.Error)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.Join(err, errLobbyNotFound)
		}
		return nil, err
	}
	return lobby, nil
}

func (r *lobbyRepository) updateLobbyEntity(ctx context.Context, lobby *LobbyEntity) (*LobbyEntity, error) {
	tx, cancel := r.store.GetDatabaseWithContext(ctx)
	defer cancel()

	result := tx.Save(lobby)
	if result.Error != nil {
		return nil, fmt.Errorf("updating lobby %s: %w", lobby.UUID, result.Error)
	}
	return lobby, nil
}
