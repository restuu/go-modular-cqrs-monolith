package errcode_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"go-modular-cqrs-monolith/platform/errcode"
)

const testCode errcode.Code = "TEST_CODE"

func TestAppError_Is_matchesBySentinelPointer(t *testing.T) {
	sentinel := errcode.New(testCode, "test error")
	assert.ErrorIs(t, sentinel, sentinel, "same pointer must match")
}

func TestAppError_Is_matchesByCode(t *testing.T) {
	sentinel := errcode.New(testCode, "test error")
	other := errcode.New(testCode, "different message")
	assert.ErrorIs(t, other, sentinel, "same Code must match regardless of message")
}

func TestAppError_Is_doesNotMatchDifferentCode(t *testing.T) {
	a := errcode.New(errcode.Code("CODE_A"), "a")
	b := errcode.New(errcode.Code("CODE_B"), "b")
	assert.False(t, errors.Is(a, b), "different codes must not match")
}

func TestAppError_Is_matchesThroughFmtWrap(t *testing.T) {
	sentinel := errcode.New(testCode, "test error")
	wrapped := errcode.Wrap(sentinel, testCode, "outer message")
	assert.ErrorIs(t, wrapped, sentinel, "errors.Is must match through Wrap")
}

func TestAppError_Unwrap_exposesOriginalCause(t *testing.T) {
	cause := errors.New("db error")
	wrapped := errcode.Wrap(cause, errcode.InternalError, "something went wrong")
	assert.ErrorIs(t, wrapped, cause, "Unwrap must expose the original cause")
}

func TestAppError_Error_returnsMessage(t *testing.T) {
	e := errcode.New(testCode, "the message")
	assert.Equal(t, "the message", e.Error())
}

func TestNew_nilCause_Unwrap_returnsNil(t *testing.T) {
	e := errcode.New(testCode, "msg")
	assert.Nil(t, e.Unwrap())
}
