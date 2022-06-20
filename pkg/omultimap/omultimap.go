// omultimap package provides common interface to implement ordered multimap, along with an
// small set of implementations.
//
// An omultimap is very similar to a omap, it also keeps the map in the insertion order when
// iterating/marshaling, but a given key can hold many values.
package omultimap

import (
	"github.com/matheusoliveira/go-ordered-map/pkg/omap"
)

type OMultiMap[K comparable, V any] interface {
	// Add a given key/value to the map.
	Put(key K, value V)
	// Get an iterator over all values of a given key.
	GetValuesOf(key K) omap.OMapIterator[K, V]
	// Delete all values stored by a giving key.
	DeleteAll(key K)
	// Delete the value currently pointed by the iterator, returning a non-nil error if failed.
	DeleteAt(omap.OMapIterator[K, V]) error
	// Same as DeleteAt but with panic in case of failure.
	MustDeleteAt(omap.OMapIterator[K, V])
	// Return an iterator at the beginning of the map.
	Iterator() omap.OMapIterator[K, V]
	// Returns the len of the map, similar to builtin len(map)
	Len() int
}
