package parser

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func ReplaceVariables(config []byte, variables interface{}) ([]byte, error) {
	r, err := regexp.Compile(`\$\{(.*?)\}`)
	if err != nil {
		return config, err
	}

	for _, match := range r.FindAllStringSubmatch(string(config), -1) {
		var (
			v     interface{}
			value string
			err   error
		)

		// find variable
		matchKeys := strings.Split(match[1], "||")
		for _, matchKey := range matchKeys {
			pipes := strings.Split(matchKey, "|")
			for _, pipe := range pipes {
				cmds := ParseCmd(strings.TrimSpace(pipe))
				v, err = Command(cmds[0], v, cmds[1:]...)

				if err != nil && errors.Is(err, errCmdNotFound) {
					v, err = getJSONValue(strings.Split(strings.TrimSpace(cmds[0]), "."), variables)
				} else {
					break
				}
			}

			if err == nil && v != nil {
				break
			}
		}
		if err != nil {
			return config, err
		}
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

func GetJSONValue(cmd string, variables interface{}) (interface{}, error) {
	if cmd == "" {
		return nil, nil
	}
	if ok, err := regexp.MatchString(`^\$\{(.*?)\}`, cmd); !ok || err != nil {
		return nil, errors.New("invalid cmd format")
	}
	path := strings.Split(cmd[2:len(cmd)-1], ".")
	return getJSONValue(path, variables)
}

func getJSONValue(path []string, variables interface{}) (interface{}, error) {
	if len(path) <= 0 {
		return variables, nil
	}

	if variables == nil {
		return nil, nil
	}

	switch variables.(type) {
	case map[string]interface{}:
		variables = (variables.(map[string]interface{}))[path[0]]
		return getJSONValue(path[1:], variables)
	}

	return variables, errors.New("invalid path")
}

func SetJSONValue(cmd string, value, variables interface{}) (interface{}, error) {
	if cmd == "" {
		return nil, nil
	}
	if ok, err := regexp.MatchString(`^\$\{(.*?)\}`, cmd); !ok || err != nil {
		return nil, errors.New("invalid cmd format")
	}
	path := strings.Split(cmd[2:len(cmd)-1], ".")
	return setJSONValue(path, value, variables), nil
}

func setJSONValue(path []string, value, variables interface{}) interface{} {
	if len(path) == 0 {
		return nil
	}

	currentMap := make(map[string]interface{})
	switch variables.(type) {
	case map[string]interface{}:
		currentMap = (variables.(map[string]interface{}))
	}

	if len(path) <= 1 {
		currentMap[path[0]] = value
		return currentMap
	}

	currentMap[path[0]] = setJSONValue(path[1:], value, currentMap[path[0]])
	return currentMap
}
