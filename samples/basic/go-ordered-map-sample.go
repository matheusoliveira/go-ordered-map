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
