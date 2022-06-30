package testhelper_test

import (
	"testing"

	th "github.com/matheusoliveira/go-ordered-map/internal/testhelper"
	"github.com/matheusoliveira/go-ordered-map/omap"
)

//// OMapIterator Mocks ////

// This mock always return false for Prev/Next but always return true for IsValid, which should be
// an impossible state of an iterator, so it is good to test failure paths.
type OMapIteratorMockAsAlwaysValid[K comparable, V any] struct {
}

func NewOMapIteratorMockAsAlwaysValid[K comparable, V any]() omap.OMapIterator[K, V] {
	return &OMapIteratorMockAsAlwaysValid[K, V]{}
}

func (it *OMapIteratorMockAsAlwaysValid[K, V]) Next() bool {
	return false
}

func (it *OMapIteratorMockAsAlwaysValid[K, V]) EOF() bool {
	return false
}

func (it *OMapIteratorMockAsAlwaysValid[K, V]) Key() K {
	var key K
	return key
}

func (it *OMapIteratorMockAsAlwaysValid[K, V]) Value() V {
	var value V
	return value
}

func (it *OMapIteratorMockAsAlwaysValid[K, V]) IsValid() bool {
	return true
}

func (it *OMapIteratorMockAsAlwaysValid[K, V]) MoveFront() omap.OMapIterator[K, V] {
	return nil
}

func (it *OMapIteratorMockAsAlwaysValid[K, V]) MoveBack() omap.OMapIterator[K, V] {
	return nil
}

func (it *OMapIteratorMockAsAlwaysValid[K, V]) Prev() bool {
	return false
}

// This mock always return key/value pairs: foo/1, bar/2 and baz/3.
type OMapFooBarBazIteratorMock struct {
	kv  []th.KeyValue[string, int]
	pos int
}

func NewOMapFooBarBazIteratorMock() omap.OMapIterator[string, int] {
	return &OMapFooBarBazIteratorMock{kv: []th.KeyValue[string, int]{{"foo", 1}, {"bar", 2}, {"baz", 3}}, pos: -1}
}

func (it *OMapFooBarBazIteratorMock) Next() bool {
	it.pos++
	return it.IsValid()
}

func (it *OMapFooBarBazIteratorMock) EOF() bool {
	return it.pos >= len(it.kv)
}

func (it *OMapFooBarBazIteratorMock) Key() string {
	return it.kv[it.pos].Key
}

func (it *OMapFooBarBazIteratorMock) Value() int {
	return it.kv[it.pos].Value
}

func (it *OMapFooBarBazIteratorMock) IsValid() bool {
	return it.pos >= 0 && it.pos < len(it.kv)
}

func (it *OMapFooBarBazIteratorMock) MoveFront() omap.OMapIterator[string, int] {
	it.pos = -1
	return it
}

func (it *OMapFooBarBazIteratorMock) MoveBack() omap.OMapIterator[string, int] {
	it.pos = len(it.kv)
	return it
}

func (it *OMapFooBarBazIteratorMock) Prev() bool {
	it.pos--
	return it.IsValid()
}

//// Unit Tests ////

func TestOK(t *testing.T) {
	it := NewOMapFooBarBazIteratorMock()
	th.ValidateIterator(t, it, true, []th.KeyValue[string, int]{{"foo", 1}, {"bar", 2}, {"baz", 3}})
}

func TestNilKV(t *testing.T) {
	tm := &th.TestingErrorsMock{}
	it := NewOMapFooBarBazIteratorMock()
	th.ValidateIteratorForward(tm, it, true, nil)
	tm.Validate(t, "expResults slices is nil")
	tm.Reset()
	th.ValidateIteratorBackward(tm, it, true, nil)
	tm.Validate(t, "expResults slices is nil")
}

func TestSlicesToKeyValue(t *testing.T) {
	it := NewOMapFooBarBazIteratorMock()
	exp := th.SlicesToKeyValue([]string{"foo", "bar", "baz"}, []int{1, 2, 3})
	th.ValidateIterator(t, it, true, exp)
	if wrong := th.SlicesToKeyValue([]string{"foo", "bar", "baz"}, []int{1, 2, 3, 4}); wrong != nil {
		t.Error("slice should be nil")
	}
	if wrong := th.SlicesToKeyValue([]string{"foo", "bar", "baz", "zaz"}, []int{1, 2, 3}); wrong != nil {
		t.Error("slice should be nil")
	}
}

func TestOverflow(t *testing.T) {
	tm := &th.TestingErrorsMock{}
	it := NewOMapFooBarBazIteratorMock()
	th.ValidateIterator(tm, it, true, []th.KeyValue[string, int]{{"foo", 1}, {"bar", 2}})
	tm.Validate(t, "overflow, expecting max of 2 values")
}

func TestWrongKey(t *testing.T) {
	tm := &th.TestingErrorsMock{}
	it := NewOMapFooBarBazIteratorMock()
	// forward
	th.ValidateIterator(tm, it, true, []th.KeyValue[string, int]{{"foo", 1}, {"bar", 2}, {"BAZ", 3}})
	tm.Validate(t, "expecting key \"BAZ\" at position 2, key \"baz\" found")
	// backward
	tm.Reset()
	th.ValidateIteratorBackward(tm, it.MoveBack(), true, []th.KeyValue[string, int]{{"foo", 1}, {"bar", 2}, {"BAZ", 3}})
	tm.Validate(t, "backward validation - expecting key \"BAZ\" at position 2, key \"baz\" found")
	// unordered
	tm.Reset()
	th.ValidateIterator(tm, it.MoveFront(), false, []th.KeyValue[string, int]{{"foo", 1}, {"bar", 2}, {"BAZ", 3}})
	tm.Validate(t, "found unexpected key/value pair: baz/3")
}

func TestWrongValue(t *testing.T) {
	tm := &th.TestingErrorsMock{}
	it := NewOMapFooBarBazIteratorMock()
	// forward
	th.ValidateIterator(tm, it, true, []th.KeyValue[string, int]{{"foo", 1}, {"bar", 2}, {"baz", 4}})
	tm.Validate(t, "invalid value for key \"baz\", at position 2, expected 4, found 3")
	// backward
	tm.Reset()
	th.ValidateIteratorBackward(tm, it.MoveBack(), true, []th.KeyValue[string, int]{{"foo", 1}, {"bar", 2}, {"baz", 4}})
	tm.Validate(t, "backward validation - invalid value for key \"baz\", at position 2, expected 4, found 3")
}

func TestExceed(t *testing.T) {
	tm := &th.TestingErrorsMock{}
	it := NewOMapFooBarBazIteratorMock()
	// forward
	th.ValidateIterator(tm, it, true, []th.KeyValue[string, int]{{"foo", 1}, {"bar", 2}, {"baz", 3}, {"zaz", 4}})
	tm.Validate(t, "values not processed: [{zaz 4}]")
	// backward
	tm.Reset()
	th.ValidateIteratorBackward(tm, it.MoveBack(), true, []th.KeyValue[string, int]{{"zaz", 4}, {"foo", 1}, {"bar", 2}, {"baz", 3}})
	tm.Validate(t, "backward validation - values not processed: [{zaz 4}]")
}

func TestForwardStopPartial(t *testing.T) {
	it := NewOMapFooBarBazIteratorMock()
	th.ValidateIteratorBackward(t, it.MoveBack(), false, []th.KeyValue[string, int]{{"bar", 2}, {"baz", 3}})
	// overflow
	tm := &th.TestingErrorsMock{}
	th.ValidateIteratorBackward(tm, it.MoveBack(), true, []th.KeyValue[string, int]{{"bar", 2}, {"baz", 3}})
	tm.Validate(t, "backward validation - overflow, expecting max of 2 values")
}

func TestMockValidations(t *testing.T) {
	tm := &th.TestingErrorsMock{}
	it := NewOMapIteratorMockAsAlwaysValid[string, int]()
	// forward
	th.ValidateIteratorForward(tm, it, true, []th.KeyValue[string, int]{})
	tm.Validate(t, "EOF returned false instead of true at end of loop", "IsValid returned true instead of false at end of loop")
	// backward
	tm.Reset()
	it.MoveFront()
	th.ValidateIteratorBackward(tm, it, true, []th.KeyValue[string, int]{})
	tm.Validate(t, "backward validation - IsValid returned true instead of false at end of loop")
}

func TestJsonToKV(t *testing.T) {
	kv := th.JsonToKV[string, int](`[["foo", 1], ["bar", 2]]`)
	if len(kv) != 2 {
		t.Errorf("expected len of 2, found %d", len(kv))
	}
	if kv := th.JsonToKV[string, int](`foo`); kv != nil {
		t.Error("expected error when given an invalid JSON")
	}
	if kv := th.JsonToKV[string, int](`["foo", 1]`); kv != nil {
		t.Error("expected error when given single array value")
	}
	if kv := th.JsonToKV[string, int](`[["foo", 1], [2, 2]]`); kv != nil {
		t.Error("expected error when given incorrect key format")
	}
	if kv := th.JsonToKV[string, int](`[["foo", 1], ["bar", "bar"]]`); kv != nil {
		t.Error("expected error when given incorrect value format")
	}
	if kv := th.JsonToKV[string, int](`[["foo", 1], ["bar", 2, 3]]`); kv != nil {
		t.Error("expected error when given incorrect array length")
	}
}
