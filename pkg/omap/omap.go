// omap package provides a common interface to implement ordered maps, along with a set of
// implementations.
//
// As it is well-known, Go's builtin map iterator order is undefined (consider random).
// An ordered map (omap for short) is very similar to a builtin map, but with the advantage that
// iterating the keys or serializing it (e.g. to JSON) will keep the keys in same order as
// originally provided.
// There is an small overhead though, mostly in memory, but usually negligible. See benchmarks for
// a few comparisons among omap implementations and builtin map.
// A given key can hold only a single value, see omultimap if you need multiple values for the same
// key.
package omap

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
	// Returns the len of the map, similar to builtin len(map)
	Len() int
}

//// Common structs ////

type mapEntry[K comparable, V any] struct {
	key   K
	value V
	next  *mapEntry[K, V]
	prev  *mapEntry[K, V]
}
