package testhelper

import (
	"fmt"
	"testing"
)

func TestValidation(t *testing.T) {
	tm := &TestingErrorsMock{}
	tm.Errorf("foo: %s", "bar")
	tm.applyValidation(func() {}, func(format string, args ...any) {
		res := fmt.Sprintf(format, args...)
		if res != "validation failed at position 0, expected message \"bar: foo\", found \"foo: bar\"" {
			t.Errorf("invalid final message: %q", res)
		}
	}, "bar: foo")
	tm.applyValidation(func() {}, func(format string, args ...any) {
		res := fmt.Sprintf(format, args...)
		if res != "expected 2 logs, found 1" {
			t.Errorf("invalid final message: %q", res)
		}
	}, "foo: bar", "foo: bar")
}
