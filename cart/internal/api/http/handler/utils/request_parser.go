package utils

import (
	"strconv"
)

func ConvStrToUint64(value string, errOnInvalid error) (uint64, error) {
	if value == "" {
		return 0, errOnInvalid
	}

	v, err := strconv.ParseUint(value, 10, 64)
	if err != nil || v == 0 {
		return 0, errOnInvalid
	}

	return v, nil
}
