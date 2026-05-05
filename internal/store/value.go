package store

type ValueType int

const (
	TypeString ValueType = iota
	TypeList
	TypeHash
	TypeSet
)

type StringResult struct {
	Value string
	OK    bool
}
