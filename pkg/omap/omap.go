package omap

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

//// Interfaces ////

// Iterator over OMap.
type OMapIterator[K comparable, V any] interface {
	// Iterate to the next record, returning true if the record was found and false otherwise.
	Next() bool
	// Returns true if the iterator is past the last record.
	EOF() bool
	// Returns the key pointing to the current record.
	Key() K
	// Returns the value pointing to the current record.
	Value() V
}

// OMap is an ordered map that holds key/value and is able to iterate over the whole data-set
// in the same order as insertion has happened.
type OMap[K comparable, V any] interface {
	// Add or update an element in the map of given key and value. If it is a new value, it should be
	// in the end of the map on iteration, if it is an update the position of the value must be
	// maintained.
	Put(key K, value V)
	// Get the value pointing by key, if found ok is true, or false otherwise.
	Get(key K) (value V, ok bool)
	// Delete the value pointing by key.
	Delete(key K)
	// Returns the iterator of this map, at the beginning.
	Iterator() OMapIterator[K, V]
}

//// Utility functions ////

func marshalJSON[K comparable, V any](it OMapIterator[K, V]) ([]byte, error) {
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

func unmarshalJSON[K comparable, V any](m OMap[K, V], b []byte) error {
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
			m.Put(key, value)
		}
	}
	return nil
}

func toString[K comparable, V any](typeName string, it OMapIterator[K, V]) string {
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

//// Common structs ////

type mapEntry[K comparable, V any] struct {
	key   K
	value V
	next  *mapEntry[K, V]
	prev  *mapEntry[K, V]
}
