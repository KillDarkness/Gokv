package store

import (
	"errors"
	"math/rand"
)

var ErrMaxKeysReached = errors.New("max keys limit reached")

type EvictionPolicy string

const (
	NoEviction     EvictionPolicy = "noeviction"
	AllKeysRandom  EvictionPolicy = "allkeys-random"
	VolatileRandom EvictionPolicy = "volatile-random"
	AllKeysLRU     EvictionPolicy = "allkeys-lru"
	VolatileLRU    EvictionPolicy = "volatile-lru"
)

type Options struct {
	MaxKeys        int
	EvictionPolicy EvictionPolicy
}

func (p EvictionPolicy) Valid() bool {
	switch p {
	case NoEviction, AllKeysRandom, VolatileRandom, AllKeysLRU, VolatileLRU:
		return true
	default:
		return false
	}
}

func (s *Store) evictForWriteLocked(keys []string, now int64) error {
	if s.maxKeys <= 0 {
		return nil
	}

	needed := 0
	seen := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		if _, duplicate := seen[key]; duplicate {
			continue
		}
		seen[key] = struct{}{}

		entry, ok := s.data[key]
		if ok && entryExpired(entry, now) {
			delete(s.data, key)
			ok = false
		}
		if !ok {
			needed++
		}
	}
	if needed == 0 || len(s.data)+needed <= s.maxKeys {
		return nil
	}
	if s.evictionPolicy == NoEviction {
		return ErrMaxKeysReached
	}

	for len(s.data)+needed > s.maxKeys {
		victim, ok := s.evictionCandidateLocked()
		if !ok {
			return ErrMaxKeysReached
		}
		delete(s.data, victim)
	}
	return nil
}

func (s *Store) evictionCandidateLocked() (string, bool) {
	switch s.evictionPolicy {
	case AllKeysRandom:
		return randomKey(s.data, false)
	case VolatileRandom:
		return randomKey(s.data, true)
	case AllKeysLRU:
		return lruKey(s.data, false)
	case VolatileLRU:
		return lruKey(s.data, true)
	default:
		return "", false
	}
}

func randomKey(data map[string]Entry, volatileOnly bool) (string, bool) {
	keys := make([]string, 0, len(data))
	for key, entry := range data {
		if volatileOnly && entry.ExpiresAt == 0 {
			continue
		}
		keys = append(keys, key)
	}
	if len(keys) == 0 {
		return "", false
	}
	return keys[rand.Intn(len(keys))], true
}

func lruKey(data map[string]Entry, volatileOnly bool) (string, bool) {
	var victim string
	var oldest int64
	found := false
	for key, entry := range data {
		if volatileOnly && entry.ExpiresAt == 0 {
			continue
		}
		if !found || entry.LastAccess < oldest {
			victim = key
			oldest = entry.LastAccess
			found = true
		}
	}
	return victim, found
}
