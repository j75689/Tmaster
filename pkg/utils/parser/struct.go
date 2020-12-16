package parser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"

	"github.com/thoas/go-funk"
)

func ReplaceSystemVariables(config []byte, systemVariables interface{}) ([]byte, error) {
	r, err := regexp.Compile(`\#\{(.*?)\}`)
	if err != nil {
		return config, err
	}
	for _, match := range r.FindAllStringSubmatch(string(config), -1) {
		var (
			value string
		)
		// find variable
		v := funk.Get(systemVariables, match[1])
		if v != nil {
			nValue := ""
			vv, err := json.Marshal(v)
			if err != nil {
				nValue = strconv.Quote(fmt.Sprint(v))
			} else {
				nValue = strconv.Quote(string(vv))
			}
			value = nValue[1 : len(nValue)-1]
		} else {
			value = `\"` + match[1] + `\"`
		}
		// replace
		config = bytes.Replace(config, []byte(match[0]), []byte(value), -1)
	}
	return config, nil
}
