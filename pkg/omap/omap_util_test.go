package omap

import (
	"testing"
)

func TestInvalidJSONAtFirstToken(t *testing.T) {
	m := New[string, int]()
	err := UnmarshalJSON(m.Put, []byte("not a valid json"))
	if err == nil {
		t.Fatal("expected an error")
	}
}

func TestInvalidJSONAfterFirstKey(t *testing.T) {
	m := New[string, int]()
	err := UnmarshalJSON(m.Put, []byte("{\"foo\": 1, bar}"))
	if err == nil {
		t.Fatal("expected an error")
	}
}

func TestNonStringKeyJSON(t *testing.T) {
	m := New[int, int]()
	err := UnmarshalJSON(m.Put, []byte("{\"1\": 2}"))
	if err == nil {
		t.Fatal("expected an error")
	}
}
