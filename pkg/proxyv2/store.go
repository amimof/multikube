package proxy

import (
	"sync/atomic"
)

type RuntimeStore struct {
	current atomic.Pointer[Runtime]
}

func NewRuntimeStore() *RuntimeStore {
	s := &RuntimeStore{}
	s.current.Store(&Runtime{
		Version:   0,
		Listeners: map[string]*ListenerRuntime{},
	})
	return s
}

func (s *RuntimeStore) Load() *Runtime {
	return s.current.Load()
}

func (s *RuntimeStore) Store(rt *Runtime) {
	s.current.Store(rt)
}
