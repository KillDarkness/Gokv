package store

type Entry struct {
	Type       ValueType
	Value      any
	ExpiresAt  int64
	LastAccess int64
}

type SnapshotEntry struct {
	Type       ValueType `json:"type"`
	Value      string    `json:"value"`
	ExpiresAt  int64     `json:"expires_at"`
	LastAccess int64     `json:"last_access"`
}
