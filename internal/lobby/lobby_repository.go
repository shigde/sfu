package lobby

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/storage"
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

func (r *lobbyRepository) getOrCreateLobby(liveStreamId uuid.UUID) (*lobby, error) {
	r.locker.Lock()
	defer r.locker.Unlock()
	currentLobby, ok := r.lobbies[liveStreamId]
	if !ok {
		lobby := newLobby(liveStreamId, r.rtpEngine)
		entity, err := r.queryLobbyEntity(context.Background(), liveStreamId.String())
		if err != nil {
			return nil, fmt.Errorf("fetching lobby entity: %w", err)
		}

		entity.IsRunning = true
		if entity, err = r.updateLobbyEntity(context.Background(), entity); err != nil {
			return nil, fmt.Errorf("updating lobby entity as running: %w", err)
		}

		lobby.entity = entity
		r.lobbies[liveStreamId] = lobby
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

func (r *lobbyRepository) Delete(id uuid.UUID) bool {
	r.locker.Lock()
	defer r.locker.Unlock()
	if lobby, ok := r.lobbies[id]; ok {
		lobby.entity.IsRunning = false
		_, _ = r.updateLobbyEntity(context.Background(), lobby.entity)
		delete(r.lobbies, id)
		return true
	}
	return false
}

func (r *lobbyRepository) Len() int {
	r.locker.RLock()
	defer r.locker.RUnlock()
	return len(r.lobbies)
}

func (r *lobbyRepository) queryLobbyEntity(ctx context.Context, liveStreamId string) (*LobbyEntity, error) {
	tx, cancel := r.store.GetDatabaseWithContext(ctx)
	defer cancel()

	lobby := &LobbyEntity{}
	result := tx.Where("live_stream_id=?", liveStreamId).Find(lobby)
	if result.Error != nil {
		err := fmt.Errorf("finding lobby for stream %s: %w", liveStreamId, result.Error)
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
