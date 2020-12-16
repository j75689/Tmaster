package collection

import (
	"reflect"
	"strings"

	"github.com/j75689/Tmaster/pkg/graph/model"
)

// ContainsError returns true while target in slice of source
func ContainsError(target *model.ErrorCode, source []*model.ErrorCode) bool {
	all := model.ErrorCodeAll
	for _, s := range source {
		if reflect.DeepEqual(target, s) || reflect.DeepEqual(s, &all) {
			return true
		}
	}
	return false
}

func ContainsErrorMessage(msg string, templates []*string) bool {
	for _, template := range templates {
		if template != nil {
			if strings.Contains(msg, *template) {
				return true
			}
		}
	}
	return false
}
