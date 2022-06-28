package testhelper

import (
	"errors"
)

func AssertErrNil(t TestingT, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Errorf("%s\nerror: %v", msg, err)
	}
}

func AssertErrNotNil(t TestingT, err error, msg string) {
	t.Helper()
	if err == nil {
		t.Errorf("%v", msg)
	}
}

func AssertErrIs(t TestingT, err error, target error, msg string) {
	t.Helper()
	if !errors.Is(err, target) {
		t.Errorf("%v\nerror: %v", msg, err)
	}
}
