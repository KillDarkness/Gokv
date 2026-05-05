package store

import (
	"strconv"
	"sync"
	"time"
)

type Store struct {
	mu             sync.RWMutex
	data           map[string]Entry
	keyCount       int
	maxKeys        int
	evictionPolicy EvictionPolicy
}

func New() *Store {
	return NewWithOptions(Options{EvictionPolicy: NoEviction})
}

func NewWithOptions(opts Options) *Store {
	policy := opts.EvictionPolicy
	if policy == "" {
		policy = NoEviction
	}
	if !policy.Valid() {
		policy = NoEviction
	}
	return &Store{data: make(map[string]Entry), maxKeys: opts.MaxKeys, evictionPolicy: policy}
}

func (s *Store) Set(key string, value string) error {
	now := time.Now().UnixNano()

	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.evictForWriteLocked([]string{key}, now); err != nil {
		return err
	}

	s.setEntryLocked(key, Entry{Type: TypeString, Value: value, LastAccess: now})
	return nil
}

func (s *Store) MSet(values map[string]string) error {
	now := time.Now().UnixNano()

	s.mu.Lock()
	defer s.mu.Unlock()

	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	if err := s.evictForWriteLocked(keys, now); err != nil {
		return err
	}

	for key, value := range values {
		s.setEntryLocked(key, Entry{Type: TypeString, Value: value, LastAccess: now})
	}
	return nil
}

func (s *Store) Get(key string) (string, bool) {
	now := time.Now().UnixNano()
	if !s.evictionPolicy.TracksAccess() {
		s.mu.RLock()
		entry, ok := s.data[key]
		if !ok {
			s.mu.RUnlock()
			return "", false
		}
		if entryExpired(entry, now) {
			s.mu.RUnlock()
			s.deleteExpiredKeys([]string{key}, now)
			return "", false
		}
		if entry.Type != TypeString {
			s.mu.RUnlock()
			return "", false
		}
		value, ok := entry.Value.(string)
		s.mu.RUnlock()
		return value, ok
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.data[key]
	if !ok {
		return "", false
	}
	if entryExpired(entry, now) {
		s.deleteEntryLocked(key)
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
	s.setEntryLocked(key, entry)

	return value, true
}

func (s *Store) MGet(keys ...string) []StringResult {
	now := time.Now().UnixNano()
	if !s.evictionPolicy.TracksAccess() {
		s.mu.RLock()
		results := make([]StringResult, 0, len(keys))
		expired := make([]string, 0)
		for _, key := range keys {
			entry, ok := s.data[key]
			if !ok {
				results = append(results, StringResult{})
				continue
			}
			if entryExpired(entry, now) {
				expired = append(expired, key)
				results = append(results, StringResult{})
				continue
			}
			value, ok := entry.Value.(string)
			if entry.Type != TypeString || !ok {
				results = append(results, StringResult{})
				continue
			}
			results = append(results, StringResult{Value: value, OK: true})
		}
		s.mu.RUnlock()
		if len(expired) > 0 {
			s.deleteExpiredKeys(expired, now)
		}
		return results
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	results := make([]StringResult, 0, len(keys))
	for _, key := range keys {
		entry, ok := s.data[key]
		if !ok {
			results = append(results, StringResult{})
			continue
		}
		if entryExpired(entry, now) {
			s.deleteEntryLocked(key)
			results = append(results, StringResult{})
			continue
		}
		value, ok := entry.Value.(string)
		if entry.Type != TypeString || !ok {
			results = append(results, StringResult{})
			continue
		}
		entry.LastAccess = now
		s.setEntryLocked(key, entry)
		results = append(results, StringResult{Value: value, OK: true})
	}

	return results
}

func (s *Store) Delete(keys ...string) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	var deleted int64
	for _, key := range keys {
		if _, ok := s.data[key]; ok {
			s.deleteEntryLocked(key)
			deleted++
		}
	}
	return deleted
}

func (s *Store) FlushDB() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data = make(map[string]Entry)
	s.keyCount = 0
}

func (s *Store) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.keyCount
}

func (s *Store) Snapshot() map[string]SnapshotEntry {
	now := time.Now().UnixNano()

	s.mu.RLock()
	defer s.mu.RUnlock()

	snapshot := make(map[string]SnapshotEntry, len(s.data))
	for key, entry := range s.data {
		if entryExpired(entry, now) {
			continue
		}
		if entry.Type != TypeString {
			continue
		}
		value, ok := entry.Value.(string)
		if !ok {
			continue
		}
		snapshot[key] = SnapshotEntry{Type: entry.Type, Value: value, ExpiresAt: entry.ExpiresAt, LastAccess: entry.LastAccess}
	}
	return snapshot
}

func (s *Store) Restore(snapshot map[string]SnapshotEntry) {
	now := time.Now().UnixNano()

	s.mu.Lock()
	defer s.mu.Unlock()

	s.data = make(map[string]Entry, len(snapshot))
	s.keyCount = 0
	for key, entry := range snapshot {
		storeEntry := Entry{Type: entry.Type, Value: entry.Value, ExpiresAt: entry.ExpiresAt, LastAccess: entry.LastAccess}
		if entryExpired(storeEntry, now) {
			continue
		}
		s.setEntryLocked(key, storeEntry)
	}
}

func (s *Store) Exists(keys ...string) int64 {
	now := time.Now().UnixNano()

	s.mu.RLock()
	expired := make([]string, 0)

	var count int64
	for _, key := range keys {
		entry, ok := s.data[key]
		if !ok {
			continue
		}
		if entryExpired(entry, now) {
			expired = append(expired, key)
			continue
		}
		count++
	}
	s.mu.RUnlock()

	if len(expired) > 0 {
		s.deleteExpiredKeys(expired, now)
	}
	return count
}

func (s *Store) deleteExpiredKeys(keys []string, now int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, key := range keys {
		entry, ok := s.data[key]
		if ok && entryExpired(entry, now) {
			s.deleteEntryLocked(key)
		}
	}
}

func (s *Store) Increment(key string, delta int64) (int64, error) {
	now := time.Now().UnixNano()

	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.data[key]
	if !ok || entryExpired(entry, now) {
		if err := s.evictForWriteLocked([]string{key}, now); err != nil {
			return 0, err
		}
		value := delta
		if ok {
			s.deleteEntryLocked(key)
		}
		s.setEntryLocked(key, Entry{Type: TypeString, Value: strconv.FormatInt(value, 10), LastAccess: now})
		return value, nil
	}
	if entry.Type != TypeString {
		return 0, strconv.ErrSyntax
	}

	current, ok := entry.Value.(string)
	if !ok {
		return 0, strconv.ErrSyntax
	}
	value, err := strconv.ParseInt(current, 10, 64)
	if err != nil {
		return 0, err
	}
	value += delta
	entry.Value = strconv.FormatInt(value, 10)
	entry.LastAccess = now
	s.setEntryLocked(key, entry)

	return value, nil
}

func (s *Store) setEntryLocked(key string, entry Entry) {
	if _, ok := s.data[key]; !ok {
		s.keyCount++
	}
	s.data[key] = entry
}

func (s *Store) deleteEntryLocked(key string) {
	if _, ok := s.data[key]; !ok {
		return
	}
	delete(s.data, key)
	if s.keyCount > 0 {
		s.keyCount--
	}
}
