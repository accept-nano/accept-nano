package maplock

import "sync"

type MapLock struct {
	locks map[string]*sync.Mutex
	m     sync.Mutex
}

func New() *MapLock {
	return &MapLock{
		locks: make(map[string]*sync.Mutex),
	}
}

func (m *MapLock) Lock(key string) {
	m.m.Lock()
	l, ok := m.locks[key]
	if !ok {
		l = new(sync.Mutex)
		m.locks[key] = l
	}
	m.m.Unlock()
	l.Lock()
}

func (m *MapLock) Unlock(key string) {
	m.m.Lock()
	l := m.locks[key]
	m.m.Unlock()
	l.Unlock()
}
