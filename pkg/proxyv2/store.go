package proxy

import (
	"sync/atomic"
)

type RuntimeStore struct {
	current atomic.Pointer[RuntimeConfig]
}

func NewRuntimeStore() *RuntimeStore {
	s := &RuntimeStore{}
	s.current.Store(&RuntimeConfig{
		Version:  0,
		Routes:   CompiledRoutes{},
		Backends: make(map[string]*BackendPool),
	})
	return s
}

func (s *RuntimeStore) Load() *RuntimeConfig {
	return s.current.Load()
}

func (s *RuntimeStore) Store(rt *RuntimeConfig) {
	s.current.Store(rt)
}
