package store

import (
	"context"
	"time"
)

func (s *Store) Expire(key string, ttl time.Duration) bool {
	now := time.Now().UnixNano()

	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.data[key]
	if !ok || entryExpired(entry, now) {
		if ok {
			s.deleteEntryLocked(key)
		}
		return false
	}
	if ttl <= 0 {
		s.deleteEntryLocked(key)
		return true
	}

	entry.ExpiresAt = now + ttl.Nanoseconds()
	s.setEntryLocked(key, entry)
	return true
}

func (s *Store) TTL(key string) (time.Duration, bool, bool) {
	now := time.Now().UnixNano()

	s.mu.RLock()

	entry, ok := s.data[key]
	if !ok {
		s.mu.RUnlock()
		return 0, false, false
	}
	if entryExpired(entry, now) {
		s.mu.RUnlock()
		s.deleteExpiredKeys([]string{key}, now)
		return 0, false, false
	}
	if entry.ExpiresAt == 0 {
		s.mu.RUnlock()
		return 0, true, false
	}
	ttl := time.Duration(entry.ExpiresAt - now)
	s.mu.RUnlock()
	return ttl, true, true
}

func (s *Store) StartJanitor(ctx context.Context, interval time.Duration) <-chan struct{} {
	done := make(chan struct{})

	go func() {
		defer close(done)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.DeleteExpired()
			}
		}
	}()

	return done
}

func (s *Store) DeleteExpired() int64 {
	now := time.Now().UnixNano()

	s.mu.Lock()
	defer s.mu.Unlock()

	var deleted int64
	for key, entry := range s.data {
		if entryExpired(entry, now) {
			s.deleteEntryLocked(key)
			deleted++
		}
	}
	return deleted
}

func entryExpired(entry Entry, now int64) bool {
	return entry.ExpiresAt > 0 && entry.ExpiresAt <= now
}
