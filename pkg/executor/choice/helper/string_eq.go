package helper

import (
	"fmt"
	"strconv"
)

func NewStringEq() Helper {
	return Helper(func(variable string, targetString *string, targetInt *int, targetFloat *float64) (bool, error) {
		if targetString == nil {
			return false, fmt.Errorf("nil string value")
		}
		return variable == strconv.Quote(*targetString), nil
	})
}
