package omap_test

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/matheusoliveira/go-ordered-map/pkg/omap"
)

const (
	implBuiltin    = "Builtin"
	implSimple     = "Simple"
	implLinked     = "Linked"
	implLinkedHash = "LinkedHash"
	implSync       = "Sync"
)

type implDetail struct {
	name                   string
	isOrdered              bool
	isParallelSafe         bool
	initializerStrInt      func() omap.OMap[string, int]
	initializerLargeObjInt func() omap.OMap[LargeObject, int]
}

var implementations []implDetail

func init() {
	omap.EnableOMapBuiltin = true
	implementations = []implDetail{
		{implBuiltin, false, false, omap.NewOMapBuiltin[string, int], omap.NewOMapBuiltin[LargeObject, int]},
		{implSimple, true, false, omap.NewOMapSimple[string, int], omap.NewOMapSimple[LargeObject, int]},
		{implLinked, true, false, omap.NewOMapLinked[string, int], omap.NewOMapLinked[LargeObject, int]},
		{implLinkedHash, true, false, omap.NewOMapLinkedHash[string, int], omap.NewOMapLinkedHash[LargeObject, int]},
		{implSync, true, true, omap.NewOMapSync[string, int], omap.NewOMapSync[LargeObject, int]},
	}
}

func TestBuiltinDeny(t *testing.T) {
	defer func() {
		omap.EnableOMapBuiltin = true
		if r := recover(); r == nil {
			t.Fatal("expected to panic, nil found")
		}
	}()
	omap.EnableOMapBuiltin = false
	_ = omap.NewOMapBuiltin[string, int]()
}

func validateGet(t *testing.T, m omap.OMap[string, int], key string, expected int) bool {
	if v, ok := m.Get(key); !ok {
		t.Errorf("key \"%s\" not found", key)
		return false
	} else if v != expected {
		t.Errorf("expecing value %d for key \"%s\", %d found", expected, key, v)
		return false
	}
	return true
}

type expectedResult[K comparable, V comparable] struct {
	Key   K
	Value V
}

func validateIterator[K comparable, V comparable](t *testing.T, m omap.OMap[K, V], impl implDetail, expResults []expectedResult[K, V]) {
	i := 0
	it := m.Iterator()
	for it.Next() {
		if i >= len(expResults) {
			t.Fatalf("overflow, expecting max of %d value, %d found so far", len(expResults), i)
		}
		key := it.Key()
		value := it.Value()
		exp := expResults[i]
		if !impl.isOrdered {
			found := false
			for _, e := range expResults {
				if e.Key == key {
					found = true
					exp = e
					break
				}
			}
			if !found {
				t.Fatalf("found unexpected key/value pair: %v/%v", key, value)
			}
		}
		if exp.Key != key {
			t.Errorf("expecting key \"%v\" at postion %d, key \"%v\" found", exp.Key, i, key)
		}
		if exp.Value != value {
			t.Errorf("invalid value for key \"%v\", at position %d, expected %v, found %v", key, i, exp.Value, value)
		}
		i++
	}
	if !it.EOF() {
		t.Errorf("EOF returned false instead of true at end of loop")
	}
}

//// Tests ////

func TestNewIsOMapLinked(t *testing.T) {
	exp := reflect.TypeOf(omap.NewOMapLinked[string, int]())
	res := reflect.TypeOf(omap.New[string, int]())
	if res != exp {
		t.Errorf("expected type %v, found %v", exp, res)
	}
}

func TestUnmarshalJSONError(t *testing.T) {
	for _, impl := range implementations {
		t.Run(impl.name, func(t *testing.T) {
			// fail
			invalidJsons := [][]byte{
				[]byte(`"foo"`),         // valid json, but not a map object
				[]byte(`what?`),         // invalid json
				[]byte(`{"ok": what?}`), // valid key, but invalid value
				[]byte(`{"ok": "123"}`), // valid key, but value not int
				[]byte(`{123: 456}`),    // invalid key
			}
			for _, invalidJson := range invalidJsons {
				invalidMap := impl.initializerStrInt()
				if err := json.Unmarshal(invalidJson, &invalidMap); err == nil {
					t.Errorf("expected error on json: %s", string(invalidJson))
				}
			}
		})
	}
}

// Force error to test if MarshalJSON/UnmarshalJSON is treating error correctly
type failonly struct {
	val string
}

func (f *failonly) Equal(e failonly) bool {
	return true
}

func (f failonly) UnmarshalJSON(b []byte) error {
	return fmt.Errorf("can't do that")
}

func (f failonly) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("can't do that")
}

func TestJSONFailurePaths(t *testing.T) {
	for _, impl := range implementations {
		t.Run(impl.name, func(t *testing.T) {
			var mKeyInvalid omap.OMap[failonly, string]
			var mValInvalid omap.OMap[string, failonly]
			switch impl.name {
			case implBuiltin:
				mKeyInvalid = omap.NewOMapBuiltin[failonly, string]()
				mValInvalid = omap.NewOMapBuiltin[string, failonly]()
			case implSimple:
				mKeyInvalid = omap.NewOMapSimple[failonly, string]()
				mValInvalid = omap.NewOMapSimple[string, failonly]()
			case implLinked:
				mKeyInvalid = omap.NewOMapLinked[failonly, string]()
				mValInvalid = omap.NewOMapLinked[string, failonly]()
			case implLinkedHash:
				mKeyInvalid = omap.NewOMapLinkedHash[failonly, string]()
				mValInvalid = omap.NewOMapLinkedHash[string, failonly]()
			case implSync:
				mKeyInvalid = omap.NewOMapSync[failonly, string]()
				mValInvalid = omap.NewOMapSync[string, failonly]()
			default:
				t.Errorf("method not available for Unmarshal: %s", impl.name)
			}
			mKeyInvalid.Put(failonly{"hello"}, "hello")
			mValInvalid.Put("world", failonly{"world"})
			if _, err := json.Marshal(mKeyInvalid); err == nil {
				t.Error("expected error with invalid key")
			}
			if _, err := json.Marshal(mValInvalid); err == nil {
				t.Error("expected error with invalid value")
			}
		})
	}
}

func TestUnmarshalJSON(t *testing.T) {
	for _, impl := range implementations {
		t.Run(impl.name, func(t *testing.T) {
			//lint:ignore U1000 false-positive, reported at https://github.com/dominikh/go-tools/issues/1289
			type person struct {
				Id   int    `json:"id"`
				Name string `json:"name"`
			}
			type parent[M omap.OMap[string, []person]] struct {
				Test string `json:"test"`
				Map  M      `json:"map"`
			}
			data := []byte(`[{"test":"full","map":{"foo":[{"id":1,"name":"foo's name"}],"bar":[],"baz":null}},{"test":"empty","map":{}},{"test":"empty","map":null}]`)

			var errUnmarshal error
			var errMarshal error
			var redec []byte
			switch impl.name {
			case implBuiltin:
				p := make([]parent[*omap.OMapBuiltin[string, []person]], 0)
				errUnmarshal = json.Unmarshal(data, &p)
				redec, errMarshal = json.Marshal(p)
			case implSimple:
				p := make([]parent[*omap.OMapSimple[string, []person]], 0)
				errUnmarshal = json.Unmarshal(data, &p)
				redec, errMarshal = json.Marshal(p)
			case implLinked:
				p := make([]parent[*omap.OMapLinked[string, []person]], 0)
				errUnmarshal = json.Unmarshal(data, &p)
				redec, errMarshal = json.Marshal(p)
			case implLinkedHash:
				p := make([]parent[*omap.OMapLinkedHash[string, []person]], 0)
				errUnmarshal = json.Unmarshal(data, &p)
				redec, errMarshal = json.Marshal(p)
			case implSync:
				p := make([]parent[*omap.OMapSync[string, []person]], 0)
				errUnmarshal = json.Unmarshal(data, &p)
				redec, errMarshal = json.Marshal(p)
			default:
				t.Errorf("method not available for Unmarshal: %s", impl.name)
			}

			if errUnmarshal != nil {
				t.Error(errUnmarshal)
			} else if errMarshal != nil {
				t.Error(errMarshal)
			} else if impl.isOrdered && string(redec) != string(data) {
				t.Errorf("FAIL! returned json is not correct: %s\n", string(redec))
			}
		})
	}
}

func TestDeleteAndMarshalJSON(t *testing.T) {
	for _, impl := range implementations {
		t.Run(impl.name, func(t *testing.T) {
			mymap := impl.initializerStrInt()
			mymap.Put("c", 1)
			mymap.Put("b", 2)
			mymap.Put("a", 3)
			mymap.Delete("b")
			if _, ok := mymap.Get("b"); ok {
				t.Errorf("expected \"b\" to be deleted")
			}
			// check single value
			validateGet(t, mymap, "c", 1)
			// json.Marshal
			data, err := json.Marshal(mymap)
			if err != nil {
				t.Error(err)
			}
			if impl.isOrdered {
				exp := `{"c":1,"a":3}`
				if string(data) != exp {
					t.Errorf("expected %s, found %s", exp, string(data))
				}
			}
			// iterate over all results, see if they are in order
			expResults := []expectedResult[string, int]{
				{"c", 1},
				{"a", 3},
			}
			validateIterator(t, mymap, impl, expResults)
		})
	}
}

func TestOverwriteValue(t *testing.T) {
	for _, impl := range implementations {
		t.Run(impl.name, func(t *testing.T) {
			m := impl.initializerStrInt()
			m.Put("C", 3)
			m.Put("B", 2)
			m.Put("A", 1)
			// overwrite in different order (shouldn't change original order)
			m.Put("A", 10)
			m.Put("B", 20)
			m.Put("C", 30)
			// get
			validateGet(t, m, "A", 10)
			validateGet(t, m, "B", 20)
			validateGet(t, m, "C", 30)
			// iterate
			expResults := []expectedResult[string, int]{
				{"C", 30},
				{"B", 20},
				{"A", 10},
			}
			validateIterator(t, m, impl, expResults)
		})
	}
}

func TestStringer(t *testing.T) {
	for _, impl := range implementations {
		t.Run(impl.name, func(t *testing.T) {
			m := impl.initializerStrInt()
			m.Put("foo", 0)
			m.Put("bar", 2)
			m.Put("x", -1)
			m.Put("baz", 3)
			m.Delete("x")
			m.Put("foo", 1)
			validateIterator(t, m, impl, []expectedResult[string, int]{{"foo", 1}, {"bar", 2}, {"baz", 3}})
			str := fmt.Sprint(m)
			exp := "omap.OMap" + impl.name + "[foo:1 bar:2 baz:3]"
			if impl.isOrdered && str != exp {
				t.Errorf("expected %s, found %s", exp, str)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	for _, impl := range implementations {
		t.Run(impl.name, func(t *testing.T) {
			if !impl.isOrdered {
				return
			}
			m := impl.initializerStrInt()
			m.Put("a", 0)
			m.Put("b", 1)
			m.Put("c", 2)
			m.Put("d", 3)
			m.Put("e", 4)
			validateIterator(t, m, impl, []expectedResult[string, int]{{"a", 0}, {"b", 1}, {"c", 2}, {"d", 3}, {"e", 4}})
			// delete head
			m.Delete("a")
			validateIterator(t, m, impl, []expectedResult[string, int]{{"b", 1}, {"c", 2}, {"d", 3}, {"e", 4}})
			// delete tail
			m.Delete("e")
			validateIterator(t, m, impl, []expectedResult[string, int]{{"b", 1}, {"c", 2}, {"d", 3}})
			// delete in the middle
			m.Delete("c")
			validateIterator(t, m, impl, []expectedResult[string, int]{{"b", 1}, {"d", 3}})
			// empty
			m.Delete("b")
			validateIterator(t, m, impl, []expectedResult[string, int]{{"d", 3}})
			m.Delete("d")
			validateIterator(t, m, impl, []expectedResult[string, int]{})
		})
	}
}

// Will only do an actual test if executed with `-race`
func TestRaceCondition(t *testing.T) {
	const nValues = 10
	for _, impl := range implementations {
		if impl.isParallelSafe {
			var wg sync.WaitGroup
			m := impl.initializerStrInt()
			wg.Add(nValues * 2)
			for i := 0; i < nValues; i++ {
				go func(key string, value int) {
					defer wg.Done()
					time.Sleep(10 * time.Millisecond)
					m.Put(key, value*10)
					m.Put(key+"-to-be-deleted", value)
					time.Sleep(10 * time.Millisecond)
					m.Put(key, value)
					v, ok := m.Get(key)
					if !ok {
						t.Errorf("key %s not found", key)
					}
					if v != value {
						t.Errorf("expected %d at key %s, found %d", value, key, v)
					}
					m.Delete(key + "-to-be-deleted")
				}(strconv.Itoa(i), i)
				go func() {
					defer wg.Done()
					time.Sleep(10 * time.Millisecond)
					for it := m.Iterator(); it.Next(); {
					}
				}()
			}
			wg.Wait()
			cnt := 0
			for it := m.Iterator(); it.Next(); {
				cnt++
				if strconv.Itoa(it.Value()) != it.Key() {
					t.Errorf("expected value %s, found %d", it.Key(), it.Value())
				}
			}
			if cnt != nValues {
				t.Errorf("expected %d values, found %d", cnt, nValues)
			}
		}
	}
}

//// Benchmarks ////

const nValues = 100000
const repeatPut = 2
const strLen = 5000

func genStr(i int) string {
	return strings.Repeat(strconv.Itoa(i), strLen)
}

func putAllValues(m omap.OMap[string, int], values []string) {
	for i, str := range values {
		m.Put(str, i)
	}
}

// Just put many values in the map, outside of benchmark, and then iterate through the map to
// check time taken for full iteration.
// Conclusion: omap implementations using linked list (Linked and LinkedHash) are faster to iterate
// than builtin map, since they have a data struct well design and optimized for that.
func BenchmarkIteration(b *testing.B) {
	values := make([]string, nValues)
	for i := 0; i < nValues; i++ {
		values[i] = strconv.Itoa(i)
	}
	bm := make(map[string]int)
	for i, k := range values {
		bm[k] = i
	}
	b.Run("map", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			for k, v := range bm {
				_, _ = k, v
			}
		}
	})
	for _, impl := range implementations {
		mymap := impl.initializerStrInt()
		putAllValues(mymap, values)
		b.Run(impl.name, func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				for it := mymap.Iterator(); it.Next(); {
					_, _ = it.Key(), it.Value()
				}
			}
		})
	}
}

// Calls json.Marshal to convert a single map of (string, int) to JSON.
// Conclusion: the results here are similar to BenchmarkIteration, since all cases have to iterate
// while building the final JSON.
func BenchmarkMarshalJSON(b *testing.B) {
	values := make([]string, nValues/10)
	for i := 0; i < nValues/10; i++ {
		values[i] = strconv.Itoa(i)
	}
	bm := make(map[string]int)
	for i, k := range values {
		bm[k] = i
	}
	b.Run("map", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			json.Marshal(bm)
		}
	})
	for _, impl := range implementations {
		mymap := impl.initializerStrInt()
		putAllValues(mymap, values)
		b.Run(impl.name, func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				json.Marshal(mymap)
			}
		})
	}
}

// Put values into the map with a short key length, pre-generating the keys before the benchmarks,
// so key size is not accounted in memory.
// Conclusion: since all implementations of omap have to build a separate data structure on Put, it
// is expected that they are slower than builtin map, the trade-off seems acceptable if you you
// need to iterate (or serialize) the map or if have few keys.
func BenchmarkShortStrKeysPut(b *testing.B) {
	values := make([]string, nValues)
	for i := 0; i < nValues; i++ {
		values[i] = strconv.Itoa(i)
	}
	b.Run("map", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			mymap := make(map[string]int)
			for repeat := 0; repeat < repeatPut; repeat++ {
				for i, str := range values {
					mymap[str] = i
				}
			}
		}
	})
	for _, impl := range implementations {
		b.Run(impl.name, func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				mymap := impl.initializerStrInt()
				for repeat := 0; repeat < repeatPut; repeat++ {
					putAllValues(mymap, values)
				}
			}
		})
	}
}

// Put large string keys in map of int value, pre-generating the keys before the benchmarks,
// so key size is not accounted in memory.
// Conclusion: the trade-off here is very similar to BenchmarkShortStrKeysPut, with the advantage
// that using a large key actually improve the relative performance, compared to short key.
func BenchmarkLargeStrKeysPut(b *testing.B) {
	values := make([]string, nValues)
	for i := 0; i < nValues; i++ {
		values[i] = genStr(i)
	}
	b.Run("map", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			mymap := make(map[string]int)
			for repeat := 0; repeat < repeatPut; repeat++ {
				for i, str := range values {
					mymap[str] = i
				}
			}
		}
	})
	for _, impl := range implementations {
		b.Run(impl.name, func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				mymap := impl.initializerStrInt()
				for repeat := 0; repeat < repeatPut; repeat++ {
					putAllValues(mymap, values)
				}
			}
		})
	}
}

// Put large string keys in map of int value, but unlike BenchmarkShortStrKeysPut, this benchmark
// generates the key inside the benchmark, so both key generation time and key memory is accounted
// in the result.
// Conclusion: when the time of large keys generation is accounted in the benchmark, the relative
// performance loss compared to BenchmarkLargeStrKeysPut is actually better.
func BenchmarkLargeStrKeysPutGen(b *testing.B) {
	nValues := nValues / 10
	b.Run("map", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			mymap := make(map[string]int)
			for repeat := 0; repeat < repeatPut; repeat++ {
				for i := 0; i < nValues; i++ {
					mymap[genStr(i)] = i
				}
			}
		}
	})
	for _, impl := range implementations {
		b.Run(impl.name, func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				mymap := impl.initializerStrInt()
				for repeat := 0; repeat < repeatPut; repeat++ {
					for i := 0; i < nValues; i++ {
						mymap.Put(genStr(i), i)
					}
				}
			}
		})
	}
}

// Generate a map of large string keys, same as BenchmarkShortStrKeysPut, and then run the
// benchmark only to get the values of a random key. All sub-benchmarks use same random
// seed.
// Conclusion: except for LinkedHash, the implementations basically map the Get operation to a
// builtin map, so it is expected that the difference is minor. LinkedHash is more complex, so
// it is expected to be slower. All good here.
func BenchmarkLargeStrKeysGet(b *testing.B) {
	seed := time.Now().UnixNano()
	values := make([]string, nValues)
	for i := 0; i < nValues; i++ {
		values[i] = genStr(i)
	}
	bmap := make(map[string]int)
	for i, str := range values {
		bmap[str] = i
	}
	b.Run("map", func(b *testing.B) {
		rnd := rand.New(rand.NewSource(seed))
		for n := 0; n < b.N; n++ {
			_ = bmap[values[rnd.Intn(nValues)]]
		}
	})
	for _, impl := range implementations {
		mymap := impl.initializerStrInt()

		putAllValues(mymap, values)
		b.Run(impl.name, func(b *testing.B) {
			rnd := rand.New(rand.NewSource(seed))
			for n := 0; n < b.N; n++ {
				_, _ = mymap.Get(values[rnd.Intn(nValues)])
			}
		})
	}
}

// Generate a map of large string keys, same as BenchmarkShortStrKeysPut, and then run the
// benchmark to iterate over all key/value pairs.
// Conclusion: the performance iteration with large keys is even better than short keys.
func BenchmarkLargeStrKeysIterate(b *testing.B) {
	values := make([]string, nValues)
	for i := 0; i < nValues; i++ {
		values[i] = genStr(i)
	}
	for _, impl := range implementations {
		mymap := impl.initializerStrInt()
		putAllValues(mymap, values)
		b.Run(impl.name, func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				for it := mymap.Iterator(); it.Next(); {
				}
			}
		})
	}
}

// Generate a map of large strings keys and int value, and get all values one by one.
func BenchmarkLargeStrKeysPutGet(b *testing.B) {
	values := make([]string, nValues)
	for i := 0; i < nValues; i++ {
		values[i] = genStr(i)
	}
	b.Run("map", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			mymap := make(map[string]int)
			for i, str := range values {
				mymap[str] = i
			}
			cnt := 0
			for i, str := range values {
				cnt++
				val, ok := mymap[str]
				if !ok {
					b.Fatalf("Key of %d not found!\n", i)
					break
				}
				if val != i {
					b.Fatalf("Invalid value: expected %d, got %d\n", i, val)
					break
				}
			}
			if cnt != len(values) {
				b.Fatalf("Iteration failed, expected to found %d records, found %d\n", len(values), cnt)
			}
		}
	})
	for _, impl := range implementations {
		b.Run(impl.name, func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				mymap := impl.initializerStrInt()
				putAllValues(mymap, values)
				cnt := 0
				for i, str := range values {
					cnt++
					mymap.Get(str)
					val, ok := mymap.Get(str)
					if !ok {
						b.Fatalf("Key of %d not found!\n", i)
						break
					}
					if val != i {
						b.Fatalf("Invalid value: expected %d, got %d\n", i, val)
						break
					}
				}
				if cnt != len(values) {
					b.Fatalf("Iteration failed, expected to found %d records, found %d\n", len(values), cnt)
				}
			}
		})
	}
}

// Benchmark of a large struct as key
type LargeObject struct {
	Id      uint32
	BigAttr [32768]byte
}

type LargeObjectHash struct {
	LargeObject
}

func (h LargeObject) Equal(e LargeObject) bool {
	return h.Id == e.Id
}

func (h LargeObject) HashSum32() uint32 {
	return h.Id
}

func (h LargeObjectHash) HashSum32() uint32 {
	return h.Id
}

// Generate a map of large strings keys and int value, and get all values one by one.
// Conclusion: this test is designed specifically for LinkedHash implementation, and is actually
// the only use-case where this implementation is a good fit, and as expected it is the fastest
// of omap implementations, although still slower than builtin map. Albeit, it seems a very
// specific and unusual use case.
func BenchmarkLargeObjectKey(b *testing.B) {
	const nValues = nValues / 10
	b.Run("map", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			mymap := make(map[LargeObject]int)
			for i := 0; i < nValues; i++ {
				obj := LargeObject{Id: uint32(i)}
				mymap[obj] = i
			}
			cnt := 0
			for k, v := range mymap {
				cnt++
				if k.Id != uint32(v) {
					b.Fatalf("key of id %d, but value %d\n", k.Id, v)
				}
			}
			if cnt != nValues {
				b.Fatalf("Iteration failed, expected to found %d records, found %d\n", nValues, cnt)
			}
		}
	})
	for _, impl := range implementations {
		b.Run(impl.name, func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				mymap := impl.initializerLargeObjInt()
				for i := 0; i < nValues; i++ {
					obj := LargeObject{Id: uint32(i)}
					mymap.Put(obj, i)
				}
				cnt := 0
				for it := mymap.Iterator(); it.Next(); {
					cnt++
					if it.Key().Id != uint32(it.Value()) {
						b.Fatalf("key of id %d, but value %d\n", it.Key().Id, it.Value())
					}
				}
				if cnt != nValues {
					b.Fatalf("Iteration failed, expected to found %d records, found %d\n", nValues, cnt)
				}
				if cnt != nValues {
					b.Fatalf("Iteration failed, expected to found %d records, found %d\n", nValues, cnt)
				}
			}
		})
	}
}
