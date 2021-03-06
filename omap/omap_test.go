package omap_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	th "github.com/matheusoliveira/go-ordered-map/internal/testhelper"
	"github.com/matheusoliveira/go-ordered-map/omap"
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
	// TODO: we wrap New* calls into anonym function due to a golangci-lint bug, we can simple use
	//       the function after golangci-lint solves the bug, see PR #8 for more details.
	//       Issue: https://github.com/golangci/golangci-lint/issues/2859
	omap.EnableOMapBuiltin = true
	implementations = []implDetail{
		{
			implBuiltin,
			false,
			false,
			func() omap.OMap[string, int] { return omap.NewOMapBuiltin[string, int]() },
			func() omap.OMap[LargeObject, int] { return omap.NewOMapBuiltin[LargeObject, int]() },
		},
		{
			implSimple,
			true,
			false,
			func() omap.OMap[string, int] { return omap.NewOMapSimple[string, int]() },
			func() omap.OMap[LargeObject, int] { return omap.NewOMapSimple[LargeObject, int]() },
		},
		{
			implLinked,
			true,
			false,
			func() omap.OMap[string, int] { return omap.NewOMapLinked[string, int]() },
			func() omap.OMap[LargeObject, int] { return omap.NewOMapLinked[LargeObject, int]() },
		},
		{
			implLinkedHash,
			true,
			false,
			func() omap.OMap[string, int] { return omap.NewOMapLinkedHash[string, int]() },
			func() omap.OMap[LargeObject, int] { return omap.NewOMapLinkedHash[LargeObject, int]() },
		},
		{
			implSync,
			true,
			true,
			func() omap.OMap[string, int] { return omap.NewOMapSync[string, int]() },
			func() omap.OMap[LargeObject, int] { return omap.NewOMapSync[LargeObject, int]() },
		},
	}
}

func validateGet(t *testing.T, m omap.OMap[string, int], key string, expected int) bool {
	t.Helper()
	if v, ok := m.Get(key); !ok {
		t.Errorf("key \"%s\" not found", key)
		return false
	} else if v != expected {
		t.Errorf("expecing value %d for key \"%s\", %d found", expected, key, v)
		return false
	}
	return true
}

func TestIteratorToSlice(t *testing.T) {
	for _, impl := range implementations {
		t.Run(impl.name, func(t *testing.T) {
			m := impl.initializerStrInt()
			m.Put("foo", 10)
			m.Put("foo", 11)
			m.Put("foo", 12)
			m.Put("bar", 20)
			m.Put("bar", 21)
			m.Put("bar", 22)
			m.Put("baz", 30)
			m.Put("baz", 31)
			m.Put("baz", 32)
			if m.Len() != 3 {
				t.Errorf("expected %T.Len() of 3, found %d", m, m.Len())
			}
			itKeys := m.Iterator()
			itVals := m.Iterator()
			keys := omap.IteratorKeysToSlice(itKeys)
			vals := omap.IteratorValuesToSlice(itVals)
			if len(keys) != 3 {
				t.Errorf("expected keys size of 3, found %d", len(keys))
			} else if impl.isOrdered && !(keys[0] == "foo" && keys[1] == "bar" && keys[2] == "baz") {
				t.Errorf("invalid keys, expected [foo bar baz], found %v", keys)
			}
			if len(vals) != 3 {
				t.Errorf("expected vals size of 3, found %d", len(vals))
			} else if impl.isOrdered && !(vals[0] == 12 && vals[1] == 22 && vals[2] == 32) {
				t.Errorf("invalid vals, expected [12 22 32], found %v", vals)
			}
			if !itKeys.EOF() {
				t.Error("expected itKeys to be at EOF")
			}
			if !itVals.EOF() {
				t.Error("expected itVals to be at EOF")
			}
		})
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
				th.AssertErrNotNil(t, json.Unmarshal(invalidJson, &invalidMap), "expecting an error on json.Unmarshal")
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
			th.ValidateIterator(t, mymap.Iterator(), impl.isOrdered, th.JsonToKV[string, int](`[["c",1],["a",3]]`))
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
			th.ValidateIterator(t, m.Iterator(), impl.isOrdered, th.JsonToKV[string, int](`[["C",30],["B",20],["A",10]]`))
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
			th.ValidateIterator(t, m.Iterator(), impl.isOrdered, th.JsonToKV[string, int](`[["foo",1],["bar",2],["baz",3]]`))
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
			th.ValidateIterator(t, m.Iterator(), impl.isOrdered, th.JsonToKV[string, int](`[["a",0],["b",1],["c",2],["d",3],["e",4]]`))
			// delete head
			m.Delete("a")
			th.ValidateIterator(t, m.Iterator(), impl.isOrdered, th.JsonToKV[string, int](`[["b",1],["c",2],["d",3],["e",4]]`))
			// delete tail
			m.Delete("e")
			th.ValidateIterator(t, m.Iterator(), impl.isOrdered, th.JsonToKV[string, int](`[["b",1],["c",2],["d",3]]`))
			// delete in the middle
			m.Delete("c")
			th.ValidateIterator(t, m.Iterator(), impl.isOrdered, th.JsonToKV[string, int](`[["b",1],["d",3]]`))
			// empty
			m.Delete("b")
			th.ValidateIterator(t, m.Iterator(), impl.isOrdered, th.JsonToKV[string, int](`[["d",3]]`))
			m.Delete("d")
			th.ValidateIterator(t, m.Iterator(), impl.isOrdered, th.JsonToKV[string, int](`[]`))
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

func TestUnmarshalJSONUtilErrAtFirstToken(t *testing.T) {
	m := omap.New[string, int]()
	th.AssertErrNotNil(t, omap.UnmarshalJSON(m.Put, []byte("not a valid json")), "expecting an error with invalid JSON given")
}

func TestUnmarshalJSONUtilErrAfterFirstKey(t *testing.T) {
	m := omap.New[string, int]()
	th.AssertErrNotNil(t, omap.UnmarshalJSON(m.Put, []byte("{\"foo\": 1, bar}")), "expecting an error with invalid JSON given")
}

func TestUnmarshalJSONUtilErrNonStringKeyJSON(t *testing.T) {
	m := omap.New[int, int]()
	th.AssertErrNotNil(t, omap.UnmarshalJSON(m.Put, []byte("{\"1\": 2}")), "expecting an error with invalid JSON given")
}

func TestItMove(t *testing.T) {
	for _, impl := range implementations {
		if !impl.isOrdered {
			continue
		}
		t.Run(impl.name, func(t *testing.T) {
			m := impl.initializerStrInt()
			m.Put("a", 1)
			m.Put("b", 2)
			m.Put("c", 3)
			expected := []th.KeyValue[string, int]{{Key: "a", Value: 1}, {Key: "b", Value: 2}, {Key: "c", Value: 3}}
			it := m.Iterator()
			it.MoveBack()
			expPos := len(expected) - 1
			for it.Prev() {
				if expPos < 0 {
					t.Fatalf("found key/val = %q/%v after last position", it.Key(), it.Value())
				}
				if it.Key() != expected[expPos].Key {
					t.Errorf("expected key %q, found %q", expected[expPos].Key, it.Key())
				}
				if it.Value() != expected[expPos].Value {
					t.Errorf("expected value %v, found %v", expected[expPos].Value, it.Value())
				}
				expPos--
			}
			if it.Prev() {
				t.Error("another call to prev should be false, found true")
			}
			if it.IsValid() {
				t.Error("iterator should not be valid")
			}
			if expPos >= 0 {
				t.Errorf("missing keys/vals: %v", expected[:expPos+1])
			}
			// check if it can move forward again
			it.MoveFront()
			th.ValidateIterator(t, it, impl.isOrdered, expected)
		})
	}
}

func TestGetIteratorAt(t *testing.T) {
	for _, impl := range implementations {
		if !impl.isOrdered {
			continue
		}
		t.Run(impl.name, func(t *testing.T) {
			m := impl.initializerStrInt()
			m.Put("a", 1)
			m.Put("b", 2)
			m.Put("deleted", 3)
			m.Put("d", 4)
			m.Delete("deleted")
			for it1 := m.Iterator(); it1.Next(); {
				it2 := m.GetIteratorAt(it1.Key())
				if !it2.IsValid() {
					t.Errorf("expected iterator to be valid for key %q", it1.Key())
				}
				if it1.Key() != it2.Key() || it1.Value() != it2.Value() {
					t.Errorf("expected key/val = %q/%v, found %q/%v", it1.Key(), it1.Value(), it2.Key(), it2.Value())
				}
			}
			if itNotFound := m.GetIteratorAt("foo"); itNotFound.IsValid() {
				t.Errorf("expected %T.GetIteratorAt(\"foo\") to be not valid, found a valid one", m)
			}
			if itDeleted := m.GetIteratorAt("deleted"); itDeleted.IsValid() {
				t.Errorf("expected %T.GetIteratorAt(\"deleted\") to be not valid, found a valid one", m)
			}
			if itA := m.GetIteratorAt("a"); !itA.IsValid() {
				t.Errorf("expected %T.GetIteratorAt(\"a\") to be valid, found an invalid one", m)
			} else {
				th.ValidateIterator(t, itA, impl.isOrdered, th.JsonToKV[string, int](`[["b",2],["d",4]]`))
			}
			if itB := m.GetIteratorAt("b"); !itB.IsValid() {
				t.Errorf("expected %T.GetIteratorAt(\"b\") to be valid, found an invalid one", m)
			} else {
				th.ValidateIterator(t, itB, impl.isOrdered, th.JsonToKV[string, int](`[["d",4]]`))
			}
			if itD := m.GetIteratorAt("d"); !itD.IsValid() {
				t.Errorf("expected %T.GetIteratorAt(\"d\") to be valid, found an invalid one", m)
			} else {
				th.ValidateIterator(t, itD, impl.isOrdered, th.JsonToKV[string, int](`[]`))
			}
		})
	}
}

func TestPutAfterErrors(t *testing.T) {
	for _, impl := range implementations {
		if !impl.isOrdered {
			continue
		}
		t.Run(impl.name, func(t *testing.T) {
			var invalidIt omap.OMapIterator[string, int]
			m := impl.initializerStrInt()
			th.AssertErrIs(t, m.PutAfter(invalidIt, "x", 0), omap.ErrInvalidIteratorType, "expected PutAfter with invalid it to fail")
			invalidMapIt := impl.initializerStrInt().Iterator()
			th.AssertErrIs(t, m.PutAfter(invalidMapIt, "x", 0), omap.ErrInvalidIteratorMap, "expected PutAfter with different map to fail")
			th.AssertErrIs(t, m.PutAfter(m.Iterator().MoveBack(), "x", 0), omap.ErrInvalidIteratorPos, "expected PutAfter with iterator at EOF to fail")
			m.Put("x", 0)
			deletedRefIt := m.GetIteratorAt("x")
			m.Delete("x")
			th.AssertErrIs(t, m.PutAfter(deletedRefIt, "x", 0), omap.ErrInvalidIteratorPos, "expected PutAfter with iterator at deleted key to fail")
			if impl.name != implSimple { // OMapSimple can't validate this, as it always seek the key/value by the position
				m.Put("x", 0)
				th.AssertErrIs(t, m.PutAfter(deletedRefIt, "x", 0), omap.ErrInvalidIteratorPos, "expected PutAfter with iterator at old reference to fail")
			}
		})
	}
}

func TestPutAfter(t *testing.T) {
	for _, impl := range implementations {
		if !impl.isOrdered {
			continue
		}
		t.Run(impl.name, func(t *testing.T) {
			m := impl.initializerStrInt()
			// Basically put: foo/1, bar/2, baz/3. But in a complex way ????
			// Add bar, then baz
			th.AssertErrNil(t, m.PutAfter(m.Iterator(), "bar", 2), "")
			th.AssertErrNil(t, m.PutAfter(m.GetIteratorAt("bar"), "baz", 3), "")
			// Add foo before baz (which is actually BOF)
			itBeforeBar := m.GetIteratorAt("bar")
			itBeforeBar.Prev()
			th.AssertErrNil(t, m.PutAfter(itBeforeBar, "foo", 1), "")
			// Now, validate
			th.ValidateIterator(t, m.Iterator(), impl.isOrdered, th.JsonToKV[string, int](`[["foo",1],["bar",2],["baz",3]]`))
			// PutAfter at head
			th.AssertErrNil(t, m.PutAfter(m.Iterator(), "HEAD", 0), "")
			// PutAfter at tail
			itBack := m.Iterator().MoveBack()
			itBack.Prev()
			th.AssertErrNil(t, m.PutAfter(itBack, "TAIL", 0), "")
			th.ValidateIterator(t, m.Iterator(), impl.isOrdered, th.JsonToKV[string, int](`[["HEAD",0],["foo",1],["bar",2],["baz",3],["TAIL",0]]`))
			// PutAfter in the middle, after each key, adding upper case values after lower case value
			for _, k := range []string{"foo", "bar", "baz"} {
				kUpper := strings.ToUpper(k)
				it := m.GetIteratorAt(k)
				th.AssertErrNil(t, m.PutAfter(it, kUpper, 0), "")
			}
			th.ValidateIterator(t, m.Iterator(), impl.isOrdered, th.JsonToKV[string, int](`[["HEAD",0],["foo",1],["FOO",0],["bar",2],["BAR",0],["baz",3],["BAZ",0],["TAIL",0]]`))
			// Overwrite, change values from 0 to 42, adding upper cases values in the middle before the lower case value
			itBack = m.Iterator().MoveBack()
			itBack.Prev()
			th.AssertErrNil(t, m.PutAfter(itBack, "TAIL", 42), "")
			th.AssertErrNil(t, m.PutAfter(m.Iterator(), "HEAD", 42), "")
			for _, k := range []string{"foo", "bar", "baz"} {
				kUpper := strings.ToUpper(k)
				it := m.GetIteratorAt(k)
				it.Prev()
				th.AssertErrNil(t, m.PutAfter(it, kUpper, 42), "")
			}
			th.ValidateIterator(t, m.Iterator(), impl.isOrdered, th.JsonToKV[string, int](`[["HEAD",42],["FOO",42],["foo",1],["BAR",42],["bar",2],["BAZ",42],["baz",3],["TAIL", 42]]`))
		})
	}
}

func TestMove(t *testing.T) {
	for _, impl := range implementations {
		if !impl.isOrdered {
			continue
		}
		t.Run(impl.name, func(t *testing.T) {
			m := impl.initializerStrInt()
			m.Put("a", 0)
			m.Put("b", 1)
			m.Put("c", 2)
			m.Put("d", 3)
			m.Put("e", 4)
			if err := omap.MoveAfter(m, "a", "b"); err != nil {
				t.Errorf("unexpected error: %v", err)
			} else {
				th.ValidateIterator(t, m.Iterator(), impl.isOrdered, th.JsonToKV[string, int](`[["b",1],["a",0],["c",2],["d",3],["e",4]]`))
			}
			if err := omap.MoveBefore(m, "a", "b"); err != nil {
				t.Errorf("unexpected error: %v", err)
			} else {
				th.ValidateIterator(t, m.Iterator(), impl.isOrdered, th.JsonToKV[string, int](`[["a",0],["b",1],["c",2],["d",3],["e",4]]`))
			}
			if err := omap.MoveAfter(m, "a", "e"); err != nil {
				t.Errorf("unexpected error: %v", err)
			} else {
				th.ValidateIterator(t, m.Iterator(), impl.isOrdered, th.JsonToKV[string, int](`[["b",1],["c",2],["d",3],["e",4],["a",0]]`))
			}
			if err := omap.MoveFirst(m, "c"); err != nil {
				t.Errorf("unexpected error: %v", err)
			} else {
				th.ValidateIterator(t, m.Iterator(), impl.isOrdered, th.JsonToKV[string, int](`[["c",2],["b",1],["d",3],["e",4],["a",0]]`))
			}
			if err := omap.MoveLast(m, "d"); err != nil {
				t.Errorf("unexpected error: %v", err)
			} else {
				th.ValidateIterator(t, m.Iterator(), impl.isOrdered, th.JsonToKV[string, int](`[["c",2],["b",1],["e",4],["a",0],["d",3]]`))
			}
			// error paths - invalid keys:
			th.AssertErrIs(t, omap.MoveFirst(m, "what"), omap.ErrKeyNotFound, "expecing MoveFirst with invalid refKey to fail with ErrKeyNotFound")
			th.AssertErrIs(t, omap.MoveLast(m, "what"), omap.ErrKeyNotFound, "expecing MoveLast with invalid refKey to fail with ErrKeyNotFound")
			th.AssertErrIs(t, omap.MoveAfter(m, "a", "what"), omap.ErrKeyNotFound, "expecing MoveAfter with invalid refKey to fail with ErrKeyNotFound")
			th.AssertErrIs(t, omap.MoveAfter(m, "what", "a"), omap.ErrKeyNotFound, "expecing MoveAfter with invalid targetKey to fail with ErrKeyNotFound")
			th.AssertErrIs(t, omap.MoveBefore(m, "a", "what"), omap.ErrKeyNotFound, "expecing MoveBefore with invalid refKey to fail with ErrKeyNotFound")
			th.AssertErrIs(t, omap.MoveBefore(m, "what", "a"), omap.ErrKeyNotFound, "expecing MoveBefore with invalid targetKey to fail with ErrKeyNotFound")
		})
	}
}

func TestNotImplementedPanics(t *testing.T) {
	validatePanic := func(t *testing.T, msg string, fct func()) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("%s, no panic called", msg)
			}
		}()
		fct()
	}
	m := omap.NewOMapBuiltin[string, int]()
	validatePanic(t, fmt.Sprintf("%T.GetIteratorAt(...)", m), func() { m.GetIteratorAt("foo") })
	validatePanic(t, fmt.Sprintf("%T.PutAfter(...)", m), func() { _ = m.PutAfter(m.Iterator(), "foo", 1) })
	it := m.Iterator()
	validatePanic(t, fmt.Sprintf("%T.MoveFront()", it), func() { it.MoveFront() })
	validatePanic(t, fmt.Sprintf("%T.MoveBack()", it), func() { it.MoveBack() })
	validatePanic(t, fmt.Sprintf("%T.Prev()", it), func() { it.Prev() })
}
