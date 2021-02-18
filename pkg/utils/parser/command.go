package parser

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
	"strings"
)

type (
	_CommandName string
	_CommandFunc func(args ...string) (interface{}, error)
)

const (
	_Replace      _CommandName = "replace"
	_Base64Encode _CommandName = "base64encode"
	_Base64Decode _CommandName = "base64decode"
	_Quote        _CommandName = "quote"
	_UnQuote      _CommandName = "unquote"
)

var (
	_ReplaceCommand = func(pipeValue interface{}, args ...string) (interface{}, error) {
		if len(args) < 1 || len(args) > 2 {
			return nil, errInvalidArgument
		}

		oriValue := args[0]
		newValue := ""
		if len(args) == 2 {
			newValue = args[1]
		}

		switch pipeValue.(type) {
		case string:
			return strings.Replace(pipeValue.(string), oriValue, newValue, -1), nil
		case map[string]interface{}:
			value, err := json.Marshal(pipeValue)
			if err != nil {
				return nil, errInvalidArgument
			}
			value = bytes.Replace(value, []byte(oriValue), []byte(newValue), -1)
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
	}

	_Base64EncodeCommand = func(pipeValue interface{}, args ...string) (interface{}, error) {
		switch pipeValue.(type) {
		case string:
			return b64.StdEncoding.EncodeToString([]byte(pipeValue.(string))), nil
		case map[string]interface{}:
			value, err := json.Marshal(pipeValue)
			if err != nil {
				return nil, errInvalidArgument
			}
			return b64.StdEncoding.EncodeToString(value), nil
		case nil:
			return nil, nil
		}
		return nil, errInvalidArgument
	}

	_Base64DecodeCommand = func(pipeValue interface{}, args ...string) (interface{}, error) {
		switch pipeValue.(type) {
		case string:
			decodeStr, err := b64.StdEncoding.DecodeString(pipeValue.(string))
			if err != nil {
				return nil, errInvalidArgument
			}
			return string(decodeStr), nil
		case nil:
			return nil, nil
		}
		return nil, errInvalidArgument
	}

	_QuoteCommand = func(pipeValue interface{}, args ...string) (interface{}, error) {
		switch pipeValue.(type) {
		case string:
			v := strconv.Quote(pipeValue.(string))
			return v[1 : len(v)-1], nil
		case nil:
			return nil, nil
		}
		return nil, errInvalidArgument
	}

	_UnQuoteCommand = func(pipeValue interface{}, args ...string) (interface{}, error) {
		switch pipeValue.(type) {
		case string:
			v, err := strconv.Unquote(pipeValue.(string))
			if err != nil {
				return nil, errInvalidArgument
			}
			return v, nil
		case nil:
			return nil, nil
		}
		return nil, errInvalidArgument
	}
)

var (
	commands = map[_CommandName]func(pipeValue interface{}, args ...string) (interface{}, error){
		_Replace:      _ReplaceCommand,
		_Base64Encode: _Base64EncodeCommand,
		_Base64Decode: _Base64DecodeCommand,
		_Quote:        _QuoteCommand,
		_UnQuote:      _UnQuoteCommand,
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
		} else if strings.HasPrefix(key, `"`) && strings.HasSuffix(key, `"`) {
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
