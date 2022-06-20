[![Build status](https://github.com/matheusoliveira/go-ordered-map/actions/workflows/build.yml/badge.svg)](https://github.com/matheusoliveira/go-ordered-map/actions/workflows/build.yml)
[![Codecov](https://codecov.io/gh/matheusoliveira/go-ordered-map/branch/main/graph/badge.svg?token=H4SjidS9Yq)](https://codecov.io/gh/matheusoliveira/go-ordered-map)
[![Go Report Card](https://goreportcard.com/badge/github.com/matheusoliveira/go-ordered-map)](https://goreportcard.com/report/github.com/matheusoliveira/go-ordered-map)
[![Total alerts](https://img.shields.io/lgtm/alerts/g/matheusoliveira/go-ordered-map.svg?logo=lgtm&logoWidth=18)](https://lgtm.com/projects/g/matheusoliveira/go-ordered-map/alerts/)
[![Go Reference](https://pkg.go.dev/badge/github.com/matheusoliveira/go-ordered-map@main/pkg/omap.svg)](https://pkg.go.dev/github.com/matheusoliveira/go-ordered-map@main/pkg/omap)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](https://opensource.org/licenses/MIT)

Collection of ordered map and ordered multimap implementations:
- [omap.OMapLinked](https://pkg.go.dev/github.com/matheusoliveira/go-ordered-map@main/pkg/omap#OMapLinked)
  implements an ordered map using a double-linked list to keep the ordering. It is the default and
  recommended implementation of `OMap`.
- [omap.OMapSync](https://pkg.go.dev/github.com/matheusoliveira/go-ordered-map@main/pkg/omap#OMapSync)
  implements an ordered map using OMapLinked underneath and providing synchronization to be
  parallel-safe.  Use this only if you need to operate on the map in multiple goroutines at same
  time (writing once and reading many concurrently is safe, only writing in parallel with other
  reads/writes is an issue).
- [omultimap.OMultiMapLinked](https://pkg.go.dev/github.com/matheusoliveira/go-ordered-map@main/pkg/omultimap#OMultiMapLinked)
  implements an ordered multimap that can hold many values per key, and still keep then in order
  using a linked list internally
- [omultimap.OMultiMapSync](https://pkg.go.dev/github.com/matheusoliveira/go-ordered-map@main/pkg/omultimap#OMultiMapSync)
  implements an ordered multimap using OMultiMapLinked underneath and providing synchronization to be
  parallel-safe

Implementation not recommended, in general (use only if you prove it better):
- [omap.OMapLinkedHash](https://pkg.go.dev/github.com/matheusoliveira/go-ordered-map@main/pkg/omap#OMapLinkedHash)
  implements an ordered map similar to OMapLinked, but hashes the value before saving, it is
  adivised to use only in very few occasions, see [benchmarks](docs/benchmarks.md) for details.
  This implementation is an attempt to have a map without the memory overhead to copy keys twice
  (one in the element and another in the mapping), but it has proven not to be very useful in
  most cases, so the recommendation is to use this only, and only if, you have a heavy object being
  used as key and can provide a custom hashing implementation. **Note:** do not use this with string
  key, even if you have large string keys, since Golang internal string is immutable and [use a
  pointer to the actual string](https://cs.opensource.google/go/go/+/refs/tags/go1.18:src/cmd/compile/internal/types/size.go;l=28-33),
  copying it around is not an issue for `OMapLinked`.

Implementations used only for testing (do not use this in production code):
- [omap.OMapSimple](https://pkg.go.dev/github.com/matheusoliveira/go-ordered-map@main/pkg/omap#OMapSimple)
  provides a very simple implementation of `OMap` using an slice to keep the
  order of the keys. It has proven not a very efficient implementation on benchmarks, but as it has
  been kept around as it is a very simple implementation and helps on fuzz and unit testing.
- [omap.OMapBuiltin](https://pkg.go.dev/github.com/matheusoliveira/go-ordered-map@main/pkg/omap#OMapBuiltin)
  DO NOT USE THIS! It is used internally only to test/compare with builtin map using
  the `OMap` interface, but iterating over this map is a bit expensive (use of goroutine and
  channels, which can leak) and it is not ordered.

Requirements:
- Go version >= 1.18, since it needs generics.
- That is it, no external dependency.

## Usage

Installation:

```sh
go get -u github.com/matheusoliveira/go-ordered-map/
```

Then simple import `omap` or `omultimap` and use `New*` functions. Import paths:
- `"github.com/matheusoliveira/go-ordered-map/pkg/omap"`
- `"github.com/matheusoliveira/go-ordered-map/pkg/omultimap"`

# omap

This package implements map in Golang, very similar to builtin map, but in which the iteration
returns keys in the same ordering as originally inserted.

The time complexity of operations are the same as builtin maps, with very little memory overhead,
and an advantage of faster full-iteration of long maps, see [benchmarks](docs/benchmarks.md) for more
details.

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

See [omap API reference](https://pkg.go.dev/github.com/matheusoliveira/go-ordered-map@main/pkg/omap) for more details.

# omultimap

It is an ordered multimap, each key can hold multiple values and iterating over the full map will return in
the same ordering as originally inserted. It is handy for cases when you want the multiple values
for the same key or if you want to marshal/unmarshal and keep the original state.

Example:

```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/matheusoliveira/go-ordered-map/pkg/omultimap"
	"github.com/matheusoliveira/go-ordered-map/pkg/omap"
)

func main() {
	mm := omultimap.New[string, int]()
	mm.Put("foo", 1)
	mm.Put("bar", 2)
	mm.Put("baz", 3)
	mm.Put("foo", 4)
	mm.Put("bar", 5)
	mm.Put("baz", 6)
	mm.Put("foo", 7)
	mm.Put("bar", 8)
	mm.Put("baz", 9)
	fmt.Println("iterate all keys/values:")
	for it := mm.Iterator(); it.Next(); {
		fmt.Printf("  %q / %d\n", it.Key(), it.Value())
	}
	// iterate values of a given key
	fmt.Printf("values of foo: ")
	fooIt := mm.GetValuesOf("foo")
	fmt.Println(omap.IteratorValuesToSlice(fooIt))
	// marshal JSON
	fmt.Printf("marshal output:\n  ")
	js, _ := json.Marshal(mm)
	fmt.Println(string(js))
	// unmarshal JSON
	fmt.Printf("unmarshal output:\n  ")
	mm2 := omultimap.New[string, int]()
	json.Unmarshal(js, mm2)
	fmt.Println(mm2)
}
```

Output:
```
iterate all keys/values:
  "foo" / 1
  "bar" / 2
  "baz" / 3
  "foo" / 4
  "bar" / 5
  "baz" / 6
  "foo" / 7
  "bar" / 8
  "baz" / 9
values of foo: [1 4 7]
marshal output:
  {"foo":1,"bar":2,"baz":3,"foo":4,"bar":5,"baz":6,"foo":7,"bar":8,"baz":9}
unmarshal output:
  omultimap.OMultiMapLinked[foo:1 bar:2 baz:3 foo:4 bar:5 baz:6 foo:7 bar:8 baz:9]
```

See [omultimap API reference](https://pkg.go.dev/github.com/matheusoliveira/go-ordered-map@main/pkg/omultimap) for more details.

# Features

- [x] zero external dependencies, use only stdlib packages
- [x] insert, update and delete keys into the map
- [x] implements json.Marshaler to convert map to json, keeping keys order, of course
- [x] implements json.Unmarshaler to convert json to map, keeping keys order, of course
- [x] implements fmt.Stringer to convert map to string, in a similar fashion as builtin map
- [x] support multiple implementations
- [x] Get performance should be very close to builtin map (see [benchmarks](docs/benchmarks.md))
- [x] have builtin sync implementation
- [x] implements multimap (`omultimap`)
- [ ] improve iterator capabilities:
  - [ ] start iterator at specific key
  - [ ] support reverse ordering iterator
  - [ ] support add/remove at iterator position

Did I miss anything? Create an [issue](https://github.com/matheusoliveira/go-ordered-map/issues) or open a [pull request](https://github.com/matheusoliveira/go-ordered-map/pulls) and let's discuss

# Benchmarks

See [benchmarks](docs/benchmarks.md) for more details.
