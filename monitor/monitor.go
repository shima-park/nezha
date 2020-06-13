package monitor

import (
	"expvar"
	"sync"
	"time"
)

type Var = expvar.Var

type KeyValue = expvar.KeyValue

type Elapsed time.Duration

func (e Elapsed) String() string {
	d := time.Duration(e)
	d = d.Truncate(time.Second)
	return d.String()
}

type Time time.Time

func (t Time) String() string {
	return time.Time(t).Format("2006-01-02 15:04:05")
}

type String string

func (s String) String() string {
	return string(s)
}

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
	if v := m.m[m.namespace].Get(key); v != nil {
		return v
	}
	return String("")
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
