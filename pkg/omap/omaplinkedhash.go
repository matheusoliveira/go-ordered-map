package omap

import (
	"fmt"
	"hash/maphash"
)

//// OMapLinkedHash ////

// Objects that want to create a custom hashing for the key used by OMapLinkedHash must implement
// this interface, giving a HashSum32 func that returns the hash of the object as an uint32.
type Hasher interface {
	HashSum32() uint32
}

// Hash func to be used by a given key
type hasherFunc[K comparable] func(key *K) uint32

// Implement an ordered map using a linked list but saving the key as an uint32 hash instead of
// copying the key into the map. This implementation should only be used when you have a very
// large object struct as K key, and preferable this object should implement Hasher
// interface to provide a performant hashing algorithm for the type.
type OMapLinkedHash[K comparable, V any] struct {
	//hasher maphash.Hash
	m      map[uint32][]*mapEntry[*K, V]
	head   *mapEntry[*K, V]
	tail   *mapEntry[*K, V]
	hasher hasherFunc[K]
}

// Implement OMapIterator for OMapLinkedHash
type OMapLinkedHashIterator[K comparable, V any] struct {
	cursor *mapEntry[*K, V]
	bof    bool
}

// Return a new OMap based on OMapLinkedHash implementation, see OMapLinkedHash type for more
// details of the implementation.
func NewOMapLinkedHash[K comparable, V any]() OMap[K, V] {
	var m OMapLinkedHash[K, V]
	m.init()
	return &m
}

func (m *OMapLinkedHash[K, V]) init() {
	m.m = make(map[uint32][]*mapEntry[*K, V])
	m.head = nil
	m.tail = nil
	m.setupHasher()
}

func (m *OMapLinkedHash[K, V]) setupHasher() {
	var key K
	switch any(key).(type) {
	case Hasher:
		m.hasher = func(key *K) uint32 {
			h, _ := any(*key).(Hasher)
			return h.HashSum32()
		}
	case string:
		var mh maphash.Hash
		m.hasher = func(key *K) uint32 {
			s, _ := any(*key).(string)
			mh.WriteString(s)
			ret := mh.Sum64()
			mh.Reset()
			return uint32(ret)
		}
	case uint32:
		m.hasher = func(key *K) uint32 {
			i, _ := any(*key).(uint32)
			return i
		}
	case int32:
		m.hasher = func(key *K) uint32 {
			i, _ := any(*key).(int32)
			return uint32(i)
		}
	case uint64:
		m.hasher = func(key *K) uint32 {
			i, _ := any(*key).(uint64)
			return uint32(i)
		}
	case int64:
		m.hasher = func(key *K) uint32 {
			i, _ := any(*key).(int64)
			return uint32(i)
		}
	case float32:
		m.hasher = func(key *K) uint32 {
			i, _ := any(*key).(float32)
			return uint32(i)
		}
	case float64:
		m.hasher = func(key *K) uint32 {
			i, _ := any(*key).(float64)
			return uint32(i)
		}
	default:
		// always work, but it's slow
		var mh maphash.Hash
		m.hasher = func(key *K) uint32 {
			s := fmt.Sprint(*key)
			mh.WriteString(s)
			ret := mh.Sum64()
			mh.Reset()
			return uint32(ret)
		}
	}
}

func (m *OMapLinkedHash[K, V]) Put(key K, value V) {
	elems, pos, hashedKey := m.getEntry(&key)
	if pos >= 0 {
		elems[pos].value = value
	} else {
		entry := &mapEntry[*K, V]{
			key:   &key,
			value: value,
			next:  nil,
			prev:  m.tail,
		}
		if elems != nil {
			m.m[hashedKey] = append(elems, entry)
		} else {
			m.m[hashedKey] = []*mapEntry[*K, V]{entry}
		}
		if m.head == nil {
			m.head = entry
			m.tail = entry
		} else {
			m.tail.next = entry
			m.tail = entry
		}
	}
}

func (m *OMapLinkedHash[K, V]) getEntry(key *K) ([]*mapEntry[*K, V], int, uint32) {
	var elems []*mapEntry[*K, V]
	var ok bool
	hashedKey := m.hasher(key)
	if elems, ok = m.m[hashedKey]; ok {
		for i, entry := range elems {
			if *entry.key == *key {
				return elems, i, hashedKey
			}
		}
	}
	// not found
	return elems, -1, hashedKey
}

func (m *OMapLinkedHash[K, V]) Get(key K) (V, bool) {
	elems, pos, _ := m.getEntry(&key)
	if pos >= 0 {
		return elems[pos].value, true
	} else {
		var val V
		return val, false
	}
}

func (m *OMapLinkedHash[K, V]) Delete(key K) {
	elems, pos, hashedKey := m.getEntry(&key)
	if pos >= 0 {
		entry := elems[pos]
		if entry != nil {
			if m.head == entry {
				m.head = entry.next
			}
			if m.tail == entry {
				m.tail = entry.prev
			}
			if entry.prev != nil {
				entry.prev.next = entry.next
			}
			if entry.next != nil {
				entry.next.prev = entry.prev
			}
			m.m[hashedKey] = append(elems[0:pos], elems[pos+1:]...)
		}
	}
}

func (m *OMapLinkedHash[K, V]) Iterator() OMapIterator[K, V] {
	return &OMapLinkedHashIterator[K, V]{cursor: m.head, bof: true}
}

// Implement fmt.Stringer
func (m *OMapLinkedHash[K, V]) String() string {
	return toString[K, V]("omap.OMapLinkedHash", m.Iterator())
}

// Implement json.Marshaler interface.
func (m OMapLinkedHash[K, V]) MarshalJSON() ([]byte, error) {
	buffer, err := marshalJSON(m.Iterator())
	return buffer, err
}

// Implement json.Unmarshaler interface.
func (m *OMapLinkedHash[K, V]) UnmarshalJSON(b []byte) error {
	m.init()
	return unmarshalJSON[K, V](m, b)
}

func (it *OMapLinkedHashIterator[K, V]) Next() bool {
	if !it.bof {
		it.cursor = it.cursor.next
	} else {
		it.bof = false
	}
	return it.cursor != nil
}

func (it *OMapLinkedHashIterator[K, V]) EOF() bool {
	return !it.bof && it.cursor == nil
}

func (it *OMapLinkedHashIterator[K, V]) Key() K {
	return *it.cursor.key
}

func (it *OMapLinkedHashIterator[K, V]) Value() V {
	return it.cursor.value
}

/*
// hashString computes the Fowler–Noll–Vo hash of s.
// Copy from go/src/cmd/vendor/golang.org/x/tools/go/types/typeutil/map.go:252
func hashString(s string) uint32 {
	var h uint32
	if len(s) < 32 {
		for i := 0; i < len(s); i++ {
			h ^= uint32(s[i])
			h *= 16777619
		}
	} else {
		for i := 0; i < len(s); i += 2 {
			h ^= uint32(s[i])
			h *= 16777619
		}
	}
	return h
}
*/
