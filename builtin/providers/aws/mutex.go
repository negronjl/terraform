package aws

import (
	"log"
	"sync"
)

// MutexKV is a simple key/value store for arbitrary mutexes. It can be used to
// serialize changes across arbitrary collaborators that share knowledge of the
// keys they must serialize on.
//
// The initial use case is to let aws_security_group_rule resources serialize
// their access to individual security groups based on SG ID.
type MutexKV struct {
	sync.Mutex
	store map[string]*sync.Mutex
	once  sync.Once
}

func (m *MutexKV) init() {
	m.store = make(map[string]*sync.Mutex)
}

// Returns a mutex for the given key, no guarantee of its lock status
func (m *MutexKV) Get(key string) *sync.Mutex {
	m.once.Do(m.init)
	m.Lock()
	defer m.Unlock()
	mutex, ok := m.store[key]
	if !ok {
		mutex = &sync.Mutex{}
		m.store[key] = mutex
	}
	return mutex
}

// This is a global MutexKV for use within this plugin.
var mutexStore = &MutexKV{}

// Acquires a lock for the given key, returns a function to unlock it
func acquireLockFor(key string) func() {
	mutex := mutexStore.Get(key)
	log.Printf("[LOCK] Locking %q: %#v", key, mutexStore)
	mutex.Lock()
	return func() {
		log.Printf("[LOCK] Unlocking %q: %#v", key, mutexStore)
		mutex.Unlock()
	}
}
