package metrics

import (
	"sync/atomic"
	"time"
)

type Metrics struct {
	startedAt         time.Time
	commands          atomic.Uint64
	errors            atomic.Uint64
	totalConnections  atomic.Uint64
	activeConnections atomic.Int64
}

func New() *Metrics {
	return &Metrics{startedAt: time.Now()}
}

func (m *Metrics) IncCommands() {
	m.commands.Add(1)
}

func (m *Metrics) IncErrors() {
	m.errors.Add(1)
}

func (m *Metrics) IncConnections() {
	m.totalConnections.Add(1)
	m.activeConnections.Add(1)
}

func (m *Metrics) DecConnections() {
	m.activeConnections.Add(-1)
}

func (m *Metrics) Commands() uint64 {
	return m.commands.Load()
}

func (m *Metrics) Errors() uint64 {
	return m.errors.Load()
}

func (m *Metrics) TotalConnections() uint64 {
	return m.totalConnections.Load()
}

func (m *Metrics) ActiveConnections() int64 {
	return m.activeConnections.Load()
}

func (m *Metrics) Uptime() time.Duration {
	return time.Since(m.startedAt)
}
