package omaptestsutil

import (

	"github.com/matheusoliveira/go-ordered-map/pkg/omap"
)

type KeyValue[K comparable, V comparable] struct {
	Key   K
	Value V
}

type MockableTesting interface {
	Helper()
	Errorf(format string, args ...any)
}

func ValidateIterator[K comparable, V comparable](t MockableTesting, it omap.OMapIterator[K, V], isOrdered bool, expResults []KeyValue[K, V]) bool {
	t.Helper()
	startValid := it.IsValid() // check if started as valid iterator (not at BOF), to skip checking boundaries at backward validation
	forward := ValidateIteratorForward(t, it, isOrdered, expResults)
	if forward && isOrdered {
		return ValidateIteratorBackward(t, it, !startValid, expResults)
	} else {
		return forward
	}
}

func ValidateIteratorForward[K comparable, V comparable](t MockableTesting, it omap.OMapIterator[K, V], isOrdered bool, expResults []KeyValue[K, V]) bool {
	t.Helper()
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

func ValidateIteratorBackward[K comparable, V comparable](t MockableTesting, it omap.OMapIterator[K, V], shouldBeInvalidAtEnd bool, expResults []KeyValue[K, V]) bool {
	// Validate it backwards now
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
