package helper

import (
	"fmt"
	"strconv"
)

func NewIntEq() Helper {
	return Helper(func(variable string, targetString *string, targetInt *int, targetFloat *float64) (bool, error) {
		if targetInt == nil {
			return false, fmt.Errorf("nil int value")
		}

		vv, err := strconv.ParseInt(variable, 10, 64)
		if err != nil {
			return false, err
		}

		return vv == int64(*targetInt), nil
	})
}
