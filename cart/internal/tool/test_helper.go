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
