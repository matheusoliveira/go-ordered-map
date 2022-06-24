package omap

import (
	"fmt"
)

//// OMapSimple ////

// Implements a OMap interface using a very simple algorithm: it basically keeps a
// map[K]V to hold the mappings, and a []K slice to keep the order (hence doubling
// the memory used to store the keys, compared to a simple Go map).
type OMapSimple[K comparable, V any] struct {
	m    map[K]V
	keys []K
}

// Iterator over a OMapSimple, should be created through OMapSimple.Iterator() function.
type OMapSimpleIterator[K comparable, V any] struct {
	i int
	m *OMapSimple[K, V]
}

// Create a new OMap instance using OMapSimple implementation.
func NewOMapSimple[K comparable, V any]() OMap[K, V] {
	var m OMapSimple[K, V]
	m.init()
	return &m
}

func (m *OMapSimple[K, V]) init() {
	m.m = make(map[K]V)
}

// Add/overwrite the value in the map on the given key.
// Important to note that if a key existed and is being overwritten, the order of the old key
// insertion position will remain when iterating the map.
// Complexity: O(1)
func (m *OMapSimple[K, V]) Put(key K, value V) {
	if _, ok := m.m[key]; !ok {
		m.keys = append(m.keys, key)
	}
	m.m[key] = value
}

func (m *OMapSimple[K, V]) PutAfter(interfaceIt OMapIterator[K, V], key K, value V) error {
	if it, ok := interfaceIt.(*OMapSimpleIterator[K, V]); !ok {
		return fmt.Errorf("%w - expected OMapSimple found %T", ErrInvalidIteratorType, interfaceIt)
	} else if it.m != m {
		return ErrInvalidIteratorMap
	} else if !it.IsValid() && it.i != -1 {
		return ErrInvalidIteratorPos
	} else {
		if it.i >= 0 && it.Key() == key {
			// simple case, just overwrite
			m.Put(key, value)
			return nil
		}
		pos := it.i + 1
		tmp := make([]K, 0, len(m.keys)+1)
		tmp = append(tmp, m.keys[0:pos]...)
		tmp = append(tmp, key)
		tmp = append(tmp, m.keys[pos:]...)
		if _, ok := m.m[key]; ok {
			for i, k := range tmp {
				if k == key && i != pos {
					// move all values to the left
					copy(tmp, tmp[0:i])
					copy(tmp[i:], tmp[i+1:])
					tmp = tmp[0 : len(tmp)-1]
					break
				}
			}
		}
		m.keys = tmp
		m.m[key] = value
		return nil
	}
}

// Get the value pointing to the given key, returning true as second argument if found, and
// false otherwise.
// Complexity: O(1), same as builtin map[key]
func (m *OMapSimple[K, V]) Get(key K) (V, bool) {
	v, ok := m.m[key]
	return v, ok
}

func (m *OMapSimple[K, V]) GetIteratorAt(key K) OMapIterator[K, V] {
	if _, ok := m.m[key]; ok {
		for i, k := range m.keys {
			if k == key {
				return &OMapSimpleIterator[K, V]{i: i, m: m}
			}
		}
	}
	return &OMapSimpleIterator[K, V]{i: len(m.keys), m: m}
}

// Delete the value pointing to the given key.
// Complexity: O(n)
func (m *OMapSimple[K, V]) Delete(key K) {
	if _, ok := m.m[key]; !ok {
		return
	}
	//delete(m.m, key)
	//*
	pos := -1
	for i := 0; i < len(m.keys); i++ {
		if m.keys[i] == key {
			pos = i
			break
		}
	}
	if pos >= 0 {
		m.keys = append(m.keys[0:pos], m.keys[pos+1:]...)
		delete(m.m, key)
	}
	//*/
}

// Return an iterator to navigate the map.
func (m *OMapSimple[K, V]) Iterator() OMapIterator[K, V] {
	return &OMapSimpleIterator[K, V]{i: -1, m: m}
}

func (m *OMapSimple[K, V]) Len() int {
	return len(m.m)
}

// Implement fmt.Stringer
func (m *OMapSimple[K, V]) String() string {
	return IteratorToString[K, V]("omap.OMapSimple", m.Iterator())
}

// Implement json.Marshaler interface.
func (m OMapSimple[K, V]) MarshalJSON() ([]byte, error) {
	buffer, err := MarshalJSON(m.Iterator())
	return buffer, err
}

// Implement json.Unmarshaler interface.
func (m *OMapSimple[K, V]) UnmarshalJSON(b []byte) error {
	m.init()
	return UnmarshalJSON[K, V](m.Put, b)
}

// Move iterator to the next record, returning true if there is a next value and false otherwise.
// Complexity: in general should be O(1), but it needs to skip deleted keys, so if there M deleted
// keys on the current position, it will be O(M). It is a trade-off to avoid making Delete O(N).
func (it *OMapSimpleIterator[K, V]) Next() bool {
	for it.i++; it.i < len(it.m.keys); it.i++ {
		// ignore deleted keys
		if _, ok := it.m.m[it.Key()]; ok {
			break
		}
	}
	return it.i < len(it.m.keys)
}

// Returns true if iterator has reached the end
func (it OMapSimpleIterator[K, V]) EOF() bool {
	return it.i >= len(it.m.keys)
}

// Return the key at current record.
// Calling this function when EOF() is true will cause a panic.
func (it OMapSimpleIterator[K, V]) Key() K {
	return it.m.keys[it.i]
}

// Return the value at current record.
// Calling this function when EOF() is true will cause a panic.
func (it OMapSimpleIterator[K, V]) Value() V {
	key := it.Key()
	return it.m.m[key]
}

func (it OMapSimpleIterator[K, V]) IsValid() bool {
	return it.i >= 0 && it.i < len(it.m.keys)
}

func (it *OMapSimpleIterator[K, V]) MoveFront() OMapIterator[K, V] {
	it.i = -1
	return it
}

func (it *OMapSimpleIterator[K, V]) MoveBack() OMapIterator[K, V] {
	it.i = len(it.m.keys)
	return it
}

func (it *OMapSimpleIterator[K, V]) Prev() bool {
	it.i--
	return it.IsValid()
}
