package testhelpers

import "errors"

var ErrForTest = errors.New("error for test")

type NeedCallWithErr struct {
	NeedCall bool
	Err      error
}

func NewNeedCallWithErr(err error) NeedCallWithErr {
	return NeedCallWithErr{
		NeedCall: true,
		Err:      err,
	}
}

type NeedCallWithErrAndResult[T any] struct {
	NeedCall bool
	Result   T
	Err      error
}

func NewNeedCallWithErrAndResult[T any](result T, err error) NeedCallWithErrAndResult[T] {
	return NeedCallWithErrAndResult[T]{
		NeedCall: true,
		Result:   result,
		Err:      err,
	}
}
