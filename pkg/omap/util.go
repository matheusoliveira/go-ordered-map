package omap

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)


//// Utility functions ////

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

func IteratorKeysToSlice[K comparable, V any](it OMapIterator[K, V]) []K {
	ret := make([]K, 0)
	for it.Next() {
		ret = append(ret, it.Key())
	}
	return ret
}

func IteratorValuesToSlice[K comparable, V any](it OMapIterator[K, V]) []V {
	ret := make([]V, 0)
	for it.Next() {
		ret = append(ret, it.Value())
	}
	return ret
}
