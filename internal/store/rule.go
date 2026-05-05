package store

import (
	"sort"
	"time"
)

type Rule struct {
	Prefix string
	TTL    time.Duration
}

func (s *Store) SetRule(prefix string, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.rules[prefix] = Rule{Prefix: prefix, TTL: ttl}
}

func (s *Store) DeleteRule(prefix string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.rules[prefix]; !ok {
		return false
	}
	delete(s.rules, prefix)
	return true
}

func (s *Store) Rules() []Rule {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rules := make([]Rule, 0, len(s.rules))
	for _, rule := range s.rules {
		rules = append(rules, rule)
	}
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Prefix < rules[j].Prefix
	})
	return rules
}
