package backend

import (
	"sync"
	"time"
)

type LockStore struct {
	mutex sync.Mutex
	locks map[string]lock
}

func NewLockStore() *LockStore {
	return &LockStore{locks: map[string]lock{}}
}

func (s *LockStore) Get(id string) string {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	held, found := s.locks[id]
	if !found || time.Now().After(held.expiry) {
		delete(s.locks, id)
		return ""
	}
	return held.value
}

func (s *LockStore) Set(id, value string) {
	s.mutex.Lock()
	s.locks[id] = lock{value: value, expiry: time.Now().Add(lockTTL)}
	s.mutex.Unlock()
}

func (s *LockStore) Clear(id string) {
	s.mutex.Lock()
	delete(s.locks, id)
	s.mutex.Unlock()
}
