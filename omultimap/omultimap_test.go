package omultimap_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"testing"

	th "github.com/matheusoliveira/go-ordered-map/internal/testhelper"
	"github.com/matheusoliveira/go-ordered-map/omap"
	"github.com/matheusoliveira/go-ordered-map/omultimap"
)

const (
	implLinked = "Linked"
	implSync   = "Sync"
)

type implDetail struct {
	name              string
	isOrdered         bool
	isParallelSafe    bool
	initializerStrStr func() omultimap.OMultiMap[string, string]
}

var implementations []implDetail

func init() {
	// TODO: we wrap New* calls into anonym function due to a golangci-lint bug, we can simple use
	//       the function after golangci-lint solves the bug, see PR #8 for more details.
	//       Issue: https://github.com/golangci/golangci-lint/issues/2859
	omap.EnableOMapBuiltin = true
	implementations = []implDetail{
		{
			implLinked,
			true,
			false,
			func() omultimap.OMultiMap[string, string] { return omultimap.NewOMultiMapLinked[string, string]() },
		},
		{
			implSync,
			true,
			true,
			func() omultimap.OMultiMap[string, string] { return omultimap.NewOMultiMapSync[string, string]() },
		},
	}
}

func TestNewIsOMultiMapLinked(t *testing.T) {
	exp := reflect.TypeOf(omultimap.NewOMultiMapLinked[string, int]())
	res := reflect.TypeOf(omultimap.New[string, int]())
	if res != exp {
		t.Errorf("expected type %v, found %v", exp, res)
	}
}

func TestBasicOperations(t *testing.T) {
	keys := []string{"foo", "bar", "baz"}
	values := []string{"1", "2", "3", "4"}
	fooExp := make([]th.KeyValue[string, string], len(values))
	expected := make([]th.KeyValue[string, string], len(keys)*len(values))
	for i, v := range values {
		fooExp[i] = th.KeyValue[string, string]{Key: "foo", Value: v}
		for j, k := range keys {
			expected[i*len(keys)+j] = th.KeyValue[string, string]{
				Key:   k,
				Value: v,
			}
		}
	}
	for _, impl := range implementations {
		t.Run(impl.name, func(t *testing.T) {
			mm := impl.initializerStrStr()
			// no-op
			mm.Put("foo")
			mm.Put("x")
			// add key/val pairs
			for _, exp := range expected {
				mm.Put(exp.Key, exp.Value)
			}
			if mm.Len() != len(expected) {
				t.Errorf("expected len of %d, found %d", len(expected), mm.Len())
			}
			it := mm.Iterator()
			itFoo := mm.GetValuesOf("foo")
			itFoo.MoveFront()
			// Validate
			th.ValidateIterator(t, it, true, expected)
			th.ValidateIterator(t, itFoo, true, fooExp)
			// MoveBack
			th.ValidateIteratorBackward(t, it.MoveBack(), true, expected)
			th.ValidateIteratorBackward(t, itFoo.MoveBack(), true, fooExp)
			if it.Prev() {
				t.Error("expected it.Prev to return false")
			}
			if itFoo.Prev() {
				t.Error("expected itFoo.Prev to return false")
			}
			// MoveFront
			th.ValidateIteratorForward(t, it.MoveFront(), true, expected)
			th.ValidateIteratorForward(t, itFoo.MoveFront(), true, fooExp)
			if it.Next() {
				t.Error("expected it.Next to return false")
			}
			if itFoo.Next() {
				t.Error("expected itFoo.Next to return false")
			}
		})
	}
}

func TestMultiPut(t *testing.T) {
	for _, impl := range implementations {
		t.Run(impl.name, func(t *testing.T) {
			mm := impl.initializerStrStr()
			mm.Put("x")
			mm.Put("x", "1", "2", "3")
			mm.Put("y", "4", "5", "6")
			mm.Put("x", "7", "8", "9")
			exp := th.JsonToKV[string, string](`[["x","1"],["x","2"],["x","3"],["y","4"],["y","5"],["y","6"],["x","7"],["x","8"],["x","9"]]`)
			expX := th.JsonToKV[string, string](`[["x","1"],["x","2"],["x","3"],["x","7"],["x","8"],["x","9"]]`)
			expY := th.JsonToKV[string, string](`[["y","4"],["y","5"],["y","6"]]`)
			th.ValidateIterator(t, mm.Iterator(), true, exp)
			th.ValidateIterator(t, mm.GetValuesOf("x"), true, expX)
			th.ValidateIterator(t, mm.GetValuesOf("y"), true, expY)
		})
	}
}

func TestDeleteAt(t *testing.T) {
	for _, impl := range implementations {
		t.Run(impl.name, func(t *testing.T) {
			mm := impl.initializerStrStr()
			mm.Put("foo", "1", "2", "3")
			mm.Put("foo", "4")
			th.ValidateIterator(t, mm.GetValuesOf("foo"), true, th.JsonToKV[string, string](`[["foo","1"],["foo","2"],["foo","3"],["foo","4"]]`))
			delIt := mm.Iterator()
			if !delIt.Next() {
				t.Fatal("expected delIt.Next() to return true, first time")
			}
			if !delIt.Next() {
				t.Fatal("expected delIt.Next() to return true, second time")
			}
			th.AssertErrNil(t, mm.DeleteAt(delIt), "unexpected error on DeleteAt")
			// still points to deleted key/val
			if delIt.Key() != "foo" || delIt.Value() != "2" {
				t.Errorf("expected key/val = \"foo\"/\"2\", found %q/%q", delIt.Key(), delIt.Value())
			}
			// check if deleted on the map
			th.ValidateIterator(t, mm.GetValuesOf("foo"), true, th.JsonToKV[string, string](`[["foo","1"],["foo","3"],["foo","4"]]`))
			// validate rest of iterator
			th.ValidateIterator(t, delIt, true, th.JsonToKV[string, string](`[["foo","3"],["foo","4"]]`))
			// delete first key
			if delFirst := mm.Iterator(); !delFirst.Next() {
				t.Error("expected delFirst.Next() to return true")
			} else if err := mm.DeleteAt(delFirst); err != nil {
				t.Errorf("unexpected error: %v", err)
			} else {
				th.ValidateIterator(t, mm.Iterator(), true, th.JsonToKV[string, string](`[["foo","3"],["foo","4"]]`))
			}
			// delete last key
			if delLast := mm.Iterator(); !delLast.Next() {
				t.Error("expected delLast.Next() to return true, first time")
			} else if !delLast.Next() {
				t.Error("expected delLast.Next() to return true, second time")
			} else if err := mm.DeleteAt(delLast); err != nil {
				t.Errorf("unexpected error: %v", err)
			} else {
				th.ValidateIteratorForward(t, mm.Iterator(), true, th.JsonToKV[string, string](`[["foo","3"]]`))
			}
		})
	}
}

func TestDeleteAtErrors(t *testing.T) {
	for _, impl := range implementations {
		t.Run(impl.name, func(t *testing.T) {
			mm := impl.initializerStrStr()
			mm.Put("foo", "bar")
			mm2 := impl.initializerStrStr()
			mm2.Put("foo", "bar")
			it2 := mm2.Iterator()
			th.AssertErrIs(t, mm.DeleteAt(it2), omap.ErrInvalidIteratorMap, "expected DeleteAt of different map to fail with ErrInvalidIteratorMap")
			th.AssertErrIs(t, mm.DeleteAt(mm.Iterator()), omap.ErrInvalidIteratorPos, "expected DeleteAt of not started iterator to fail with ErrInvalidIteratorPos")
			// try delete an iterator at EOF
			itEof := mm.Iterator()
			for itEof.Next() {
			}
			if !itEof.EOF() {
				t.Error("expected EOF")
			}
			th.AssertErrIs(t, mm.DeleteAt(itEof), omap.ErrInvalidIteratorPos, "expected DeleteAt of EOF iterator to fail with ErrInvalidIteratorPos")
			// try delete after calling DeleteAll on same key
			mmNotFoundAll := impl.initializerStrStr()
			mmNotFoundAll.Put("foo", "bar")
			if itNotFoundAll := mmNotFoundAll.Iterator(); !itNotFoundAll.Next() {
				t.Error("expected next to return true")
			} else {
				mmNotFoundAll.Put("foo", "baz")
				mmNotFoundAll.DeleteAll("foo")
				th.AssertErrIs(t, mmNotFoundAll.DeleteAt(itNotFoundAll), omap.ErrInvalidIteratorKey, "expected delete at after deleting the key to fail with ErrInvalidIteratorKey")
			}
			// delete from iterator and try to delete again, should fail
			mmNotFoundOne := impl.initializerStrStr()
			mmNotFoundOne.Put("delete-one", "1", "2", "3")
			if itNotFoundOne := mmNotFoundOne.Iterator(); !itNotFoundOne.Next() {
				t.Error("expected next to return true")
			} else if err := mmNotFoundOne.DeleteAt(itNotFoundOne); err != nil {
				t.Errorf("expected first attempt to mmNotFoundOne.DeleteAt to not fail, error: %v", err)
			} else {
				th.AssertErrIs(t, mmNotFoundOne.DeleteAt(itNotFoundOne), omap.ErrInvalidIteratorKey, "expected second attempt to mmNotFoundOne.DeleteAt to fail ErrInvalidIteratorKey")
			}
			// delete all of a given key and try deleting it again from a previous iterator
			mmDeleteAllIt := impl.initializerStrStr()
			mmDeleteAllIt.Put("to-delete", "1", "2", "3")
			mmDeleteAllIt.Put("to-keep", "1")
			itTry := mmDeleteAllIt.Iterator()
			itTry.Next()
			if itTry.Key() != "to-delete" {
				t.Errorf("expecting \"to-delete\", found %q", itTry.Key())
			} else {
				for it := mmDeleteAllIt.Iterator(); it.Next(); {
					if it.Key() == "to-delete" {
						if err := mmDeleteAllIt.DeleteAt(it); err != nil {
							t.Errorf("unexpected error: %v", err)
						}
					}
				}
				if elems := omap.IteratorValuesToSlice(mmDeleteAllIt.GetValuesOf("to-delete")); len(elems) != 0 {
					t.Errorf("should not found any value with \"to-delete\" key, found %d values", len(elems))
				}
				th.AssertErrIs(t, mmDeleteAllIt.DeleteAt(itTry), omap.ErrInvalidIteratorKey, "expected to mmDeleteAllIt.DeleteAt at same reference again to fail with ErrInvalidIteratorKey")
			}
			// test if MustDeleteAt panics
			func() {
				defer func() {
					r := recover()
					if r == nil {
						t.Error("expected MustDeleteAt to panic, at defer")
					}
				}()
				mm1 := impl.initializerStrStr()
				mm2 := impl.initializerStrStr()
				mm1.MustDeleteAt(mm2.Iterator())
				t.Error("expected MustDeleteAt to panic, after call")
			}()
			om := omap.New[string, string]()
			invalidIt := om.Iterator()
			mmInvalidIt := impl.initializerStrStr()
			th.AssertErrIs(t, mmInvalidIt.DeleteAt(invalidIt), omap.ErrInvalidIteratorType, "expected mmInvalidIt.DeleteAt to fail with ErrInvalidIteratorType")
		})
	}
}

func TestJSON(t *testing.T) {
	expectedJson := `{"foo":"1","bar":"2","foo":"3","bar":"4","foo":"5"}`
	expectedStr := `[foo:1 bar:2 foo:3 bar:4 foo:5]`
	for _, impl := range implementations {
		t.Run(impl.name, func(t *testing.T) {
			mm := impl.initializerStrStr()
			mm.Put("foo", "1")
			mm.Put("bar", "2")
			mm.Put("foo", "3")
			mm.Put("bar", "4")
			mm.Put("foo", "5")
			if js, err := json.Marshal(mm); err != nil {
				t.Errorf("json.Marshal failed with error: %v", err)
			} else if string(js) != expectedJson {
				t.Errorf("unexpected json output: %v", string(js))
			}
			// Call Unmarshal and Marshal again
			mm2 := impl.initializerStrStr()
			if err := json.Unmarshal([]byte(expectedJson), mm2); err != nil {
				t.Errorf("json.Unmarshal failed with error: %v", err)
			} else if js, err := json.Marshal(mm2); err != nil {
				t.Errorf("json.Marshal failed with error: %v", err)
			} else if string(js) != expectedJson {
				t.Errorf("unexpected json output: %v", string(js))
			}
			th.ValidateIterator(t, mm2.Iterator(), true, th.JsonToKV[string, string](`[["foo","1"],["bar","2"],["foo","3"],["bar","4"],["foo","5"]]`))
			// Test Stringfier
			exp := "omultimap.OMultiMap" + impl.name + expectedStr
			if fmt.Sprint(mm2) != exp {
				t.Errorf("expected %q, found %q", exp, fmt.Sprint(mm2))
			}
		})
	}
}

func TestPutAfter(t *testing.T) {
	for _, impl := range implementations {
		t.Run(impl.name, func(t *testing.T) {
			mm := impl.initializerStrStr()
			// invalid iterator
			om := omap.New[string, string]()
			invalidIt := om.Iterator()
			th.AssertErrIs(t, mm.PutAfter(invalidIt, "x", "y"), omap.ErrInvalidIteratorType, "expected mmInvalidIt.PutAfter to fail with ErrInvalidIteratorType")
			// add at begin
			if err := mm.PutAfter(mm.Iterator(), "foo", "1"); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			th.ValidateIterator(t, mm.Iterator(), true, th.JsonToKV[string, string](`[["foo","1"]]`))
			// add at end
			itEnd := mm.Iterator().MoveBack()
			itEnd.Prev()
			if err := mm.PutAfter(itEnd, "foo", "3"); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			th.ValidateIterator(t, mm.Iterator(), true, th.JsonToKV[string, string](`[["foo","1"],["foo","3"]]`))
			// add at the middle
			itMiddle := mm.Iterator()
			itMiddle.Next()
			if err := mm.PutAfter(itMiddle, "foo", "2"); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			th.ValidateIterator(t, mm.Iterator(), true, th.JsonToKV[string, string](`[["foo","1"],["foo","2"],["foo","3"]]`))
			th.ValidateIterator(t, mm.GetValuesOf("foo"), true, th.JsonToKV[string, string](`[["foo","1"],["foo","2"],["foo","3"]]`))
			// add different keys
			its := []omap.OMapIterator[string, string]{mm.Iterator(), mm.Iterator(), mm.Iterator(), mm.Iterator()}
			// its[0] stays at front, nothing to do
			// its[1] stays at foo/1:
			its[1].Next() // foo/1
			// its[2] stays at foo/2:
			its[2].Next() // foo/1
			its[2].Next() // foo/2
			// its[3] stays at foo/3:
			its[3].Next() // foo/1
			its[3].Next() // foo/2
			its[3].Next() // foo/3
			for i, it := range its {
				if err := mm.PutAfter(it, "bar", "bar/"+strconv.Itoa(i)); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
			th.ValidateIterator(t, mm.Iterator(), true, th.JsonToKV[string, string](`[["bar","bar/0"],["foo","1"],["bar","bar/1"],["foo","2"],["bar","bar/2"],["foo","3"],["bar","bar/3"]]`))
			th.ValidateIterator(t, mm.GetValuesOf("foo"), true, th.JsonToKV[string, string](`[["foo","1"],["foo","2"],["foo","3"]]`))
			th.ValidateIterator(t, mm.GetValuesOf("bar"), true, th.JsonToKV[string, string](`[["bar","bar/0"],["bar","bar/1"],["bar","bar/2"],["bar","bar/3"]]`))
		})
	}
}
