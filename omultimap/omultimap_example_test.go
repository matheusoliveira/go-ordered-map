package omultimap_test

import (
	"encoding/json"
	"fmt"

	"github.com/matheusoliveira/go-ordered-map/omultimap"
	"github.com/matheusoliveira/go-ordered-map/omap"
)

func Example() {
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
	err := json.Unmarshal(js, mm2)
	fmt.Println(mm2, err)

	// Output:
	// iterate all keys/values:
	//   "foo" / 1
	//   "bar" / 2
	//   "baz" / 3
	//   "foo" / 4
	//   "bar" / 5
	//   "baz" / 6
	//   "foo" / 7
	//   "bar" / 8
	//   "baz" / 9
	// values of foo: [1 4 7]
	// marshal output:
	//   {"foo":1,"bar":2,"baz":3,"foo":4,"bar":5,"baz":6,"foo":7,"bar":8,"baz":9}
	// unmarshal output:
	//   omultimap.OMultiMapLinked[foo:1 bar:2 baz:3 foo:4 bar:5 baz:6 foo:7 bar:8 baz:9] <nil>
}

