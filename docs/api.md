# omap

```go
import "github.com/matheusoliveira/go-ordered-map/pkg/omap"
```


## Usage

```go
var EnableOMapBuiltin = false
```
This is a safe var, since OMapBuiltin should be used only for testings, since it
is not actually an ordered map, one must explicitly set this variable to `true`
before using it, or it will panic on initialization.

#### type Hasher

```go
type Hasher interface {
	HashSum32() uint32
}
```

Objects that want to create a custom hashing for the key used by OMapLinkedHash
must implement this interface, giving a HashSum32 func that returns the hash of
the object as an uint32.

#### type OMap

```go
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
```

OMap is an ordered map that holds key/value and is able to iterate over the
whole data-set in the same order as insertion has happened.

#### func  New

```go
func New[K comparable, V any]() OMap[K, V]
```
Create a new map using the default implementation, which is considered the best
trade-off among all. Currently, OMapLinked is the winner.

#### func  NewOMapBuiltin

```go
func NewOMapBuiltin[K comparable, V any]() OMap[K, V]
```

#### func  NewOMapLinked

```go
func NewOMapLinked[K comparable, V any]() OMap[K, V]
```
Return a new OMap based on OMapLinked implementation, see OMapLinked type for
more details of the implementation.

#### func  NewOMapLinkedHash

```go
func NewOMapLinkedHash[K comparable, V any]() OMap[K, V]
```
Return a new OMap based on OMapLinkedHash implementation, see OMapLinkedHash
type for more details of the implementation.

#### func  NewOMapSimple

```go
func NewOMapSimple[K comparable, V any]() OMap[K, V]
```
Create a new OMap instance using OMapSimple implementation.

#### func  NewOMapSync

```go
func NewOMapSync[K comparable, V any]() OMap[K, V]
```
Create a new OMap instance using OMapSync implementation.

#### type OMapBuiltin

```go
type OMapBuiltin[K comparable, V any] struct {
}
```

DO NOT USE THIS FOR REAL!!! Implements OMap interface but not very strictly,
should be use only for comparison with builtin map

#### func (*OMapBuiltin[K, V]) Delete

```go
func (m *OMapBuiltin[K, V]) Delete(key K)
```

#### func (*OMapBuiltin[K, V]) Get

```go
func (m *OMapBuiltin[K, V]) Get(key K) (V, bool)
```

#### func (*OMapBuiltin[K, V]) Iterator

```go
func (m *OMapBuiltin[K, V]) Iterator() OMapIterator[K, V]
```

#### func (OMapBuiltin[K, V]) MarshalJSON

```go
func (m OMapBuiltin[K, V]) MarshalJSON() ([]byte, error)
```
Implement json.Marshaler interface.

#### func (*OMapBuiltin[K, V]) Put

```go
func (m *OMapBuiltin[K, V]) Put(key K, value V)
```

#### func (*OMapBuiltin[K, V]) String

```go
func (m *OMapBuiltin[K, V]) String() string
```
Implement fmt.Stringer

#### func (OMapBuiltin[K, V]) UnmarshalJSON

```go
func (m OMapBuiltin[K, V]) UnmarshalJSON(b []byte) error
```
Implement json.Unmarshaler interface.

#### type OMapBuiltinData

```go
type OMapBuiltinData[K comparable, V any] struct {
}
```


#### type OMapBuiltinIterator

```go
type OMapBuiltinIterator[K comparable, V any] struct {
}
```

Iterator over a OMapSimple, should be created through OMapSimple.Iterator()
function.

#### func (*OMapBuiltinIterator[K, V]) EOF

```go
func (it *OMapBuiltinIterator[K, V]) EOF() bool
```

#### func (*OMapBuiltinIterator[K, V]) Key

```go
func (it *OMapBuiltinIterator[K, V]) Key() K
```

#### func (*OMapBuiltinIterator[K, V]) Next

```go
func (it *OMapBuiltinIterator[K, V]) Next() bool
```

#### func (*OMapBuiltinIterator[K, V]) Value

```go
func (it *OMapBuiltinIterator[K, V]) Value() V
```

#### type OMapIterator

```go
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
```

Iterator over OMap.

#### type OMapLinked

```go
type OMapLinked[K comparable, V any] struct {
}
```

Implements an ordered map using double-linked list for iteration.

#### func (*OMapLinked[K, V]) Delete

```go
func (m *OMapLinked[K, V]) Delete(key K)
```

#### func (*OMapLinked[K, V]) Get

```go
func (m *OMapLinked[K, V]) Get(key K) (V, bool)
```

#### func (*OMapLinked[K, V]) Iterator

```go
func (m *OMapLinked[K, V]) Iterator() OMapIterator[K, V]
```

#### func (OMapLinked[K, V]) MarshalJSON

```go
func (m OMapLinked[K, V]) MarshalJSON() ([]byte, error)
```
Implement json.Marshaler interface.

#### func (*OMapLinked[K, V]) Put

```go
func (m *OMapLinked[K, V]) Put(key K, value V)
```

#### func (*OMapLinked[K, V]) String

```go
func (m *OMapLinked[K, V]) String() string
```
Implement fmt.Stringer

#### func (*OMapLinked[K, V]) UnmarshalJSON

```go
func (m *OMapLinked[K, V]) UnmarshalJSON(b []byte) error
```
Implement json.Unmarshaler interface.

#### type OMapLinkedHash

```go
type OMapLinkedHash[K comparable, V any] struct {
}
```

Implement an ordered map using a linked list but saving the key as an uint32
hash instead of copying the key into the map. This implementation should only be
used when you have a very large object struct as K key, and preferable this
object should implement Hasher interface to provide a performant hashing
algorithm for the type.

#### func (*OMapLinkedHash[K, V]) Delete

```go
func (m *OMapLinkedHash[K, V]) Delete(key K)
```

#### func (*OMapLinkedHash[K, V]) Get

```go
func (m *OMapLinkedHash[K, V]) Get(key K) (V, bool)
```

#### func (*OMapLinkedHash[K, V]) Iterator

```go
func (m *OMapLinkedHash[K, V]) Iterator() OMapIterator[K, V]
```

#### func (OMapLinkedHash[K, V]) MarshalJSON

```go
func (m OMapLinkedHash[K, V]) MarshalJSON() ([]byte, error)
```
Implement json.Marshaler interface.

#### func (*OMapLinkedHash[K, V]) Put

```go
func (m *OMapLinkedHash[K, V]) Put(key K, value V)
```

#### func (*OMapLinkedHash[K, V]) String

```go
func (m *OMapLinkedHash[K, V]) String() string
```
Implement fmt.Stringer

#### func (*OMapLinkedHash[K, V]) UnmarshalJSON

```go
func (m *OMapLinkedHash[K, V]) UnmarshalJSON(b []byte) error
```
Implement json.Unmarshaler interface.

#### type OMapLinkedHashIterator

```go
type OMapLinkedHashIterator[K comparable, V any] struct {
}
```

Implement OMapIterator for OMapLinkedHash

#### func (*OMapLinkedHashIterator[K, V]) EOF

```go
func (it *OMapLinkedHashIterator[K, V]) EOF() bool
```

#### func (*OMapLinkedHashIterator[K, V]) Key

```go
func (it *OMapLinkedHashIterator[K, V]) Key() K
```

#### func (*OMapLinkedHashIterator[K, V]) Next

```go
func (it *OMapLinkedHashIterator[K, V]) Next() bool
```

#### func (*OMapLinkedHashIterator[K, V]) Value

```go
func (it *OMapLinkedHashIterator[K, V]) Value() V
```

#### type OMapLinkedIterator

```go
type OMapLinkedIterator[K comparable, V any] struct {
}
```

Implements OMapIterator for OMapLinked.

#### func (*OMapLinkedIterator[K, V]) EOF

```go
func (it *OMapLinkedIterator[K, V]) EOF() bool
```

#### func (*OMapLinkedIterator[K, V]) Key

```go
func (it *OMapLinkedIterator[K, V]) Key() K
```

#### func (*OMapLinkedIterator[K, V]) Next

```go
func (it *OMapLinkedIterator[K, V]) Next() bool
```

#### func (*OMapLinkedIterator[K, V]) Value

```go
func (it *OMapLinkedIterator[K, V]) Value() V
```

#### type OMapSimple

```go
type OMapSimple[K comparable, V any] struct {
}
```

Implements a OMap interface using a very simple algorithm: it basically keeps a
map[K]V to hold the mappings, and a []K slice to keep the order (hence doubling
the memory used to store the keys, compared to a simple Go map).

#### func (*OMapSimple[K, V]) Delete

```go
func (m *OMapSimple[K, V]) Delete(key K)
```
Delete the value pointing to the given key. Complexity: O(n)

#### func (*OMapSimple[K, V]) Get

```go
func (m *OMapSimple[K, V]) Get(key K) (V, bool)
```
Get the value pointing to the given key, returning true as second argument if
found, and false otherwise. Complexity: O(1), same as builtin map[key]

#### func (*OMapSimple[K, V]) Iterator

```go
func (m *OMapSimple[K, V]) Iterator() OMapIterator[K, V]
```
Return an iterator to navigate the map.

#### func (OMapSimple[K, V]) MarshalJSON

```go
func (m OMapSimple[K, V]) MarshalJSON() ([]byte, error)
```
Implement json.Marshaler interface.

#### func (*OMapSimple[K, V]) Put

```go
func (m *OMapSimple[K, V]) Put(key K, value V)
```
Add/overwrite the value in the map on the given key. Important to note that if a
key existed and is being overwritten, the order of the old key insertion
position will remain when iterating the map. Complexity: O(1)

#### func (*OMapSimple[K, V]) String

```go
func (m *OMapSimple[K, V]) String() string
```
Implement fmt.Stringer

#### func (*OMapSimple[K, V]) UnmarshalJSON

```go
func (m *OMapSimple[K, V]) UnmarshalJSON(b []byte) error
```
Implement json.Unmarshaler interface.

#### type OMapSimpleIterator

```go
type OMapSimpleIterator[K comparable, V any] struct {
}
```

Iterator over a OMapSimple, should be created through OMapSimple.Iterator()
function.

#### func (OMapSimpleIterator[K, V]) EOF

```go
func (it OMapSimpleIterator[K, V]) EOF() bool
```
Returns true if iterator has reached the end

#### func (OMapSimpleIterator[K, V]) Key

```go
func (it OMapSimpleIterator[K, V]) Key() K
```
Return the key at current record. Calling this function when EOF() is true will
cause a panic.

#### func (*OMapSimpleIterator[K, V]) Next

```go
func (it *OMapSimpleIterator[K, V]) Next() bool
```
Move iterator to the next record, returning true if there is a next value and
false otherwise. Complexity: in general should be O(1), but it needs to skip
deleted keys, so if there M deleted keys on the current position, it will be
O(M). It is a trade-off to avoid making Delete O(N).

#### func (OMapSimpleIterator[K, V]) Value

```go
func (it OMapSimpleIterator[K, V]) Value() V
```
Return the value at current record. Calling this function when EOF() is true
will cause a panic.

#### type OMapSync

```go
type OMapSync[K comparable, V any] struct {
}
```

Implements a OMap interface using a very simple algorithm: it basically keeps a
map[K]V to hold the mappings, and a []K slice to keep the order (hence doubling
the memory used to store the keys, compared to a simple Go map).

#### func (*OMapSync[K, V]) Delete

```go
func (m *OMapSync[K, V]) Delete(key K)
```
Delete the value pointing to the given key. Complexity: same as builtin
[delete](https://pkg.go.dev/builtin#delete)

#### func (*OMapSync[K, V]) Get

```go
func (m *OMapSync[K, V]) Get(key K) (V, bool)
```
Get the value pointing to the given key, returning true as second argument if
found, and false otherwise. Complexity: O(1), same as builtin map[key]

#### func (*OMapSync[K, V]) Iterator

```go
func (m *OMapSync[K, V]) Iterator() OMapIterator[K, V]
```
Return an iterator to navigate the map.

#### func (*OMapSync[K, V]) MarshalJSON

```go
func (m *OMapSync[K, V]) MarshalJSON() ([]byte, error)
```
Implement json.Marshaler interface.

#### func (*OMapSync[K, V]) Put

```go
func (m *OMapSync[K, V]) Put(key K, value V)
```
Add/overwrite the value in the map on the given key. Important to note that if a
key existed and is being overwritten, the order of the old key insertion
position will remain when iterating the map. Complexity: O(1)

#### func (*OMapSync[K, V]) String

```go
func (m *OMapSync[K, V]) String() string
```
Implement fmt.Stringer interface.

#### func (*OMapSync[K, V]) UnmarshalJSON

```go
func (m *OMapSync[K, V]) UnmarshalJSON(b []byte) error
```
Implement json.Unmarshaler interface.

#### type OMapSyncIterator

```go
type OMapSyncIterator[K comparable, V any] struct {
}
```

Iterator over a OMapSync, should be created through OMapSync.Iterator()
function.

#### func (OMapSyncIterator[K, V]) EOF

```go
func (it OMapSyncIterator[K, V]) EOF() bool
```
Returns true if iterator has reached the end

#### func (OMapSyncIterator[K, V]) Key

```go
func (it OMapSyncIterator[K, V]) Key() K
```
Return the key at current record. Calling this function when EOF() is true will
cause a panic.

#### func (*OMapSyncIterator[K, V]) Next

```go
func (it *OMapSyncIterator[K, V]) Next() bool
```
Move iterator to the next record, returning true if there is a next value and
false otherwise. Complexity: in general should be O(1), but it needs to skip
deleted keys, so if there M deleted keys on the current position, it will be
O(M). It is a trade-off to avoid making Delete O(N).

#### func (OMapSyncIterator[K, V]) Value

```go
func (it OMapSyncIterator[K, V]) Value() V
```
Return the value at current record. Calling this function when EOF() is true
will cause a panic.
