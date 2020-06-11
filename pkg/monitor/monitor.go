package monitor

import (
	"expvar"
	"sync"
)

type Var = expvar.Var

type KeyValue = expvar.KeyValue

type Monitor interface {
	With(namespace string) Monitor
	Add(key string, delta int64)
	AddFloat(key string, delta float64)
	Delete(key string)
	Do(f func(namespace string, kv KeyValue))
	Get(key string) Var
	Set(key string, av Var)
	String() string
}

type monitor struct {
	namespace string
	lock      *sync.RWMutex
	m         map[string]*expvar.Map
}

func NewMonitor(namespace string) Monitor {
	return &monitor{
		namespace: namespace,
		lock:      &sync.RWMutex{},
		m: map[string]*expvar.Map{
			namespace: expvar.NewMap(namespace),
		},
	}
}

func (m *monitor) With(namespace string) Monitor {
	m.lock.Lock()
	if _, ok := m.m[namespace]; !ok {
		m.m[namespace] = expvar.NewMap(namespace)
	}
	m.lock.Unlock()

	return &monitor{
		namespace: namespace,
		lock:      m.lock,
		m:         m.m,
	}
}

func (m *monitor) Add(key string, delta int64) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.m[m.namespace].Add(key, delta)
}

func (m *monitor) AddFloat(key string, delta float64) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.m[m.namespace].AddFloat(key, delta)
}

func (m *monitor) Delete(key string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.m[m.namespace].Delete(key)
}

func (m *monitor) Do(f func(namespace string, kv KeyValue)) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	for namespace, m := range m.m {
		m.Do(func(kv KeyValue) {
			f(namespace, kv)
		})
	}
}

func (m *monitor) Get(key string) Var {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.m[m.namespace].Get(key)
}

func (m *monitor) Set(key string, av Var) {
	m.lock.Lock()
	moni := m.m[m.namespace]
	moni.Set(key, av)
	m.lock.Unlock()
}

func (m *monitor) String() string {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.m[m.namespace].String()
}
