package storage

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func checkTimeStamp(tb testing.TB, expectedCurrent, actualCurrent uint32) { //nolint:thelper // TODO: Verify if tb can be nil
	if tb != nil {
		tb.Helper()
	}
	// test with some buffer in front and back of the expectedCurrent time -> because of the timing on the work machine
	assert.True(tb, actualCurrent >= expectedCurrent-1 || actualCurrent <= expectedCurrent+1)
}

func Test_TimeStampUpdater(t *testing.T) {
	t.Parallel()

	startTimeStampUpdater()

	now := uint32(time.Now().Unix())
	checkTimeStamp(t, now, atomic.LoadUint32(&timestamp))
	// one second later
	time.Sleep(1 * time.Second)
	checkTimeStamp(t, now+1, atomic.LoadUint32(&timestamp))
	// two seconds later
	time.Sleep(1 * time.Second)
	checkTimeStamp(t, now+2, atomic.LoadUint32(&timestamp))
}

func Benchmark_CalculateTimestamp(b *testing.B) {
	startTimeStampUpdater()

	var res uint32
	b.Run("fiber", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			res = atomic.LoadUint32(&timestamp)
		}
		checkTimeStamp(b, uint32(time.Now().Unix()), res)
	})
	b.Run("default", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			res = uint32(time.Now().Unix())
		}
		checkTimeStamp(b, uint32(time.Now().Unix()), res)
	})
}

func Test_Memory(t *testing.T) {
	t.Parallel()
	var store = NewMemory()
	var (
		key             = "john"
		val interface{} = []byte("doe")
		exp             = 1 * time.Second
	)

	store.Set(key, val, 0)
	store.Set(key, val, 0)

	result := store.Get(key)
	assert.Equal(t, val, result)

	result = store.Get("empty")
	assert.Equal(t, nil, result)

	store.Set(key, val, exp)
	time.Sleep(1100 * time.Millisecond)

	result = store.Get(key)
	assert.Equal(t, nil, result)

	store.Set(key, val, 0)
	result = store.Get(key)
	assert.Equal(t, val, result)

	store.Delete(key)
	result = store.Get(key)
	assert.Equal(t, nil, result)

	store.Set("john", val, 0)
	store.Set("doe", val, 0)
	store.Reset()

	result = store.Get("john")
	assert.Equal(t, nil, result)

	result = store.Get("doe")
	assert.Equal(t, nil, result)
}

// go test -v -run=^$ -bench=Benchmark_Memory -benchmem -count=4
func Benchmark_Memory(b *testing.B) {
	keyLength := 1000
	keys := make([]string, keyLength)
	for i := 0; i < keyLength; i++ {
		keys[i] = uuid.New().String()
	}
	value := []byte("joe")

	ttl := 2 * time.Second
	b.Run("fiber_memory", func(b *testing.B) {
		d := NewMemory()
		b.ReportAllocs()
		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			for _, key := range keys {
				d.Set(key, value, ttl)
			}
			for _, key := range keys {
				_ = d.Get(key)
			}
			for _, key := range keys {
				d.Delete(key)

			}
		}
	})
}
