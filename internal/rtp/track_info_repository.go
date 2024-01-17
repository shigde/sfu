package rtp

import (
	"sync"

	"golang.org/x/exp/slog"
)

type trackId = string

type trackInfoRepository struct {
	locker *sync.RWMutex
	infos  map[trackId]*TrackInfo
}

func newTrackInfoRepository() *trackInfoRepository {
	infos := make(map[trackId]*TrackInfo)
	return &trackInfoRepository{
		&sync.RWMutex{},
		infos,
	}
}

func (r *trackInfoRepository) Set(id trackId, info *TrackInfo) {
	r.locker.Lock()
	defer r.locker.Unlock()
	r.infos[id] = info
}

func (r *trackInfoRepository) Get(id trackId) (*TrackInfo, bool) {
	r.locker.RLock()
	defer r.locker.RUnlock()
	info, ok := r.infos[id]
	return info, ok
}

func (r *trackInfoRepository) Delete(id trackId) bool {
	r.locker.Lock()
	defer r.locker.Unlock()

	if _, ok := r.infos[id]; ok {
		delete(r.infos, id)
		return true
	}
	return false
}

func (r *trackInfoRepository) Contains(id trackId) bool {
	r.locker.RLock()
	defer r.locker.RUnlock()

	_, ok := r.infos[id]
	return ok
}

func (r *trackInfoRepository) Len() int {
	r.locker.RLock()
	defer r.locker.RUnlock()
	return len(r.infos)
}

func (r *trackInfoRepository) getTrackSdpInfos() map[trackId]TrackSdpInfo {
	r.locker.RLock()
	defer r.locker.RUnlock()
	sdpInfos := make(map[trackId]TrackSdpInfo)
	for id, trackInfo := range r.infos {
		slog.Debug("#### repo", "trackid", id)
		sdpInfos[id] = trackInfo.TrackSdpInfo
	}
	return sdpInfos
}
