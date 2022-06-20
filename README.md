# omap

This package implements map in Golang, very similar to builtin map, but in which the iteration
returns keys in the same ordering as originally inserted.

The time complexity of operations are the same as builtin maps, with very little memory overhead,
and an advantage of faster full-iteration of long maps, see [benchmarks](docs/benchmarks.md) for more
details.

There are multiple [implementations](#implementations) in this library, but basically the
recommendation is to use the initialization functions:
* `omap.New[K, V]()` if you are using a map in non-concurrency scenario (similar to how you'd use
  builtin map)
* `omap.NewOMapSync[K, V]()` if you want to work in the map concurrently in different goroutines
  (this version just wraps the calls to the previous with `sync.RWLock`)

Requirement: Go version >= 1.18, since it needs generics.

## Usage

Installation:

```sh
go get github.com/matheusoliveira/go-ordered-map/
```

Example:

```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/matheusoliveira/go-ordered-map/pkg/omap"
)

func main() {
	m := omap.New[string, int]()
	m.Put("foo", 1)
	m.Put("x", -1)
	m.Put("bar", 2)
	m.Put("y", -1)
	m.Put("baz", 3)
	m.Delete("x")
	m.Delete("y")
	// iterate
	for it := m.Iterator(); it.Next(); {
		fmt.Printf("%s = %d\n", it.Key(), it.Value())
	}
	// JSON, keys are preserved in same order for marshal/unmarshal
	input := []byte(`{"hi":"Hello","name":"World!"}`)
	hello := omap.New[string, string]()
	json.Unmarshal(input, &hello)
	fmt.Println(hello)
	// marshal JSON
	output, _ := json.Marshal(hello)
	fmt.Println(string(output))
	if string(output) == string(input) {
		fmt.Println("Sucess!")
	}
}
```

Output:
```
foo = 1
bar = 2
baz = 3
omap.OMapLinked[hi:Hello name:World!]
{"hi":"Hello","name":"World!"}
Sucess!
```

See [API reference](docs/api.md) for more details.

## Features

- [x] zero external dependencies, use only stdlib packages
- [x] insert, update and delete keys into the map
- [x] implements json.Marshaler to convert map to json, keeping keys order, of course
- [x] implements json.Unmarshaler to convert json to map, keeping keys order, of course
- [x] implements fmt.Stringer to convert map to string, in a similar fashion as builtin map
- [x] support multiple implementations
- [x] Get performance should be very close to builtin map (see [benchmarks](docs/benchmarks.md))
- [x] have builtin sync implementation
- [ ] improve iterator capabilities:
  - [ ] start iterator at specific key
  - [ ] support reverse ordering iterator
  - [ ] support add/remove at iterator position
- Did I miss anything? Add an issue or open a PR and let's discuss

## Implementations

Recommendation:
* If you do not work with the map in multiple goroutines, use the `omap.New` function to create a
  map and let the implementation details aside
* If you want a concurrency-safe implementation, use `omap.NewOMapSync`

If you wish more details, this package is based on interfaces, and provide a set of ready-to-use
implementations, as the following:
* `OMapLinked`: is the default implementation returned by `omap.New`, uses a double-linked list
  with all the elements on the map, each key points to an entry with the key, the value and
  two pointers to prev/next elements, so it can iterate in order.
* `OMapLinkedSync`: uses `OMapLinked` underneath, but providing synchronization with `sync.RWLock`,
  use this only if you need to operate on the map in multiple goroutines at same time (writing once
  and reading many concurrently is safe, only writing is an issue).
* `OMapLinkedHash`: values are stored similar to `OMapLinked`, but key in the map is hashed before
  storing as an attempt to implement a map without the memory overhead to copy key values
  twice (one in the element and another in the mapping), but it has proven not to be very useful in
  most cases, so the recommendation is to use this only, and only if, you have a heavy object being
  used as key and can provide a custom hashing implementation. **Note:** do not use this with string
  key, even if you have large string keys, since Golang internal string is immutable and [use a
  pointer to the actual string](https://cs.opensource.google/go/go/+/refs/tags/go1.18:src/cmd/compile/internal/types/size.go;l=28-33),
  copying it around is not an issue.
* `OMapSimple`: this provides a very simple implementation of `OMap` using an slice to keep the
  order of the keys. It has proven not a very efficient implementation on benchmarks, but as it has
  been kept around as it is a very simple implementation and helps on fuzz and unit testing.
* `OMapBuiltin`: DO NOT USE THIS! It is used internally only to test/compare with builtin map using
  the `OMap` interface, but iterating over this map is a bit expensive (use of goroutine and
  channels) and it is not ordered.

## Benchmarks

See [benchmarks](docs/benchmarks.md) for more details.

