package omap

import (
	"testing"
)

func TestInvalidJSONAtFirstToken(t *testing.T) {
	m := New[string, int]()
	err := unmarshalJSON(m, []byte("not a valid json"))
	if err == nil {
		t.Fatal("expected an error")
	}
}

func TestInvalidJSONAfterFirstKey(t *testing.T) {
	m := New[string, int]()
	err := unmarshalJSON(m, []byte("{\"foo\": 1, bar}"))
	if err == nil {
		t.Fatal("expected an error")
	}
}

func TestNonStringKeyJSON(t *testing.T) {
	m := New[int, int]()
	err := unmarshalJSON(m, []byte("{\"1\": 2}"))
	if err == nil {
		t.Fatal("expected an error")
	}
}
