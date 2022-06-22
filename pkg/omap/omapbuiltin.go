package omap

import (
	"encoding/json"
)

// This is a safe var, since OMapBuiltin should be used only for testings, since it is not
// actually an ordered map, one must explicitly set this variable to `true` before using
// it, or it will panic on initialization.
var EnableOMapBuiltin = false

//// OMapBuiltin ////

// DO NOT USE THIS FOR REAL!!!
// Implements OMap interface but not very strictly, should be use only for comparison
// with builtin map
type OMapBuiltin[K comparable, V any] struct {
	m map[K]V
}

type OMapBuiltinData[K comparable, V any] struct {
	key   K
	value V
}

// Iterator over a OMapSimple, should be created through OMapSimple.Iterator() function.
type OMapBuiltinIterator[K comparable, V any] struct {
	ch    chan OMapBuiltinData[K, V]
	m     map[K]V
	eof   bool
	key   K
	value V
}

func NewOMapBuiltin[K comparable, V any]() OMap[K, V] {
	var m OMapBuiltin[K, V]
	m.init()
	return &m
}

func (m *OMapBuiltin[K, V]) init() {
	if !EnableOMapBuiltin {
		panic("OMapBuiltin is not ordered, should be used only for testing, if you really want to use it do \"omap.EnableOMapBuiltin = true\" before using")
	}
	m.m = make(map[K]V)
}

func (m *OMapBuiltin[K, V]) Put(key K, value V) {
	m.m[key] = value
}

func (m *OMapBuiltin[K, V]) Get(key K) (V, bool) {
	v, ok := m.m[key]
	return v, ok
}

func (m *OMapBuiltin[K, V]) Delete(key K) {
	delete(m.m, key)
}

func (m *OMapBuiltin[K, V]) Iterator() OMapIterator[K, V] {
	ch := make(chan OMapBuiltinData[K, V], 1)
	it := &OMapBuiltinIterator[K, V]{
		ch: ch,
		m:  m.m,
	}
	go func() {
		defer close(ch)
		for key, value := range m.m {
			ch <- OMapBuiltinData[K, V]{key, value}
		}
	}()
	return it
}

func (m *OMapBuiltin[K, V]) Len() int {
	return len(m.m)
}

// Implement fmt.Stringer
func (m *OMapBuiltin[K, V]) String() string {
	return IteratorToString[K, V]("omap.OMapBuiltin", m.Iterator())
}

// Implement json.Marshaler interface.
func (m OMapBuiltin[K, V]) MarshalJSON() ([]byte, error) {
	buffer, err := json.Marshal(m.m)
	return buffer, err
}

// Implement json.Unmarshaler interface.
func (m OMapBuiltin[K, V]) UnmarshalJSON(b []byte) error {
	m.init()
	return json.Unmarshal(b, &m.m)
}

func (it *OMapBuiltinIterator[K, V]) Next() bool {
	data, ok := <-it.ch
	if !ok {
		it.eof = true
		return false
	}
	it.key = data.key
	it.value = data.value
	return true
}

func (it *OMapBuiltinIterator[K, V]) EOF() bool {
	return it.eof
}

func (it *OMapBuiltinIterator[K, V]) Key() K {
	return it.key
}

func (it *OMapBuiltinIterator[K, V]) Value() V {
	return it.value
}

func (it OMapBuiltinIterator[K, V]) IsValid() bool {
	return !it.eof
}

func (it *OMapBuiltinIterator[K, V]) MoveFront() OMapIterator[K, V] {
	panic("not implemented")
}

func (it *OMapBuiltinIterator[K, V]) MoveBack() OMapIterator[K, V] {
	panic("not implemented")
}

func (it *OMapBuiltinIterator[K, V]) Prev() bool {
	panic("not implemented")
}
