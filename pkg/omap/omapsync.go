package omap

import (
	"encoding/json"
	"sync"
)

//// OMapSync ////

// Implements a OMap interface using a very simple algorithm: it basically keeps a
// map[K]V to hold the mappings, and a []K slice to keep the order (hence doubling
// the memory used to store the keys, compared to a simple Go map).
type OMapSync[K comparable, V any] struct {
	om OMap[K, V]
	mx sync.RWMutex
}

// Iterator over a OMapSync, should be created through OMapSync.Iterator() function.
type OMapSyncIterator[K comparable, V any] struct {
	it OMapIterator[K, V]
	m  *OMapSync[K, V]
}

// Create a new OMap instance using OMapSync implementation.
func NewOMapSync[K comparable, V any]() OMap[K, V] {
	return &OMapSync[K, V]{
		om: NewOMapLinked[K, V](),
	}
}

func (m *OMapSync[K, V]) init() {
	if m.om == nil {
		func() {
			m.mx.Lock()
			defer m.mx.Unlock()
			if m.om == nil {
				m.om = New[K, V]()
			}
		}()
	}
}

// Add/overwrite the value in the map on the given key.
// Important to note that if a key existed and is being overwritten, the order of the old key
// insertion position will remain when iterating the map.
// Complexity: O(1)
func (m *OMapSync[K, V]) Put(key K, value V) {
	m.mx.Lock()
	defer m.mx.Unlock()
	m.om.Put(key, value)
}

// Get the value pointing to the given key, returning true as second argument if found, and
// false otherwise.
// Complexity: O(1), same as builtin map[key]
func (m *OMapSync[K, V]) Get(key K) (V, bool) {
	m.mx.RLock()
	defer m.mx.RUnlock()
	v, ok := m.om.Get(key)
	return v, ok
}

// Delete the value pointing to the given key.
// Complexity: same as builtin [delete](https://pkg.go.dev/builtin#delete)
func (m *OMapSync[K, V]) Delete(key K) {
	m.mx.Lock()
	defer m.mx.Unlock()
	m.om.Delete(key)
}

// Return an iterator to navigate the map.
func (m *OMapSync[K, V]) Iterator() OMapIterator[K, V] {
	m.mx.RLock()
	defer m.mx.RUnlock()
	return &OMapSyncIterator[K, V]{it: m.om.Iterator(), m: m}
}

func (m *OMapSync[K, V]) Len() int {
	return m.om.Len()
}

// Implement fmt.Stringer interface.
func (m *OMapSync[K, V]) String() string {
	return toString[K, V]("omap.OMapSync", m.Iterator())
}

// Implement json.Marshaler interface.
func (m *OMapSync[K, V]) MarshalJSON() ([]byte, error) {
	m.mx.RLock()
	defer m.mx.RUnlock()
	buffer, err := json.Marshal(m.om)
	return buffer, err
}

// Implement json.Unmarshaler interface.
func (m *OMapSync[K, V]) UnmarshalJSON(b []byte) error {
	m.init()
	m.mx.RLock()
	defer m.mx.RUnlock()
	err := json.Unmarshal(b, &m.om)
	return err
}

// Move iterator to the next record, returning true if there is a next value and false otherwise.
// Complexity: in general should be O(1), but it needs to skip deleted keys, so if there M deleted
// keys on the current position, it will be O(M). It is a trade-off to avoid making Delete O(N).
func (it *OMapSyncIterator[K, V]) Next() bool {
	it.m.mx.RLock()
	defer it.m.mx.RUnlock()
	return it.it.Next()
}

// Returns true if iterator has reached the end
func (it OMapSyncIterator[K, V]) EOF() bool {
	return it.it.EOF()
}

// Return the key at current record.
// Calling this function when EOF() is true will cause a panic.
func (it OMapSyncIterator[K, V]) Key() K {
	return it.it.Key()
}

// Return the value at current record.
// Calling this function when EOF() is true will cause a panic.
func (it OMapSyncIterator[K, V]) Value() V {
	return it.it.Value()
}
