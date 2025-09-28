package converter

import (
	"fmt"
	"math"
)

func SafeInt64ToUint32(n int64) (uint32, error) {
	if n < 0 {
		return 0, fmt.Errorf("cannot convert negative value %d to uint32", n)
	}
	if n > int64(math.MaxUint32) {
		return 0, fmt.Errorf("value %d exceeds uint32 max", n)
	}
	return uint32(n), nil // #nosec G115
}
