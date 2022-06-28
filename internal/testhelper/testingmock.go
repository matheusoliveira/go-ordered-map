package testhelper

import (
	"fmt"
	"testing"
)

type TestingT interface {
	Helper()
	Errorf(format string, args ...any)
}

type TestingErrorsMock struct {
	logs []string
}

func (tm *TestingErrorsMock) Helper() {
	// no-op
}

var _ TestingT = (*TestingErrorsMock)(nil)

func (tm *TestingErrorsMock) Errorf(format string, args ...any) {
	result := fmt.Sprintf(format, args...)
	if tm.logs == nil {
		tm.logs = make([]string, 0, 1)
	}
	tm.logs = append(tm.logs, result)
}

func (tm *TestingErrorsMock) Validate(t *testing.T, expectedErrors ...string) {
	t.Helper()
	tm.applyValidation(t.Helper, t.Errorf, expectedErrors...)
}

func (tm *TestingErrorsMock) applyValidation(helperFunc func(), errorfFunc func(string, ...any), expectedErrors ...string) {
	helperFunc()
	if len(tm.logs) != len(expectedErrors) {
		errorfFunc("expected %d logs, found %d", len(expectedErrors), len(tm.logs))
	} else {
		for i := 0; i < len(expectedErrors); i++ {
			if expectedErrors[i] != tm.logs[i] {
				errorfFunc("validation failed at position %d, expected message %q, found %q", i, expectedErrors[i], tm.logs[i])
			}
		}
	}
}

func (tm *TestingErrorsMock) Reset() {
	tm.logs = tm.logs[0:0]
}
