package omultimap

import (
	"errors"
	"encoding/json"
	"sync"

	"github.com/matheusoliveira/go-ordered-map/pkg/omap"
)

type OMultiMapSync[K comparable, V any] struct {
	omm  OMultiMap[K, V]
	lock sync.RWMutex
}

type OMultiMapSyncIterator[K comparable, V any] struct {
	m  *OMultiMapSync[K, V]
	it omap.OMapIterator[K, V]
}

type OMultiMapSyncValuesIterator[K comparable, V any] struct {
	elems []*mapEntry[K, V]
	pos   int
}

func NewOMultiMapSync[K comparable, V any]() OMultiMap[K, V] {
	ret := &OMultiMapSync[K, V]{
		omm: NewOMultiMapLinked[K, V](),
	}
	return ret
}

func (m *OMultiMapSync[K, V]) Put(key K, value V) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.omm.Put(key, value)
}

func (m *OMultiMapSync[K, V]) GetValuesOf(key K) omap.OMapIterator[K, V] {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.omm.GetValuesOf(key)
}

func (m *OMultiMapSync[K, V]) DeleteAll(key K) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.omm.DeleteAll(key)
}

func (m *OMultiMapSync[K, V]) DeleteAt(interfaceIt omap.OMapIterator[K, V]) error {
	if it, ok := interfaceIt.(*OMultiMapSyncIterator[K, V]); !ok {
		return errors.New("trying to operate on invalid map iterator")
	} else {
		m.lock.Lock()
		defer m.lock.Unlock()
		return m.omm.DeleteAt(it.it)
	}
}

func (m *OMultiMapSync[K, V]) MustDeleteAt(interfaceIt omap.OMapIterator[K, V]) {
	err := m.DeleteAt(interfaceIt)
	if err != nil {
		panic(err)
	}
}

func (m *OMultiMapSync[K, V]) Iterator() omap.OMapIterator[K, V] {
	it := m.omm.Iterator()
	return &OMultiMapSyncIterator[K, V]{it: it, m: m}
}

// Implement fmt.Stringer
func (m *OMultiMapSync[K, V]) String() string {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return omap.IteratorToString[K, V]("omultimap.OMultiMapSync", m.omm.Iterator())
}

// Implement json.Marshaler interface.
func (m *OMultiMapSync[K, V]) MarshalJSON() ([]byte, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	buffer, err := omap.MarshalJSON(m.omm.Iterator())
	return buffer, err
}

// Implement json.Unmarshaler interface.
func (m *OMultiMapSync[K, V]) UnmarshalJSON(b []byte) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	err := json.Unmarshal(b, &m.omm)
	return err
}

//// OMultiMap Iterator ////

func (it *OMultiMapSyncIterator[K, V]) Next() bool {
	it.m.lock.RLock()
	defer it.m.lock.RUnlock()
	return it.it.Next()
}

func (it *OMultiMapSyncIterator[K, V]) EOF() bool {
	return it.it.EOF()
}

func (it *OMultiMapSyncIterator[K, V]) Key() K {
	return it.it.Key()
}

func (it *OMultiMapSyncIterator[K, V]) Value() V {
	return it.it.Value()
}
