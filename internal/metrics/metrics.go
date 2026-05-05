package metrics

import (
	"sync/atomic"
	"time"
)

type Metrics struct {
	startedAt time.Time
	commands  atomic.Uint64
}

func New() *Metrics {
	return &Metrics{startedAt: time.Now()}
}

func (m *Metrics) IncCommands() {
	m.commands.Add(1)
}

func (m *Metrics) Commands() uint64 {
	return m.commands.Load()
}

func (m *Metrics) Uptime() time.Duration {
	return time.Since(m.startedAt)
}
