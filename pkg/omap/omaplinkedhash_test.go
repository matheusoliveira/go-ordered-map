package omap

import (
	"encoding/json"
	"math/rand"
	"reflect"
	"strconv"
	"testing"
)

func isHashEffective[K comparable](t *testing.T, nKeys int, keyGen func() K) {
	m := NewOMapLinkedHash[K, any]()
	mh, isMH := m.(*OMapLinkedHash[K, any])
	if !isMH {
		var key K
		t.Errorf("NewOMapLinked of key %v did not return a OMapLinkedHash", reflect.TypeOf(key))
		return
	}
	agg := make(map[uint32]int, nKeys)
	for i := 0; i < nKeys; i++ {
		key := keyGen()
		hash := mh.hasher(&key)
		if v, ok := agg[hash]; ok {
			agg[hash] = v + 1
		} else {
			agg[hash] = 1
		}
		if h, ok := any(key).(HasherTest); ok && !h.called {
			t.Fatalf("not called HashSum32 at position %d", i)
		}
	}
	cnt := 0
	for range agg {
		cnt++
	}
	ratio := float32(cnt) / float32(nKeys)
	if ratio < 0.9 {
		var key K
		t.Errorf("cache efficiency for type %v is too low (< 0.9): %.4f (%d out of %d)", reflect.TypeOf(key), ratio, cnt, nKeys)
	}
}

func randStr() string {
	return strconv.Itoa(rand.Int())
}

type HasherTest struct {
	v      uint32
	called bool
}

func (h *HasherTest) Equal(e HasherTest) bool {
	return h.v == e.v
}

func (h *HasherTest) HashSum32() uint32 {
	h.called = true
	return h.v
}

func TestHasher(t *testing.T) {
	nKeys := 1000
	isHashEffective[*HasherTest](t, nKeys, func() *HasherTest { return &HasherTest{rand.Uint32(), false} })
	isHashEffective[string](t, nKeys, func() string { return randStr() })
	isHashEffective[int](t, nKeys, func() int { return rand.Int() })
	isHashEffective[uint32](t, nKeys, func() uint32 { return uint32(rand.Uint32()) })
	isHashEffective[uint64](t, nKeys, func() uint64 { return rand.Uint64() })
	isHashEffective[int32](t, nKeys, func() int32 { return int32(rand.Int31()) })
	isHashEffective[int64](t, nKeys, func() int64 { return rand.Int63() })
	isHashEffective[float32](t, nKeys, func() float32 { return rand.Float32() * float32(rand.Intn(99999)) })
	isHashEffective[float64](t, nKeys, func() float64 { return rand.Float64() * float64(rand.Intn(99999)) })
}

type Modulo10Hasher struct {
	v uint32
}

func (h Modulo10Hasher) Equal(e HasherTest) bool {
	return h.v == e.v
}

func (h Modulo10Hasher) HashSum32() uint32 {
	return h.v % 10
}

func TestConflictHash(t *testing.T) {
	//
	var i uint32
	m := NewOMapLinkedHash[Modulo10Hasher, uint32]()
	mlh, _ := m.(*OMapLinkedHash[Modulo10Hasher, uint32])
	for i = 0; i < 500; i++ {
		m.Put(Modulo10Hasher{v: i}, i)
	}
	for i = 0; i < 500; i++ {
		if v, ok := m.Get(Modulo10Hasher{v: i}); !ok {
			for _, v := range mlh.m[0] {
				t.Errorf("%v %v", v.key, v.value)
			}
			t.Errorf("key %d not found", v)
			break
		} else if i != v {
			t.Errorf("expected %d, found %d", i, v)
			break
		}
	}
}

func TestMarshalJSON(t *testing.T) {
	type ComplexKey struct {
		Id   int
		Name string
	}
	m := NewOMapLinkedHash[ComplexKey, int]()
	m.Put(ComplexKey{1, "foo"}, 1)
	m.Put(ComplexKey{2, "bar"}, 2)
	m.Put(ComplexKey{3, "baz"}, 3)
	_, err := json.Marshal(m)
	if err == nil {
		t.Errorf("expected an error, found nil")
	}
}
