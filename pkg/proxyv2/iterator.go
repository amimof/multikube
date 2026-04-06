package proxy

import (
	"sort"
	"sync/atomic"
)

type BackendIterator interface {
	Next([]*BackendRuntime) (*BackendRuntime, bool)
}

type RoundRobinLB struct {
	rr atomic.Int32
}

func (r *RoundRobinLB) Next(targets []*BackendRuntime) (*BackendRuntime, bool) {
	n := uint64(r.rr.Add(1))
	idx := int(n % uint64(len(targets)))
	return targets[idx], true
}

type LeastConnectionsLB struct{}

func (r *LeastConnectionsLB) Next(targets []*BackendRuntime) (*BackendRuntime, bool) {
	sort.SliceStable(targets, func(i, j int) bool {
		return targets[i].Active.Load() < targets[j].Active.Load()
	})
	return targets[0], true
}
