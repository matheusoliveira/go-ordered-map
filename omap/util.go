package omap

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Iterate over the given iterator it, from the given position, and marshal the key/values into
// JSON.
//
// This is a handy function to construct a json.Marshaler implementation.
// Note: the iterator will be at EOF after this function returns with success.
func MarshalJSON[K comparable, V any](it OMapIterator[K, V]) ([]byte, error) {
	var w bytes.Buffer
	w.WriteString("{")
	first := true
	for it.Next() {
		key, err := json.Marshal(it.Key())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal key: %w", err)
		}
		val, err := json.Marshal(it.Value())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal value: %w", err)
		}
		if first {
			first = false
			w.Grow(len(key) + len(val) + 1)
		} else {
			w.Grow(len(key) + len(val) + 2)
			w.WriteString(",")
		}
		w.Write(key)
		w.WriteString(":")
		w.Write(val)
	}
	w.WriteString("}")
	return w.Bytes(), nil
}

// Process given json at b and for each key/value found, call given putFunc function with same
// definition of OMap.Put to add the given key/value into a map.
//
// This is a handy function to convert a given json to map as json.Unmarshaler interface requires.
func UnmarshalJSON[K comparable, V any](putFunc func(K, V), b []byte) error {
	reader := bytes.NewBuffer(b)
	dec := json.NewDecoder(reader)
	t, err := dec.Token()
	if err != nil {
		return fmt.Errorf("failed to get first token: %w", err)
	}
	if d, ok := t.(json.Delim); !ok || d.String() != "{" {
		return errors.New("JSON input does not start with \"{\"")
	}
	for dec.More() {
		// Get key
		keyToken, err := dec.Token()
		if err != nil {
			return fmt.Errorf("failed to get key token: %w", err)
		}
		if key, ok := keyToken.(K); !ok {
			return fmt.Errorf("could not parse token, wrong type of: %v", keyToken)
		} else {
			// Get value
			var value V
			err := dec.Decode(&value)
			if err != nil {
				return fmt.Errorf("could not decode value: %w", err)
			}
			putFunc(key, value)
		}
	}
	return nil
}

// Iterate over the given iterator it, from the given position, and marshal the key/values into
// an string.
//
// This is a handy function to implements fmt.Stringfier interface.
// Note: the iterator will be at EOF after this function returns with success.
func IteratorToString[K comparable, V any](typeName string, it OMapIterator[K, V]) string {
	var b strings.Builder
	b.Grow(len(typeName) + 1)
	b.WriteString(typeName)
	b.WriteString("[")
	first := true
	for it.Next() {
		ks := fmt.Sprint(it.Key())
		vs := fmt.Sprint(it.Value())
		if !first {
			b.WriteString(" ")
		} else {
			first = false
		}
		b.Grow(len(ks) + len(vs) + 2)
		b.WriteString(ks)
		b.WriteString(":")
		b.WriteString(vs)
	}
	b.WriteString("]")
	return b.String()
}

// Iterate over the given iterator it, from the given position, and store the keys only into an
// slice of same type.
//
// Note: the iterator will be at EOF after this function returns with success.
func IteratorKeysToSlice[K comparable, V any](it OMapIterator[K, V]) []K {
	ret := make([]K, 0)
	for it.Next() {
		ret = append(ret, it.Key())
	}
	return ret
}

// Iterate over the given iterator it, from the given position, and store the values only into an
// slice of same type.
//
// Note: the iterator will be at EOF after this function returns with success.
func IteratorValuesToSlice[K comparable, V any](it OMapIterator[K, V]) []V {
	ret := make([]V, 0)
	for it.Next() {
		ret = append(ret, it.Value())
	}
	return ret
}

// Move targetKey in the m map at first position of the map. Returns omap.ErrKeyNotFound if the key
// cannot be found in the map, or nil otherwise.
func MoveFirst[K comparable, V any](m OMap[K, V], targetKey K) error {
	itTarget := m.GetIteratorAt(targetKey)
	if !itTarget.IsValid() {
		return fmt.Errorf("%w: targetKey = \"%v\"", ErrKeyNotFound, targetKey)
	}
	return m.PutAfter(m.Iterator(), itTarget.Key(), itTarget.Value())
}

// Move targetKey in the m map at last position of the map. Returns omap.ErrKeyNotFound if the key
// cannot be found in the map, or nil otherwise.
func MoveLast[K comparable, V any](m OMap[K, V], targetKey K) error {
	itTarget := m.GetIteratorAt(targetKey)
	if !itTarget.IsValid() {
		return fmt.Errorf("%w: targetKey = \"%v\"", ErrKeyNotFound, targetKey)
	}
	itLast := m.Iterator().MoveBack()
	itLast.Prev()
	return m.PutAfter(itLast, itTarget.Key(), itTarget.Value())
}

// Finds targetKey and refKey in the m map, and move the targetKey entry to the position
// immediately after the refKey. Returns omap.ErrKeyNotFound if any of the keys cannot be found in
// the map, or nil otherwise.
func MoveAfter[K comparable, V any](m OMap[K, V], targetKey K, refKey K) error {
	itTarget := m.GetIteratorAt(targetKey)
	if !itTarget.IsValid() {
		return fmt.Errorf("%w: targetKey = \"%v\"", ErrKeyNotFound, targetKey)
	}
	itRef := m.GetIteratorAt(refKey)
	if !itRef.IsValid() {
		return fmt.Errorf("%w: refKey = \"%v\"", ErrKeyNotFound, refKey)
	}
	return m.PutAfter(itRef, itTarget.Key(), itTarget.Value())
}

// Finds targetKey and refKey in the m map, and move the targetKey entry to the position
// immediately before the refKey. Returns omap.ErrKeyNotFound if any of the keys cannot be found
// in the map, or nil otherwise.
func MoveBefore[K comparable, V any](m OMap[K, V], targetKey K, refKey K) error {
	itTarget := m.GetIteratorAt(targetKey)
	if !itTarget.IsValid() {
		return fmt.Errorf("%w: targetKey = \"%v\"", ErrKeyNotFound, targetKey)
	}
	itRef := m.GetIteratorAt(refKey)
	if !itRef.IsValid() {
		return fmt.Errorf("%w: refKey = \"%v\"", ErrKeyNotFound, refKey)
	}
	itRef.Prev()
	return m.PutAfter(itRef, itTarget.Key(), itTarget.Value())
}
