package omultimap

import (
	"errors"
	"fmt"

	"github.com/matheusoliveira/go-ordered-map/pkg/omap"
)

// Create a new OMultiMap using default implementation, currently a OMultiMapLinked.
func New[K comparable, V any]() OMultiMap[K, V] {
	return NewOMultiMapLinked[K, V]()
}

type mapEntry[K comparable, V any] struct {
	key   K
	value V
	next  *mapEntry[K, V]
	prev  *mapEntry[K, V]
}

// OMultiMapLinked implements an OMultiMap using a linked list to navigate through the key/values
// in same order as originally inserted.
type OMultiMapLinked[K comparable, V any] struct {
	m      map[K][]*mapEntry[K, V]
	head   *mapEntry[K, V]
	tail   *mapEntry[K, V]
	length int
}

// Iterator for OMultiMapLinked.
type OMultiMapLinkedIterator[K comparable, V any] struct {
	m      *OMultiMapLinked[K, V]
	cursor *mapEntry[K, V]
	bof    bool
}

// Values iterator for OMultiMapLinked.
type OMultiMapLinkedValuesIterator[K comparable, V any] struct {
	elems []*mapEntry[K, V]
	pos   int
}

// Create a new OMultiMapLinked.
func NewOMultiMapLinked[K comparable, V any]() OMultiMap[K, V] {
	ret := &OMultiMapLinked[K, V]{}
	ret.init()
	return ret
}

func (m *OMultiMapLinked[K, V]) init() {
	m.m = make(map[K][]*mapEntry[K, V])
}

// Add a given key/value to the map.
// Complexity: O(1), for each value in values slice.
func (m *OMultiMapLinked[K, V]) Put(key K, values ...V) {
	if len(values) == 0 {
		return
	}
	var buffer []*mapEntry[K, V]
	if elems, ok := m.m[key]; ok {
		// grow the slice (copy seems to be faster than append)
		tmp := make([]*mapEntry[K, V], len(elems)+len(values))
		copy(tmp, elems)
		m.m[key] = tmp
		buffer = tmp[len(elems):]
	} else {
		buffer = make([]*mapEntry[K, V], len(values))
		m.m[key] = buffer
	}
	m.length += len(values)
	prev := m.tail
	for i, value := range values {
		entry := &mapEntry[K, V]{
			key:   key,
			value: value,
			next:  nil,
			prev:  prev,
		}
		buffer[i] = entry
		if prev != nil {
			prev.next = entry
		}
		prev = entry
	}
	if m.head == nil {
		m.head = buffer[0]
		m.tail = prev
	} else {
		m.tail = prev
	}
}

// Get an iterator over all values of a given key.
// Complexity: O(1).
func (m *OMultiMapLinked[K, V]) GetValuesOf(key K) omap.OMapIterator[K, V] {
	elems := m.m[key]
	return &OMultiMapLinkedValuesIterator[K, V]{
		elems: elems,
		pos:   -1,
	}
}

func (m *OMultiMapLinked[K, V]) deleteEntryInList(entry *mapEntry[K, V]) {
	if m.head == entry {
		m.head = entry.next
	}
	if m.tail == entry {
		m.tail = entry.next
	}
	if entry.prev != nil {
		entry.prev.next = entry.next
	}
	if entry.next != nil {
		entry.next.prev = entry.prev
	}
}

// Delete all values stored by a giving key.
// Complexity: O(m) where m is the number of values pointing to the given key.
func (m *OMultiMapLinked[K, V]) DeleteAll(key K) {
	if elems, ok := m.m[key]; ok {
		for _, e := range elems {
			m.deleteEntryInList(e)
		}
		m.length -= len(elems)
		delete(m.m, key)
	}
}

func (m *OMultiMapLinked[K, V]) getIteratorEntry(interfaceIt omap.OMapIterator[K, V]) (it *OMultiMapLinkedIterator[K, V], elems []*mapEntry[K, V], pos int, err error) {
	var ok bool
	elems = nil
	pos = -1
	if it, ok = interfaceIt.(*OMultiMapLinkedIterator[K, V]); !ok {
		err = fmt.Errorf("%w - expected OMultiMapLinkedIterator found %T", omap.ErrInvalidIteratorType, interfaceIt)
		return
	} else if it.m != m {
		err = omap.ErrInvalidIteratorMap
		return
	} else if it.bof || it.cursor == nil {
		err = omap.ErrInvalidIteratorPos
		return
	} else if elems, ok = m.m[it.Key()]; !ok {
		err = fmt.Errorf("%w - key not found", omap.ErrInvalidIteratorKey)
		return
	} else {
		found := false
		for pos = range elems {
			if elems[pos] == it.cursor {
				found = true
				break
			}
		}
		if !found {
			pos = -1
			err = fmt.Errorf("%w - key found but specific entry not present", omap.ErrInvalidIteratorKey)
			return
		} else {
			// all good
			return
		}
	}
}

// Delete the value currently pointed by the iterator, returning a non-nil error if failed.
// Complexity: O(1).
func (m *OMultiMapLinked[K, V]) DeleteAt(interfaceIt omap.OMapIterator[K, V]) error {
	it, elems, pos, err := m.getIteratorEntry(interfaceIt)
	if err != nil {
		return err
	}
	if len(elems) == 1 {
		delete(m.m, it.Key())
	} else {
		m.m[it.Key()] = append(elems[0:pos], elems[pos+1:]...)
	}
	m.deleteEntryInList(it.cursor)
	m.length--
	return nil
}

// Add a given key/value to the map, after the entry pointed by iterator.
// Complexity: O(1).
func (m *OMultiMapLinked[K, V]) PutAfter(interfaceIt omap.OMapIterator[K, V], key K, value V) error {
	it, elems, pos, err := m.getIteratorEntry(interfaceIt)
	if err != nil {
		if !(errors.Is(err, omap.ErrInvalidIteratorPos) && it != nil && it.bof) {
			return err
		}
	}
	entry := &mapEntry[K, V]{
		key:   key,
		value: value,
	}
	if !it.bof {
		entry.prev = elems[pos]
		entry.next = entry.prev.next
		if entry.prev.next != nil {
			entry.prev.next.prev = entry
		}
		entry.prev.next = entry
	} else {
		entry.prev = nil
		entry.next = m.head
	}
	if it.IsValid() && it.Key() == key {
		tmp := make([]*mapEntry[K, V], 0, len(elems) + 1)
		tmp = append(tmp, elems[0:pos+1]...)
		tmp = append(tmp, entry)
		tmp = append(tmp, elems[pos+1:]...)
		m.m[key] = tmp
	} else if elemsK, ok := m.m[key]; ok {
		m.m[key] = append(elemsK, entry)
	} else {
		m.m[key] = []*mapEntry[K, V]{entry}
	}
	// update map head and tail
	if m.head == nil {
		m.head = entry
		m.tail = entry
	} else if it.bof {
		m.head.prev = entry
		m.head = entry
	} else if m.tail == entry.prev {
		m.tail = entry
	}
	return nil
}

// Same as DeleteAt but with panic in case of failure.
// Complexity: O(1).
func (m *OMultiMapLinked[K, V]) MustDeleteAt(interfaceIt omap.OMapIterator[K, V]) {
	err := m.DeleteAt(interfaceIt)
	if err != nil {
		panic(err)
	}
}

// Return an iterator at the beginning of the map.
func (m *OMultiMapLinked[K, V]) Iterator() omap.OMapIterator[K, V] {
	return &OMultiMapLinkedIterator[K, V]{cursor: m.head, bof: true, m: m}
}

// Returns the length of the map.
// Complexity: O(1).
func (m *OMultiMapLinked[K, V]) Len() int {
	return m.length
}

// Implement fmt.Stringer
func (m *OMultiMapLinked[K, V]) String() string {
	return omap.IteratorToString[K, V]("omultimap.OMultiMapLinked", m.Iterator())
}

// Implement json.Marshaler interface.
func (m OMultiMapLinked[K, V]) MarshalJSON() ([]byte, error) {
	buffer, err := omap.MarshalJSON(m.Iterator())
	return buffer, err
}

// Implement json.Unmarshaler interface.
func (m *OMultiMapLinked[K, V]) UnmarshalJSON(b []byte) error {
	m.init()
	return omap.UnmarshalJSON[K, V](func(key K, val V) { m.Put(key, val) }, b)
}

//// OMultiMap Iterator ////

func (it *OMultiMapLinkedIterator[K, V]) Next() bool {
	if !it.bof {
		it.cursor = it.cursor.next
	} else {
		it.bof = false
	}
	return it.cursor != nil
}

func (it *OMultiMapLinkedIterator[K, V]) EOF() bool {
	return !it.bof && it.cursor == nil
}

func (it *OMultiMapLinkedIterator[K, V]) Key() K {
	return it.cursor.key
}

func (it *OMultiMapLinkedIterator[K, V]) Value() V {
	return it.cursor.value
}

func (it *OMultiMapLinkedIterator[K, V]) IsValid() bool {
	return !it.bof && it.cursor != nil
}

func (it *OMultiMapLinkedIterator[K, V]) MoveFront() omap.OMapIterator[K, V] {
	it.bof = true
	it.cursor = it.m.head
	return it
}

func (it *OMultiMapLinkedIterator[K, V]) MoveBack() omap.OMapIterator[K, V] {
	it.bof = false
	it.cursor = nil
	return it
}

func (it *OMultiMapLinkedIterator[K, V]) Prev() bool {
	if it.bof {
		return false
	} else if it.cursor == nil {
		it.cursor = it.m.tail
	} else {
		it.cursor = it.cursor.prev
	}
	if it.cursor == nil {
		it.bof = true
	}
	return it.IsValid()
}

//// Values Iterator ////

func (it *OMultiMapLinkedValuesIterator[K, V]) Next() bool {
	if it.elems == nil {
		return false
	}
	it.pos++
	return it.pos < len(it.elems)
}

func (it *OMultiMapLinkedValuesIterator[K, V]) EOF() bool {
	return it.elems == nil || it.pos >= len(it.elems)
}

func (it *OMultiMapLinkedValuesIterator[K, V]) Key() K {
	return it.elems[it.pos].key
}

func (it *OMultiMapLinkedValuesIterator[K, V]) Value() V {
	return it.elems[it.pos].value
}

func (it OMultiMapLinkedValuesIterator[K, V]) IsValid() bool {
	return it.pos >= 0 && it.pos < len(it.elems)
}

func (it *OMultiMapLinkedValuesIterator[K, V]) MoveFront() omap.OMapIterator[K, V] {
	it.pos = -1
	return it
}

func (it *OMultiMapLinkedValuesIterator[K, V]) MoveBack() omap.OMapIterator[K, V] {
	it.pos = len(it.elems)
	return it
}

func (it *OMultiMapLinkedValuesIterator[K, V]) Prev() bool {
	it.pos--
	return it.IsValid()
}
