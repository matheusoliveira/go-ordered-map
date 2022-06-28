package omap_test

import (
	"encoding/json"
	"fmt"

	"github.com/matheusoliveira/go-ordered-map/pkg/omap"
)

func Example() {
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
	err := json.Unmarshal(input, &hello)
	fmt.Println(hello, err)
	// marshal JSON
	output, _ := json.Marshal(hello)
	fmt.Println(string(output))
	if string(output) == string(input) {
		fmt.Println("Sucess!")
	}
	// reverse iterator
	for it := m.Iterator().MoveBack(); it.Prev(); {
		fmt.Printf("%s = %d\n", it.Key(), it.Value())
	}

	// Output:
	// foo = 1
	// bar = 2
	// baz = 3
	// omap.OMapLinked[hi:Hello name:World!] <nil>
	// {"hi":"Hello","name":"World!"}
	// Sucess!
	// baz = 3
	// bar = 2
	// foo = 1
}
