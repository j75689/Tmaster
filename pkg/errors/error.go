package errors

import (
	"github.com/j75689/Tmaster/pkg/graph/model"
)

// error type check
var (
	_ error = &ErrRuntime{}
	_ error = &ErrPermission{}
	_ error = &ErrTaskFailed{}
	_ error = &ErrTimeout{}
)

type Error interface {
	Error() string
	Unwrap() error
	ErrCode() model.ErrorCode
}

type ErrRuntime struct{ err error }

func (e *ErrRuntime) Error() string {
	return e.err.Error()
}

func (e *ErrRuntime) Unwrap() error {
	return e.err
}

func (e *ErrRuntime) ErrCode() model.ErrorCode {
	return model.ErrorCodeRuntime
}

type ErrTimeout struct{ err error }

func (e *ErrTimeout) Error() string {
	return e.err.Error()
}

func (e *ErrTimeout) Unwrap() error {
	return e.err
}

func (e *ErrTimeout) ErrCode() model.ErrorCode {
	return model.ErrorCodeTimeout
}

type ErrTaskFailed struct{ err error }

func (e *ErrTaskFailed) Error() string {
	return e.err.Error()
}

func (e *ErrTaskFailed) Unwrap() error {
	return e.err
}

func (e *ErrTaskFailed) ErrCode() model.ErrorCode {
	return model.ErrorCodeTaskfailed
}

type ErrPermission struct{ err error }

func (e *ErrPermission) Error() string {
	return e.err.Error()
}

func (e *ErrPermission) Unwrap() error {
	return e.err
}

func (e *ErrPermission) ErrCode() model.ErrorCode {
	return model.ErrorCodePermissions
}

// NewRuntimeError returns a warpped ErrRuntime
func NewRuntimeError(e error) *ErrRuntime {
	return &ErrRuntime{e}
}

// NewTimeoutError returns a warpped ErrTimeout
func NewTimeoutError(e error) *ErrTimeout {
	return &ErrTimeout{e}
}

// NewTaskFailedError returns a warpped ErrTaskFailed
func NewTaskFailedError(e error) *ErrTaskFailed {
	return &ErrTaskFailed{e}
}

// NewPermissionError returns a warpped ErrPermission
func NewPermissionError(e error) *ErrPermission {
	return &ErrPermission{e}
}
