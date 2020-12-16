package parser

import (
	"bytes"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
)

type (
	_CommandName string
	_CommandFunc func(args ...string) (interface{}, error)
)

const (
	_Replace _CommandName = "replace"
)

var (
	commands = map[_CommandName]func(pipeValue interface{}, args ...string) (interface{}, error){
		_Replace: func(pipeValue interface{}, args ...string) (interface{}, error) {
			if len(args) != 2 {
				return nil, errInvalidArgument
			}

			switch pipeValue.(type) {
			case string:
				return strings.Replace(pipeValue.(string), args[0], args[1], -1), nil
			case map[string]interface{}:
				value, err := json.Marshal(pipeValue)
				if err != nil {
					return nil, errInvalidArgument
				}
				value = bytes.Replace(value, []byte(args[0]), []byte(args[1]), -1)
				result := map[string]interface{}{}
				err = json.Unmarshal(value, &result)
				if err != nil {
					return nil, errInvalidArgument
				}
				return result, nil
			case nil:
				return nil, nil
			}

			return nil, errInvalidArgument
		},
	}
)

var (
	errCmdNotFound     = errors.New("command not found")
	errInvalidArgument = errors.New("invalid argument")
)

func ParseCmd(key string) (keys []string) {
	r := regexp.MustCompile(`([\w\.]+)|'(.*?)'|"(.*?)"|(\$\{.*?\})`)
	temp := r.FindAllString(key, -1)
	for _, key := range temp {
		if strings.HasPrefix(key, `'`) && strings.HasSuffix(key, `'`) {
			key = key[1 : len(key)-1]
		}
		if strings.HasPrefix(key, `"`) && strings.HasSuffix(key, `"`) {
			key = key[1 : len(key)-1]
		}
		keys = append(keys, key)
	}
	return
}

func Command(commandName string, pipeValue interface{}, args ...string) (interface{}, error) {
	if command, ok := commands[_CommandName(commandName)]; ok {
		return command(pipeValue, args...)
	}
	return nil, errCmdNotFound
}
