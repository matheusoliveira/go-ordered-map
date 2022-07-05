package omap_test

import (
	"encoding/json"
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/matheusoliveira/go-ordered-map/omap"
)

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
			_, _ = json.Marshal(bm)
		}
	})
	for _, impl := range implementations {
		mymap := impl.initializerStrInt()
		putAllValues(mymap, values)
		b.Run(impl.name, func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				_, _ = json.Marshal(mymap)
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
// builtin map, so it is expected that the difference is minor, and due to random factors.
// LinkedHash is more complex, so it is expected to be slower. All good here.
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
