package omap

// Create a new map using the default implementation, which is considered the best trade-off among
// all. Currently, OMapLinked is the winner.
func New[K comparable, V any]() OMap[K, V] {
	return NewOMapLinked[K, V]()
}

//// OMapLinked ////

// Implements an ordered map using double-linked list for iteration.
type OMapLinked[K comparable, V any] struct {
	m    map[K]*mapEntry[K, V]
	head *mapEntry[K, V]
	tail *mapEntry[K, V]
}

// Implements OMapIterator for OMapLinked.
type OMapLinkedIterator[K comparable, V any] struct {
	m      *OMapLinked[K, V]
	cursor *mapEntry[K, V]
	bof    bool
}

// Return a new OMap based on OMapLinked implementation, see OMapLinked type for more
// details of the implementation.
func NewOMapLinked[K comparable, V any]() OMap[K, V] {
	m := &OMapLinked[K, V]{}
	m.init()
	return m
}

func (m *OMapLinked[K, V]) init() {
	m.m = make(map[K]*mapEntry[K, V])
	m.head = nil
	m.tail = nil
}

func (m *OMapLinked[K, V]) Put(key K, value V) {
	oldEntry, found := m.m[key]
	if found {
		// overwrite in place
		oldEntry.value = value
	} else {
		// insert at the end
		entry := &mapEntry[K, V]{
			key:   key,
			value: value,
			next:  nil,
			prev:  m.tail,
		}
		m.m[key] = entry
		if m.head == nil {
			m.head = entry
			m.tail = entry
		} else {
			m.tail.next = entry
			m.tail = entry
		}
	}
}

func (m *OMapLinked[K, V]) Get(key K) (V, bool) {
	var val V
	v, ok := m.m[key]
	if ok {
		val = v.value
	}
	return val, ok
}

func (m *OMapLinked[K, V]) Delete(key K) {
	v, ok := m.m[key]
	if ok {
		if m.head == v {
			m.head = v.next
		}
		if m.tail == v {
			m.tail = v.prev
		}
		if v.prev != nil {
			v.prev.next = v.next
		}
		if v.next != nil {
			v.next.prev = v.prev
		}
		delete(m.m, key)
	}
}

func (m *OMapLinked[K, V]) Iterator() OMapIterator[K, V] {
	return &OMapLinkedIterator[K, V]{m: m, cursor: m.head, bof: true}
}

func (m *OMapLinked[K, V]) Len() int {
	return len(m.m)
}

// Implement fmt.Stringer
func (m *OMapLinked[K, V]) String() string {
	return IteratorToString[K, V]("omap.OMapLinked", m.Iterator())
}

// Implement json.Marshaler interface.
func (m OMapLinked[K, V]) MarshalJSON() ([]byte, error) {
	buffer, err := MarshalJSON(m.Iterator())
	return buffer, err
}

// Implement json.Unmarshaler interface.
func (m *OMapLinked[K, V]) UnmarshalJSON(b []byte) error {
	m.init()
	return UnmarshalJSON[K, V](m.Put, b)
}

func (it *OMapLinkedIterator[K, V]) Next() bool {
	if !it.bof {
		it.cursor = it.cursor.next
	} else {
		it.bof = false
	}
	//return it.cursor != nil
	return it.IsValid()
}

func (it *OMapLinkedIterator[K, V]) EOF() bool {
	return !it.bof && it.cursor == nil
}

func (it *OMapLinkedIterator[K, V]) Key() K {
	return it.cursor.key
}

func (it *OMapLinkedIterator[K, V]) Value() V {
	return it.cursor.value
}

func (it *OMapLinkedIterator[K, V]) IsValid() bool {
	return !it.bof && it.cursor != nil
}

func (it *OMapLinkedIterator[K, V]) MoveFront() OMapIterator[K, V] {
	it.bof = true
	it.cursor = it.m.head
	return it
}

func (it *OMapLinkedIterator[K, V]) MoveBack() OMapIterator[K, V] {
	it.bof = false
	it.cursor = nil
	return it
}

func (it *OMapLinkedIterator[K, V]) Prev() bool {
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
