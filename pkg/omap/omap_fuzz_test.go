package omap_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/matheusoliveira/go-ordered-map/pkg/omap"
)

type operation func(t *testing.T, maps []omap.OMap[string, int], key string, val int)

func opPut(t *testing.T, maps []omap.OMap[string, int], key string, val int) {
	for _, m := range maps {
		m.Put(key, val)
		if r, ok := m.Get(key); !ok {
			t.Fatalf("opPut - failed to get after %T.Put(%q, %d)", m, key, val)
		} else if r != val {
			t.Fatalf("opPut - wrong value %d after %T.Put(%q, %d)", r, m, key, val)
		}
	}
}

func opPutTwice(t *testing.T, maps []omap.OMap[string, int], key string, val int) {
	opPut(t, maps, key, val)
	opPut(t, maps, key, val*2)
}

func opIncrement(t *testing.T, maps []omap.OMap[string, int], key string, val int) {
	validate := func() {
		var lastVal int
		var lastFound bool
		for i, m := range maps {
			val, found := m.Get(key)
			if i == 0 {
				lastVal = val
				lastFound = found
			} else if found != lastFound {
				t.Fatalf("key %q found mismatch at %T: expected %v, found %v", key, m, lastFound, found)
			} else if val != lastVal {
				t.Fatalf("key %q value mismatch at %T: expected %v, found %v", key, m, lastVal, val)
			}
		}
	}
	validate()
	for _, m := range maps {
		if currVal, ok := m.Get(key); ok {
			m.Put(key, currVal+1)
		}
	}
	validate()
}

func opDelete(t *testing.T, maps []omap.OMap[string, int], key string, val int) {
	for _, m := range maps {
		m.Delete(key)
	}
}

func opPutAfterGetAt(t *testing.T, maps []omap.OMap[string, int], key string, val int) {
	putKey := strconv.Itoa(val + 2)
	for _, m := range maps {
		m.PutAfter(m.GetIteratorAt(key), putKey, val)
	}
}

func validateMapsEquality(t *testing.T, maps []omap.OMap[string, int]) bool {
	its := make([]omap.OMapIterator[string, int], len(maps))
	firstLen := maps[0].Len()
	for i, m := range maps {
		its[i] = m.Iterator()
		if m.Len() != firstLen {
			t.Errorf("first len of %d, but found %d at %T", firstLen, m.Len(), m)
		}
	}
	for nIteration := 0; ; nIteration++ {
		allHasNext := true
		for _, it := range its {
			hasNext := it.Next()
			if hasNext && !allHasNext {
				t.Errorf("it %T has next, but a previous one ended", it)
				return false
			}
			allHasNext = hasNext
		}
		if !allHasNext {
			break
		}
		var firstKey string
		var firstVal int
		for i, it := range its {
			key := it.Key()
			val := it.Value()
			if i == 0 {
				firstKey = key
				firstVal = val
			} else if firstKey != key {
				t.Errorf("found mismatching key for iterator %T, at iteration %d, expected %q, found %q", it, nIteration, firstKey, key)
				return false
			} else if firstVal != val {
				t.Errorf("found mismatching value for iterator %T, at iteration %d", it, nIteration)
				return false
			}
		}
	}
	return true
}

func FuzzOMapImpls(f *testing.F) {
	// simplest possible
	f.Add([]byte("1234"), []byte("0123"))
	// some deleting issues found during development
	f.Add([]byte("123411"), []byte("012330"))
	f.Add([]byte("123434"), []byte("000033"))
	// others
	f.Add([]byte("12344"), []byte("01234"))
	// setup
	opMapping := []operation{
		opPut,
		opPutTwice,
		opIncrement,
		opDelete,
		opPutAfterGetAt,
	}
	opDebugMapping := []string{
		"opPut",
		"opPutTwice",
		"opIncrement",
		"opDelete",
		"opPutAfterGetAt",
	}
	f.Fuzz(func(t *testing.T, keyValues []byte, byteOps []byte) {
		if len(keyValues) == 0 || len(byteOps) == 0 {
			return
		}
		// Setup operations
		opsDebug := make([]string, len(byteOps))
		ops := make([]operation, len(byteOps))
		keys := make([]string, len(byteOps))
		vals := make([]int, len(byteOps))
		for i, op := range byteOps {
			opId := int(op) % len(opMapping)
			ops[i] = opMapping[opId]
			opsDebug[i] = opDebugMapping[opId]
			kv := int(keyValues[i%len(keyValues)])
			keys[i] = strconv.Itoa(kv)
			vals[i] = kv
		}
		// Setup maps
		maps := make([]omap.OMap[string, int], 0, len(implementations))
		for _, impl := range implementations {
			if impl.isOrdered {
				maps = append(maps, impl.initializerStrInt())
			}
		}
		// Execute each operation
		for i, op := range ops {
			op(t, maps, keys[i], vals[i])
			opsDebug[i] = fmt.Sprintf("%s(%q,%v)", opsDebug[i], keys[i], vals[i])
		}
			t.Logf("final operations (total of %d): %v", len(ops), opsDebug)
			for _, m := range maps {
				t.Logf("  - map content: %v", m)
			}
		// Iterate over all maps and see if they match perfectly
		if !validateMapsEquality(t, maps) {
			// If failed, debug final result
			t.Logf("final operations (total of %d): %v", len(ops), opsDebug)
			for _, m := range maps {
				t.Logf("  - map content: %v", m)
			}
			// Debug parcial: redo whole test to get first fail (smaller subset)
			opsDebugParcial := make([]string, 0, len(opsDebug))
			maps = maps[0:0] // reset slice
			for _, impl := range implementations {
				if impl.isOrdered {
					maps = append(maps, impl.initializerStrInt())
				}
			}
			for i, op := range ops {
				op(t, maps, keys[i], vals[i])
				opsDebugParcial = append(opsDebugParcial, opsDebug[i])
				if !validateMapsEquality(t, maps) {
					t.Logf("failed at operation %d. Parcial operations: %v", i, opsDebugParcial)
					for _, m := range maps {
						t.Logf("  - map content: %v", m)
					}
				}
			}
			t.Fatal()
		}
	})
}
