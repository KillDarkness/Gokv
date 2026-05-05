package store

type Entry struct {
	Type       ValueType
	Value      any
	ExpiresAt  int64
	LastAccess int64
}
