package testhelper_test

import (
	"errors"
	"fmt"
	"testing"

	th "github.com/matheusoliveira/go-ordered-map/internal/testhelper"
)

func TestAssertErrNil(t *testing.T) {
	var err error
	th.AssertErrNil(t, err, "unexpected error")
	tm := &th.TestingErrorsMock{}
	th.AssertErrNil(tm, errors.New("fail"), "fail path")
	tm.Validate(t, "fail path\nerror: fail")
}

func TestAssertErrNotNil(t *testing.T) {
	var err error
	th.AssertErrNotNil(t, errors.New("test"), "should not be nil")
	tm := &th.TestingErrorsMock{}
	th.AssertErrNotNil(tm, err, "fail path")
	tm.Validate(t, "fail path")
}

func TestAssertErrIs(t *testing.T) {
	targetErr := errors.New("target")
	failTargetErr := errors.New("fail")
	err := fmt.Errorf("wrapping: %w", targetErr)
	th.AssertErrIs(t, err, targetErr, "should be targetErr")
	tm := &th.TestingErrorsMock{}
	th.AssertErrIs(tm, err, failTargetErr, "fail path, wrong target")
	tm.Validate(t, "fail path, wrong target\nerror: wrapping: target")
	tm.Reset()
	th.AssertErrIs(tm, (error)(nil), targetErr, "fail path, nil error")
	tm.Validate(t, "fail path, nil error\nerror: <nil>")
}
