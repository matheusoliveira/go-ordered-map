// This package offers some helper functions to make easier to write unit tests with iterators
// and maps.
package testhelper

import (
	"encoding/json"
	"fmt"

	"github.com/matheusoliveira/go-ordered-map/pkg/omap"
)

type KeyValue[K comparable, V comparable] struct {
	Key   K
	Value V
}

func JsonToKVErr[K comparable, V comparable](in string) ([]KeyValue[K, V], error) {
	var buffer [][]json.RawMessage
	if err := json.Unmarshal([]byte(in), &buffer); err != nil {
		return nil, fmt.Errorf("top-level decode failed: %w", err)
	}
	ret := make([]KeyValue[K, V], len(buffer))
	for i := range buffer {
		var key K
		var value V
		if len(buffer[i]) != 2 {
			return nil, fmt.Errorf("expected array of len=2 at position %d, found len %d", i, len(buffer[i]))
		}
		if err := json.Unmarshal(buffer[i][0], &key); err != nil {
			return nil, fmt.Errorf("decode of key failed at position %d: %w", i, err)
		}
		if err := json.Unmarshal(buffer[i][1], &value); err != nil {
			return nil, fmt.Errorf("decode of value failed at position %d: %w", i, err)
		}
		ret[i] = KeyValue[K, V]{Key: key, Value: value}
	}
	return ret, nil
}

// Convert a JSON array in form of `[[key1, val1], [key2, val2], ...]` to []KeyValue[K, V].
// Returns nil in case of error, since validator functions will treat this specifically. One
// can use JsonToKVErr to debug the actual error.
// Notice: this function is intentionally designed to be short to build a []KeyValue from a
// hard-coded constant, so it is not well designed to be safe nor fast.
func JsonToKV[K comparable, V comparable](in string) []KeyValue[K, V] {
	ret, err := JsonToKVErr[K, V](in)
	if err != nil {
		return nil
	} else {
		return ret
	}
}

func ValidateIterator[K comparable, V comparable](t TestingT, it omap.OMapIterator[K, V], isOrdered bool, expResults []KeyValue[K, V]) bool {
	t.Helper()
	startValid := it.IsValid() // check if started as valid iterator (not at BOF), to skip checking boundaries at backward validation
	forward := ValidateIteratorForward(t, it, isOrdered, expResults)
	if forward && isOrdered {
		// Validate it backwards now
		return ValidateIteratorBackward(t, it, !startValid, expResults)
	} else {
		return forward
	}
}

func ValidateIteratorForward[K comparable, V comparable](t TestingT, it omap.OMapIterator[K, V], isOrdered bool, expResults []KeyValue[K, V]) bool {
	t.Helper()
	if expResults == nil {
		t.Errorf("expResults slices is nil")
		return false
	}
	i := 0
	for it.Next() {
		if i >= len(expResults) {
			t.Errorf("overflow, expecting max of %d values", len(expResults))
			return false
		}
		key := it.Key()
		value := it.Value()
		exp := expResults[i]
		if !isOrdered {
			found := false
			for _, e := range expResults {
				if e.Key == key {
					found = true
					exp = e
					break
				}
			}
			if !found {
				t.Errorf("found unexpected key/value pair: %v/%v", key, value)
				return false
			}
		}
		if exp.Key != key {
			t.Errorf("expecting key \"%v\" at position %d, key \"%v\" found", exp.Key, i, key)
			return false
		}
		if exp.Value != value {
			t.Errorf("invalid value for key \"%v\", at position %d, expected %v, found %v", key, i, exp.Value, value)
			return false
		}
		i++
	}
	if i != len(expResults) {
		t.Errorf("values not processed: %v", expResults[i:])
		return false
	}
	ret := true
	if !it.EOF() {
		t.Errorf("EOF returned false instead of true at end of loop")
		ret = false
	}
	if it.IsValid() {
		t.Errorf("IsValid returned true instead of false at end of loop")
		ret = false
	}
	return ret
}

func ValidateIteratorBackward[K comparable, V comparable](t TestingT, it omap.OMapIterator[K, V], shouldBeInvalidAtEnd bool, expResults []KeyValue[K, V]) bool {
	t.Helper()
	if expResults == nil {
		t.Errorf("expResults slices is nil")
		return false
	}
	i := len(expResults) - 1
	for it.Prev() {
		if i < 0 {
			if !shouldBeInvalidAtEnd {
				break
			}
			t.Errorf("backward validation - overflow, expecting max of %d values", len(expResults))
			return false
		}
		key := it.Key()
		value := it.Value()
		exp := expResults[i]
		if exp.Key != key {
			t.Errorf("backward validation - expecting key \"%v\" at position %d, key \"%v\" found", exp.Key, i, key)
			return false
		}
		if exp.Value != value {
			t.Errorf("backward validation - invalid value for key \"%v\", at position %d, expected %v, found %v", key, i, exp.Value, value)
			return false
		}
		i--
	}
	ret := true
	if i >= 0 {
		t.Errorf("backward validation - values not processed: %v", expResults[:i+1])
		ret = false
	}
	if shouldBeInvalidAtEnd && it.IsValid() {
		t.Errorf("backward validation - IsValid returned true instead of false at end of loop")
		ret = false
	}
	return ret
}

func SlicesToKeyValue[K comparable, V comparable](keys []K, values []V) []KeyValue[K, V] {
	if len(keys) != len(values) {
		return nil
	}
	ret := make([]KeyValue[K, V], len(keys))
	for i := 0; i < len(keys); i++ {
		ret[i] = KeyValue[K, V]{Key: keys[i], Value: values[i]}
	}
	return ret
}
