package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

// The fields of a declaration are written in a payload in the order the
// Ministry writes them, and the API names the inner object of a nested
// object after the declaration type. Both need the order of the keys of a
// JSON object, which the standard library does not keep, so the documents
// are read into a value of their own.

// jsonKind tells the kind of a value of a JSON document.
type jsonKind int

const (
	jsonNull jsonKind = iota
	jsonBool
	jsonNumber
	jsonString
	jsonList
	jsonObject
)

// jsonMember is a key of an object and the value it holds.
type jsonMember struct {
	name  string
	value *jsonValue
}

// jsonValue is a value of a JSON document that keeps the order of the keys
// of an object.
type jsonValue struct {
	kind    jsonKind
	text    string
	number  float64
	boolean bool
	list    []*jsonValue
	members []jsonMember
}

// parseJSON reads a JSON document.
func parseJSON(data []byte) (*jsonValue, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()

	value, err := parseJSONValue(decoder)
	if err != nil {
		return nil, err
	}

	return value, nil
}

// parseJSONValue reads the value that comes next in a JSON document.
func parseJSONValue(decoder *json.Decoder) (*jsonValue, error) {
	token, err := decoder.Token()
	if err != nil {
		return nil, fmt.Errorf("ergani: the JSON document could not be read: %w", err)
	}

	switch value := token.(type) {
	case json.Delim:
		switch value {
		case '{':
			return parseJSONObject(decoder)
		case '[':
			return parseJSONList(decoder)
		default:
			return nil, fmt.Errorf("ergani: unexpected %v in a JSON document", value)
		}
	case string:
		return &jsonValue{kind: jsonString, text: value}, nil
	case json.Number:
		number, err := value.Float64()
		if err != nil {
			return nil, fmt.Errorf("ergani: the number of a JSON document could not be read: %w", err)
		}

		return &jsonValue{kind: jsonNumber, number: number, text: value.String()}, nil
	case bool:
		return &jsonValue{kind: jsonBool, boolean: value}, nil
	case nil:
		return &jsonValue{kind: jsonNull}, nil
	default:
		return nil, fmt.Errorf("ergani: unexpected token in a JSON document")
	}
}

// parseJSONObject reads an object of a JSON document.
func parseJSONObject(decoder *json.Decoder) (*jsonValue, error) {
	object := &jsonValue{kind: jsonObject}

	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return nil, fmt.Errorf("ergani: the JSON document could not be read: %w", err)
		}

		name, ok := token.(string)
		if !ok {
			return nil, fmt.Errorf("ergani: the key of a JSON object is not a string")
		}

		value, err := parseJSONValue(decoder)
		if err != nil {
			return nil, err
		}

		object.members = append(object.members, jsonMember{name: name, value: value})
	}

	if _, err := decoder.Token(); err != nil {
		return nil, fmt.Errorf("ergani: the JSON document could not be read: %w", err)
	}

	return object, nil
}

// parseJSONList reads a list of a JSON document.
func parseJSONList(decoder *json.Decoder) (*jsonValue, error) {
	list := &jsonValue{kind: jsonList}

	for decoder.More() {
		value, err := parseJSONValue(decoder)
		if err != nil {
			return nil, err
		}

		list.list = append(list.list, value)
	}

	if _, err := decoder.Token(); err != nil {
		return nil, fmt.Errorf("ergani: the JSON document could not be read: %w", err)
	}

	return list, nil
}

// isObject tells whether a value is an object.
func (v *jsonValue) isObject() bool {
	return v != nil && v.kind == jsonObject
}

// isList tells whether a value is a list.
func (v *jsonValue) isList() bool {
	return v != nil && v.kind == jsonList
}

// isEmpty tells whether an object holds nothing.
func (v *jsonValue) isEmpty() bool {
	return !v.isObject() || len(v.members) == 0
}

// firstKey returns the first key of an object.
func (v *jsonValue) firstKey() string {
	if !v.isObject() || len(v.members) == 0 {
		return ""
	}

	return v.members[0].name
}

// get returns the value of a key of an object.
func (v *jsonValue) get(name string) *jsonValue {
	if !v.isObject() {
		return nil
	}

	for _, member := range v.members {
		if member.name == name {
			return member.value
		}
	}

	return nil
}

// text returns the text of a value that holds one.
func (v *jsonValue) asText() (string, bool) {
	if v == nil || v.kind != jsonString {
		return "", false
	}

	return v.text, true
}

// readJSONFile reads a JSON document of a file.
func readJSONFile(path string) (*jsonValue, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("ergani: the file could not be read: %w", err)
	}

	return parseJSON(data)
}
