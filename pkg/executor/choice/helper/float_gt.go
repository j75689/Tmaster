package helper

import (
	"fmt"
	"strconv"
)

func NewFloatGt() Helper {
	return Helper(func(variable string, targetString *string, targetInt *int, targetFloat *float64) (bool, error) {
		if targetFloat == nil {
			return false, fmt.Errorf("nil float value")
		}

		vv, err := strconv.ParseFloat(variable, 64)
		if err != nil {
			return false, err
		}

		return vv > *targetFloat, nil
	})
}
