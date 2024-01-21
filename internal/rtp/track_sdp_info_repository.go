package rtp

import (
	"sync"

	"github.com/google/uuid"
)

type trackSdpInfoId = uuid.UUID

type trackSdpInfoRepository struct {
	locker *sync.RWMutex
	infos  map[trackSdpInfoId]*TrackSdpInfo
}

func newTrackSdpInfoRepository() *trackSdpInfoRepository {
	infos := make(map[trackSdpInfoId]*TrackSdpInfo)
	return &trackSdpInfoRepository{
		&sync.RWMutex{},
		infos,
	}
}

func (r *trackSdpInfoRepository) Set(id trackSdpInfoId, info *TrackSdpInfo) {
	r.locker.Lock()
	defer r.locker.Unlock()
	r.infos[id] = info
}

func (r *trackSdpInfoRepository) Get(id trackSdpInfoId) (*TrackSdpInfo, bool) {
	r.locker.RLock()
	defer r.locker.RUnlock()
	info, ok := r.infos[id]
	return info, ok
}

func (r *trackSdpInfoRepository) Delete(id trackSdpInfoId) bool {
	r.locker.Lock()
	defer r.locker.Unlock()

	if _, ok := r.infos[id]; ok {
		delete(r.infos, id)
		return true
	}
	return false
}

func (r *trackSdpInfoRepository) Contains(id trackSdpInfoId) bool {
	r.locker.RLock()
	defer r.locker.RUnlock()

	_, ok := r.infos[id]
	return ok
}

func (r *trackSdpInfoRepository) Len() int {
	r.locker.RLock()
	defer r.locker.RUnlock()
	return len(r.infos)
}

func (r *trackSdpInfoRepository) getTrackSdpInfos() map[trackSdpInfoId]TrackSdpInfo {
	r.locker.RLock()
	defer r.locker.RUnlock()
	sdpInfos := make(map[trackSdpInfoId]TrackSdpInfo)
	for id, trackInfo := range r.infos {
		sdpInfos[id] = *trackInfo
	}
	return sdpInfos
}

func (r *trackSdpInfoRepository) getSdpInfoByEgressTrackId(egressTrackId string) (*TrackSdpInfo, bool) {
	r.locker.RLock()
	defer r.locker.RUnlock()
	for _, info := range r.infos {
		if info.EgressTrackId == egressTrackId {
			return info, true
		}
	}
	return nil, false
}

func (r *trackSdpInfoRepository) getTrackSdpInfoByIngressMid(ingressMid string) (*TrackSdpInfo, bool) {
	r.locker.RLock()
	defer r.locker.RUnlock()
	for _, info := range r.infos {
		if info.IngressMid == ingressMid {
			return info, true
		}
	}
	return nil, false
}

func (r *trackSdpInfoRepository) getSdpInfoByIngressTrackId(ingressTrackId string) (*TrackSdpInfo, bool) {
	r.locker.RLock()
	defer r.locker.RUnlock()
	for _, info := range r.infos {
		if info.IngressTrackId == ingressTrackId {
			return info, true
		}
	}
	return nil, false
}
