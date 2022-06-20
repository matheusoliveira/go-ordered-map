package omultimap

import (
	"errors"

	"github.com/matheusoliveira/go-ordered-map/pkg/omap"
)

func New[K comparable, V any]() OMultiMap[K, V] {
	return NewOMultiMapLinked[K, V]()
}

type mapEntry[K comparable, V any] struct {
	key   K
	value V
	next  *mapEntry[K, V]
	prev  *mapEntry[K, V]
}

type OMultiMapLinked[K comparable, V any] struct {
	m    map[K][]*mapEntry[K, V]
	head *mapEntry[K, V]
	tail *mapEntry[K, V]
}

type OMultiMapLinkedIterator[K comparable, V any] struct {
	m      *OMultiMapLinked[K, V]
	cursor *mapEntry[K, V]
	bof    bool
}

type OMultiMapLinkedValuesIterator[K comparable, V any] struct {
	elems []*mapEntry[K, V]
	pos   int
}

func NewOMultiMapLinked[K comparable, V any]() OMultiMap[K, V] {
	ret := &OMultiMapLinked[K, V]{}
	ret.init()
	return ret
}

func (m *OMultiMapLinked[K, V]) init() {
	m.m = make(map[K][]*mapEntry[K, V])
}

func (m *OMultiMapLinked[K, V]) Put(key K, value V) {
	entry := &mapEntry[K, V]{
		key:   key,
		value: value,
		next:  nil,
		prev:  m.tail,
	}
	if elems, ok := m.m[key]; ok {
		m.m[key] = append(elems, entry)
	} else {
		m.m[key] = []*mapEntry[K, V]{entry}
	}
	if m.head == nil {
		m.head = entry
		m.tail = entry
	} else {
		m.tail.next = entry
		m.tail = entry
	}
}

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

func (m *OMultiMapLinked[K, V]) DeleteAll(key K) {
	if elems, ok := m.m[key]; ok {
		for _, e := range elems {
			m.deleteEntryInList(e)
		}
		delete(m.m, key)
	}
}

func (m *OMultiMapLinked[K, V]) DeleteAt(interfaceIt omap.OMapIterator[K, V]) error {
	if it, ok := interfaceIt.(*OMultiMapLinkedIterator[K, V]); !ok {
		return errors.New("trying to operate on invalid map iterator")
	} else if it.m != m {
		return errors.New("trying to operate on different map iterator")
	} else if it.bof || it.cursor == nil {
		return errors.New("iterator not positionated")
	} else if elems, ok := m.m[it.Key()]; !ok {
		return errors.New("inconsistent state, key not found")
	} else {
		found := false
		pos := 0
		for pos = range elems {
			if elems[pos] == it.cursor {
				found = true
				break
			}
		}
		if found {
			if len(elems) == 1 {
				delete(m.m, it.Key())
			} else {
				m.m[it.Key()] = append(elems[0:pos], elems[pos+1:]...)
			}
			m.deleteEntryInList(it.cursor)
		} else {
			return errors.New("inconsistent state, key found but value not present")
		}
	}
	return nil
}

func (m *OMultiMapLinked[K, V]) MustDeleteAt(interfaceIt omap.OMapIterator[K, V]) {
	err := m.DeleteAt(interfaceIt)
	if err != nil {
		panic(err)
	}
}

func (m *OMultiMapLinked[K, V]) Iterator() omap.OMapIterator[K, V] {
	return &OMultiMapLinkedIterator[K, V]{cursor: m.head, bof: true, m: m}
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
	return omap.UnmarshalJSON[K, V](m.Put, b)
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
