package omaptestsutil_test

import (
	"fmt"
	"testing"

	omt "github.com/matheusoliveira/go-ordered-map/pkg/internal/omaptestsutil"
	"github.com/matheusoliveira/go-ordered-map/pkg/omap"
)

//// TestingMock ////

type TestingMock struct {
	logs []TestingMockLog
}

type TestingMockLog struct {
	Level   TestingMockLogLevel
	Message string
}

type TestingMockLogLevel int

const (
	Log TestingMockLogLevel = iota
	Error
	Fail
)

func (tm *TestingMock) Helper() {
	// no-op
}

func (tm *TestingMock) Errorf(format string, args ...any) {
	result := fmt.Sprintf(format, args...)
	if tm.logs == nil {
		tm.logs = make([]TestingMockLog, 0, 1)
	}
	tm.logs = append(tm.logs, TestingMockLog{Error, result})
}

func (tm *TestingMock) Validate(t *testing.T, expected ...TestingMockLog) {
	t.Helper()
	if len(tm.logs) != len(expected) {
		t.Log(tm.logs)
		t.Errorf("expected %d logs, found %d", len(expected), len(tm.logs))
	} else {
		for i := 0; i < len(expected); i++ {
			if expected[i].Level != tm.logs[i].Level {
				t.Errorf("validation failed at position %d, expected level %d, found %d", i, expected[i].Level, tm.logs[i].Level)
			} else if expected[i].Message != tm.logs[i].Message {
				t.Errorf("validation failed at position %d, expected message %q, found %q", i, expected[i].Message, tm.logs[i].Message)
			}
		}
	}
}

//// OMapIterator Mock ////

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

//// Unit Tests ////

func TestOK(t *testing.T) {
	m := omap.New[string, int]()
	m.Put("foo", 1)
	m.Put("bar", 2)
	m.Put("baz", 3)
	omt.ValidateIterator(t, m.Iterator(), true, []omt.KeyValue[string, int]{{"foo", 1}, {"bar", 2}, {"baz", 3}})
}

func TestOverflow(t *testing.T) {
	tm := &TestingMock{}
	m := omap.New[string, int]()
	m.Put("foo", 1)
	m.Put("bar", 2)
	m.Put("baz", 3)
	omt.ValidateIterator(tm, m.Iterator(), true, []omt.KeyValue[string, int]{{"foo", 1}, {"bar", 2}})
	tm.Validate(t, TestingMockLog{Error, "overflow, expecting max of 2 values"})
}

func TestWrongKey(t *testing.T) {
	tm := &TestingMock{}
	m := omap.New[string, int]()
	m.Put("foo", 1)
	m.Put("bar", 2)
	m.Put("baz", 3)
	// forward
	omt.ValidateIterator(tm, m.Iterator(), true, []omt.KeyValue[string, int]{{"foo", 1}, {"bar", 2}, {"BAZ", 3}})
	tm.Validate(t, TestingMockLog{Error, "expecting key \"BAZ\" at position 2, key \"baz\" found"})
	// backward
	tm = &TestingMock{}
	omt.ValidateIteratorBackward(tm, m.Iterator().MoveBack(), true, []omt.KeyValue[string, int]{{"foo", 1}, {"bar", 2}, {"BAZ", 3}})
	tm.Validate(t, TestingMockLog{Error, "backward validation - expecting key \"BAZ\" at position 2, key \"baz\" found"})
	// unordered
	tm = &TestingMock{}
	omt.ValidateIterator(tm, m.Iterator(), false, []omt.KeyValue[string, int]{{"foo", 1}, {"bar", 2}, {"BAZ", 3}})
	tm.Validate(t, TestingMockLog{Error, "found unexpected key/value pair: baz/3"})
}

func TestWrongValue(t *testing.T) {
	tm := &TestingMock{}
	m := omap.New[string, int]()
	m.Put("foo", 1)
	m.Put("bar", 2)
	m.Put("baz", 3)
	// forward
	omt.ValidateIterator(tm, m.Iterator(), true, []omt.KeyValue[string, int]{{"foo", 1}, {"bar", 2}, {"baz", 4}})
	tm.Validate(t, TestingMockLog{Error, "invalid value for key \"baz\", at position 2, expected 4, found 3"})
	// backward
	tm = &TestingMock{}
	omt.ValidateIteratorBackward(tm, m.Iterator().MoveBack(), true, []omt.KeyValue[string, int]{{"foo", 1}, {"bar", 2}, {"baz", 4}})
	tm.Validate(t, TestingMockLog{Error, "backward validation - invalid value for key \"baz\", at position 2, expected 4, found 3"})
}

func TestExceed(t *testing.T) {
	tm := &TestingMock{}
	m := omap.New[string, int]()
	m.Put("foo", 1)
	m.Put("bar", 2)
	m.Put("baz", 3)
	// forward
	omt.ValidateIterator(tm, m.Iterator(), true, []omt.KeyValue[string, int]{{"foo", 1}, {"bar", 2}, {"baz", 3}, {"zaz", 4}})
	tm.Validate(t, TestingMockLog{Error, "values not processed: [{zaz 4}]"})
	// backward
	tm = &TestingMock{}
	omt.ValidateIteratorBackward(tm, m.Iterator().MoveBack(), true, []omt.KeyValue[string, int]{{"zaz", 4}, {"foo", 1}, {"bar", 2}, {"baz", 3}})
	tm.Validate(t, TestingMockLog{Error, "backward validation - values not processed: [{zaz 4}]"})
}

func TestForwardStopPartial(t *testing.T) {
	m := omap.New[string, int]()
	m.Put("foo", 1)
	m.Put("bar", 2)
	m.Put("baz", 3)
	omt.ValidateIteratorBackward(t, m.Iterator().MoveBack(), false, []omt.KeyValue[string, int]{{"bar", 2}, {"baz", 3}})
	// overflow
	tm := &TestingMock{}
	omt.ValidateIteratorBackward(tm, m.Iterator().MoveBack(), true, []omt.KeyValue[string, int]{{"bar", 2}, {"baz", 3}})
	tm.Validate(t, TestingMockLog{Error, "backward validation - overflow, expecting max of 2 values"})
}

func TestMockValidations(t *testing.T) {
	tm := &TestingMock{}
	it := NewOMapIteratorMockAsAlwaysValid[string, int]()
	// forward
	omt.ValidateIteratorForward(tm, it, true, []omt.KeyValue[string, int]{})
	tm.Validate(t,
		TestingMockLog{Error, "EOF returned false instead of true at end of loop"},
		TestingMockLog{Error, "IsValid returned true instead of false at end of loop"},
	)
	// backward
	tm = &TestingMock{}
	omt.ValidateIteratorBackward(tm, it, true, []omt.KeyValue[string, int]{})
	tm.Validate(t, TestingMockLog{Error, "backward validation - IsValid returned true instead of false at end of loop"})
}
