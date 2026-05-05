package store

type EvictionPolicy string

const (
	NoEviction     EvictionPolicy = "noeviction"
	AllKeysRandom  EvictionPolicy = "allkeys-random"
	VolatileRandom EvictionPolicy = "volatile-random"
	AllKeysLRU     EvictionPolicy = "allkeys-lru"
	VolatileLRU    EvictionPolicy = "volatile-lru"
)
