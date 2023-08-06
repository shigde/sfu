// Package storage This memory storage is from: https://github.com/gofiber/fiber/blob/master/internal/memory/
package storage

import (
	"sync"
	"sync/atomic"
	"time"
)

var (
	timestampTimer sync.Once
	// timestamp please start the timer function before you use this value
	// please load the value with atomic `atomic.LoadUint32(&utils.timestamp)`
	timestamp uint32
)

type Memory struct {
	sync.RWMutex
	data map[string]item // data
}

type item struct {
	// max value is 4294967295 -> Sun Feb 07 2106 06:28:15 GMT+0000
	e uint32      // exp
	v interface{} // val
}

func NewMemory() *Memory {
	store := &Memory{
		data: make(map[string]item),
	}
	startTimeStampUpdater()
	go store.gc(1 * time.Second)
	return store
}

// Get value by key
func (s *Memory) Get(key string) interface{} {
	s.RLock()
	v, ok := s.data[key]
	s.RUnlock()
	if !ok || v.e != 0 && v.e <= atomic.LoadUint32(&timestamp) {
		return nil
	}
	return v.v
}

// Set key with value
func (s *Memory) Set(key string, val interface{}, ttl time.Duration) {
	var exp uint32
	if ttl > 0 {
		exp = uint32(ttl.Seconds()) + atomic.LoadUint32(&timestamp)
	}
	i := item{exp, val}
	s.Lock()
	s.data[key] = i
	s.Unlock()
}

// Delete key by key
func (s *Memory) Delete(key string) {
	s.Lock()
	delete(s.data, key)
	s.Unlock()
}

// Reset all keys
func (s *Memory) Reset() {
	nd := make(map[string]item)
	s.Lock()
	s.data = nd
	s.Unlock()
}

func (s *Memory) gc(sleep time.Duration) {
	ticker := time.NewTicker(sleep)
	defer ticker.Stop()
	var expired []string

	for range ticker.C {
		ts := atomic.LoadUint32(&timestamp)
		expired = expired[:0]
		s.RLock()
		for key, v := range s.data {
			if v.e != 0 && v.e <= ts {
				expired = append(expired, key)
			}
		}
		s.RUnlock()
		s.Lock()
		// Double-checked locking.
		// We might have replaced the item in the meantime.
		for i := range expired {
			v := s.data[expired[i]]
			if v.e != 0 && v.e <= ts {
				delete(s.data, expired[i])
			}
		}
		s.Unlock()
	}
}

// StartTimeStampUpdater starts a concurrent function which stores the timestamp to an atomic value per second,
// which is much better for performance than determining it at runtime each time
func startTimeStampUpdater() {
	timestampTimer.Do(func() {
		// set initial value
		atomic.StoreUint32(&timestamp, uint32(time.Now().Unix()))
		go func(sleep time.Duration) {
			ticker := time.NewTicker(sleep)
			defer ticker.Stop()

			for t := range ticker.C {
				// update timestamp
				atomic.StoreUint32(&timestamp, uint32(t.Unix()))
			}
		}(1 * time.Second) // duration
	})
}
