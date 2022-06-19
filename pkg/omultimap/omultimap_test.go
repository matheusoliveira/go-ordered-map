package omultimap_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/matheusoliveira/go-ordered-map/pkg/omap"
	"github.com/matheusoliveira/go-ordered-map/pkg/omultimap"
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
	omap.EnableOMapBuiltin = true
	implementations = []implDetail{
		{implLinked, true, false, omultimap.NewOMultiMapLinked[string, string]},
		{implSync, true, true, omultimap.NewOMultiMapSync[string, string]},
	}
}

func mustAssertSlicesEqual[V comparable](t *testing.T, msg string, test []V, expected ...V) {
	if len(test) != len(expected) {
		t.Fatalf("%s - len not match, expected %d, found %d", msg, len(expected), len(test))
	} else {
		for i, v := range test {
			if v != expected[i] {
				t.Fatalf("%s - value mismatch at position %d, expected %v, found %v", msg, i, expected[i], v)
			}
		}
	}
}

func iteratorToStringSlice[K comparable, V any](it omap.OMapIterator[K, V]) []string {
	ret := make([]string, 0)
	for it.Next() {
		ret = append(ret, fmt.Sprint(it.Key()), fmt.Sprint(it.Value()))
	}
	return ret
}

func TestNewIsOMultiMapLinked(t *testing.T) {
	exp := reflect.TypeOf(omultimap.NewOMultiMapLinked[string, int]())
	res := reflect.TypeOf(omultimap.New[string, int]())
	if res != exp {
		t.Errorf("expected type %v, found %v", exp, res)
	}
}

func TestBasicOperations(t *testing.T) {
	for _, impl := range implementations {
		t.Run(impl.name, func(t *testing.T) {
			mm := impl.initializerStrStr()
			fooValExp := []string{"1", "2", "3", "4"}
			fooKeyExp := make([]string, len(fooValExp))
			valExp := make([]string, len(fooValExp)*3)
			keyExp := make([]string, len(fooValExp)*3)
			for i, e := range fooValExp {
				fooKeyExp[i] = "foo"
				keyExp[i*3+0] = "foo"
				keyExp[i*3+1] = "bar"
				keyExp[i*3+2] = "baz"
				valExp[i*3+0] = e
				valExp[i*3+1] = e
				valExp[i*3+2] = e
				mm.Put("foo", e)
				mm.Put("bar", e)
				mm.Put("baz", e)
			}
			itFooKey := mm.GetValuesOf("foo")
			itFooVal := mm.GetValuesOf("foo")
			itFullKey := mm.Iterator()
			itFullVal := mm.Iterator()
			mustAssertSlicesEqual(t, "foo keys", omap.IteratorKeysToSlice(itFooKey), fooKeyExp...)
			mustAssertSlicesEqual(t, "foo values", omap.IteratorValuesToSlice(itFooVal), fooValExp...)
			mustAssertSlicesEqual(t, "fullmap keys", omap.IteratorKeysToSlice(itFullKey), keyExp...)
			mustAssertSlicesEqual(t, "fullmap values", omap.IteratorValuesToSlice(itFullVal), valExp...)
			mustAssertSlicesEqual(t, "EOF", []bool{itFooKey.EOF(), itFooVal.EOF(), itFullKey.EOF(), itFullVal.EOF()}, true, true, true, true)
		})
	}
}

func TestDeleteAt(t *testing.T) {
	for _, impl := range implementations {
		t.Run(impl.name, func(t *testing.T) {
			mm := impl.initializerStrStr()
			mm.Put("foo", "1")
			mm.Put("foo", "2")
			mm.Put("foo", "3")
			mm.Put("foo", "4")
			foos := omap.IteratorValuesToSlice(mm.GetValuesOf("foo"))
			mustAssertSlicesEqual(t, "fullmap", foos, "1", "2", "3", "4")
			delIt := mm.Iterator()
			mustAssertSlicesEqual(t, "delIt has next?", []bool{delIt.Next(), delIt.Next()}, true, true)
			if err := mm.DeleteAt(delIt); err != nil {
				t.Fatalf("unexpected error on DeleteAt: %v", err)
			}
			// still points to deleted key/val
			mustAssertSlicesEqual(t, "delIt still at same key/val", []string{delIt.Key(), delIt.Value()}, "foo", "2")
			// check if deleted on the map
			mustAssertSlicesEqual(t, "still has other values", omap.IteratorValuesToSlice(mm.GetValuesOf("foo")), "1", "3", "4")
			// validate rest of iterator
			mustAssertSlicesEqual(t, "check rest of iterator", iteratorToStringSlice(delIt), "foo", "3", "foo", "4")
			// delete first key
			if delFirst := mm.Iterator(); !delFirst.Next() {
				t.Error("delFirst should have Next")
			} else if err := mm.DeleteAt(delFirst); err != nil {
				t.Errorf("unexpected error: %v", err)
			} else {
				mustAssertSlicesEqual(t, "still has other keys", omap.IteratorKeysToSlice(mm.Iterator()), "foo", "foo")
				mustAssertSlicesEqual(t, "still has other values", omap.IteratorValuesToSlice(mm.Iterator()), "3", "4")
			}
			// delete last key
			if delLast := mm.Iterator(); !delLast.Next() {
				t.Error("delLast should have Next")
			} else if !delLast.Next() {
				t.Error("delLast should have Next, twice")
			} else if err := mm.DeleteAt(delLast); err != nil {
				t.Errorf("unexpected error: %v", err)
			} else {
				mustAssertSlicesEqual(t, "still has other keys", omap.IteratorKeysToSlice(mm.Iterator()), "foo")
				mustAssertSlicesEqual(t, "still has other values", omap.IteratorValuesToSlice(mm.Iterator()), "3")
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
			if err := mm.DeleteAt(it2); err == nil {
				t.Error("expected DeleteAt of different map to fail")
			}
			if err := mm.DeleteAt(mm.Iterator()); err == nil {
				t.Error("expected DeleteAt of not started iterator to fail")
			}
			// try delete an iterator at EOF
			itEof := mm.Iterator()
			for itEof.Next() {
			}
			if !itEof.EOF() {
				t.Error("expected EOF")
			}
			if err := mm.DeleteAt(itEof); err == nil {
				t.Error("expected DeleteAt of EOF iterator to fail")
			}
			// try delete after calling DeleteAll on same key
			mmNotFoundAll := impl.initializerStrStr()
			mmNotFoundAll.Put("foo", "bar")
			if itNotFoundAll := mmNotFoundAll.Iterator(); !itNotFoundAll.Next() {
				t.Error("expected next to return true")
			} else {
				mmNotFoundAll.Put("foo", "baz")
				mmNotFoundAll.DeleteAll("foo")
				if err := mmNotFoundAll.DeleteAt(itNotFoundAll); err == nil {
					t.Error("expected delete at after deleting the key to fail")
				}
			}
			// delete from iterator and try to delete again, should fail
			mmNotFoundOne := impl.initializerStrStr()
			mmNotFoundOne.Put("delete-one", "1")
			mmNotFoundOne.Put("delete-one", "2")
			mmNotFoundOne.Put("delete-one", "3")
			if itNotFoundOne := mmNotFoundOne.Iterator(); !itNotFoundOne.Next() {
				t.Error("expected next to return true")
			} else if err := mmNotFoundOne.DeleteAt(itNotFoundOne); err != nil {
				t.Errorf("expected first attempt to mmNotFoundOne.DeleteAt to not fail, error: %v", err)
			} else if err := mmNotFoundOne.DeleteAt(itNotFoundOne); err == nil {
				t.Error("expected second attempt to mmNotFoundOne.DeleteAt to fail")
			}
			// delete all of a given key and try deleting it again from a previous iterator
			mmDeleteAllIt := impl.initializerStrStr()
			mmDeleteAllIt.Put("to-delete", "1")
			mmDeleteAllIt.Put("to-delete", "2")
			mmDeleteAllIt.Put("to-delete", "3")
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
				if err := mmDeleteAllIt.DeleteAt(itTry); err == nil {
					t.Error("expected to mmDeleteAllIt.DeleteAt at same reference again to fail")
				}
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
			var invalidIt omap.OMapIterator[string, string]
			mmInvalidIt := impl.initializerStrStr()
			if err := mmInvalidIt.DeleteAt(invalidIt); err == nil {
				t.Error("expected mmInvalidIt.DeleteAt to fail")
			}
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
			mustAssertSlicesEqual(t, "keys", omap.IteratorKeysToSlice(mm2.Iterator()), "foo", "bar", "foo", "bar", "foo")
			mustAssertSlicesEqual(t, "values", omap.IteratorValuesToSlice(mm2.Iterator()), "1", "2", "3", "4", "5")
			// Test Stringfier
			exp := "omultimap.OMultiMap" + impl.name + expectedStr
			if fmt.Sprint(mm2) != exp {
				t.Errorf("expected %q, found %q", exp, fmt.Sprint(mm2))
			}
		})
	}
}
