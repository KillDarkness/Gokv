package store

type ValueType int

const (
	TypeString ValueType = iota
	TypeList
	TypeHash
	TypeSet
)
