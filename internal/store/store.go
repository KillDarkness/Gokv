package store

import (
	"sync"
	"time"
)

type Store struct {
	mu   sync.RWMutex
	data map[string]Entry
}

func New() *Store {
	return &Store{data: make(map[string]Entry)}
}

func (s *Store) Set(key string, value string) {
	now := time.Now().UnixNano()

	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[key] = Entry{Type: TypeString, Value: value, LastAccess: now}
}

func (s *Store) Get(key string) (string, bool) {
	now := time.Now().UnixNano()

	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.data[key]
	if !ok {
		return "", false
	}
	if entryExpired(entry, now) {
		delete(s.data, key)
		return "", false
	}
	if entry.Type != TypeString {
		return "", false
	}

	value, ok := entry.Value.(string)
	if !ok {
		return "", false
	}
	entry.LastAccess = now
	s.data[key] = entry

	return value, true
}

func (s *Store) Delete(keys ...string) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	var deleted int64
	for _, key := range keys {
		if _, ok := s.data[key]; ok {
			delete(s.data, key)
			deleted++
		}
	}
	return deleted
}

func (s *Store) Exists(keys ...string) int64 {
	now := time.Now().UnixNano()

	s.mu.Lock()
	defer s.mu.Unlock()

	var count int64
	for _, key := range keys {
		entry, ok := s.data[key]
		if !ok {
			continue
		}
		if entryExpired(entry, now) {
			delete(s.data, key)
			continue
		}
		count++
	}
	return count
}
