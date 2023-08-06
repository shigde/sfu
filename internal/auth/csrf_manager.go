package auth

import (
	"time"
	"unsafe"

	"github.com/shigde/sfu/internal/storage"
)

type item struct{}

type manager struct {
	storage *storage.Memory
}

func newManager() *manager {
	// Create new storage handler
	manager := &manager{
		storage: storage.NewMemory(),
	}
	return manager
}

// get raw data from storage or memory
func (m *manager) getToken(key string) string {
	var token string
	token, _ = m.storage.Get(key).(string) //nolint:errcheck // TODO: Do not ignore error
	return token
}

// set data to storage or memory
func (m *manager) setToken(key string, raw string, exp time.Duration) {
	// the key is crucial in crsf and sometimes a reference to another value which can be reused later(pool/unsafe values concept), so a copy is made here
	m.storage.Set(copyString(key), raw, exp)
}

func (m *manager) delete(key string) {
	// the key is crucial in crsf and sometimes a reference to another value which can be reused later(pool/unsafe values concept), so a copy is made here
	m.storage.Delete(key)
}

// CopyString copies a string to make it immutable
func copyString(s string) string {
	return string(unsafeBytes(s))
}

// UnsafeBytes returns a byte pointer without allocation.
func unsafeBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}
