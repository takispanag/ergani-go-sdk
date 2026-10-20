package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// The schemas of the Ministry.
//
// A schema declares the format of every field, which is what tells the type
// of the field in Go: a date, a time, a number of the Greek format, a
// count, a flag, a document or a plain string. The schemas also carry the
// official Greek description of every field, which becomes the comment of
// the field in the public API.

// xsdNamespace is the namespace of the schemas of XML Schema.
const xsdNamespace = "http://www.w3.org/2001/XMLSchema"

// The patterns of the API.
var (
	// greekAmount matches the number of the API, whose thousands are
	// separated by a dot and whose decimals are separated by a comma.
	greekAmount = regexp.MustCompile(`\(,\[0-9\]\[0-9\]\)$`)

	// greekHours matches the hours of the API, whose decimals are separated
	// by a comma.
	greekHours = regexp.MustCompile(`\(,\[0-9\]\)$`)

	// datePattern matches the date of the API.
	datePattern = regexp.MustCompile(`\[/\]\)\(\[1-2\]\)\(\[0-9\]\)\(\[0-9\]\)\(\[0-9\]\)`)

	// timePattern matches the time of the API.
	timePattern = regexp.MustCompile(`\(\[0-1\]\[0-9\]\|2\[0-3\]\):\(\[0-5\]\[0-9\]\)`)

	// oneOrTwoDigits matches a number of one or two digits.
	oneOrTwoDigits = regexp.MustCompile(`^\[0-9\]\{1,2\}$`)

	// oneToThreeDigits matches a number of one to three digits.
	oneToThreeDigits = regexp.MustCompile(`^\[0-9\]\{1,3\}$`)

	// anyDigits matches a number of any number of digits.
	anyDigits = regexp.MustCompile(`^\[0-9\]\+$`)
)

// schemaField is a field of a schema of the Ministry: its name, the kind of
// value it holds and the official description of it.
type schemaField struct {
	name string
	kind string
	doc  string
}

// schemaTypes holds the kind of every field, per schema.
type schemaTypes map[string]map[string]string

// fieldDocs holds the official description of every field.
type fieldDocs map[string]string

// readSchemas reads the type and the description of every field of every
// schema of a directory.
//
// The schemas are read in the order of their name, and the description of a
// field is the one of the first schema that carries it.
func readSchemas(dir string) (schemaTypes, fieldDocs, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "*.xsd"))
	if err != nil {
		return nil, nil, fmt.Errorf("ergani: the schemas could not be listed: %w", err)
	}
	sort.Strings(paths)

	types := schemaTypes{}
	docs := fieldDocs{}

	for _, path := range paths {
		fields, err := readSchema(path)
		if err != nil {
			return nil, nil, err
		}

		name := strings.TrimSuffix(filepath.Base(path), ".xsd")
		for _, field := range fields {
			if _, ok := types[name]; !ok {
				types[name] = map[string]string{}
			}
			types[name][field.name] = field.kind

			if field.doc != "" {
				if _, ok := docs[field.name]; !ok {
					docs[field.name] = field.doc
				}
			}
		}
	}

	return types, docs, nil
}

// elementState is the state of an element of a schema that is being read.
type elementState struct {
	name           string
	depth          int
	hasType        bool
	hasRestriction bool
	base           string
	pattern        string
	enums          []string
	doc            string
	hasDoc         bool
}

// readSchema reads the fields of a schema of the Ministry.
func readSchema(path string) ([]schemaField, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("ergani: the schema %s could not be read: %w", path, err)
	}

	decoder := xml.NewDecoder(bytes.NewReader(data))

	var (
		fields   []schemaField
		stack    []elementState
		depth    int
		inDoc    bool
		docText  strings.Builder
		docChild bool
		waiters  []int
	)

	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF { //nolint:errorlint // the decoder reports the end of the document.
				break
			}

			return nil, fmt.Errorf("ergani: the schema %s could not be read: %w", path, err)
		}

		switch value := token.(type) {
		case xml.StartElement:
			depth++

			switch {
			case isSchemaName(value.Name, "element"):
				state := elementState{name: attribute(value, "name"), depth: depth}
				stack = append(stack, state)
			case len(stack) > 0:
				top := &stack[len(stack)-1]
				switch {
				case depth == top.depth+1 && isSchemaName(value.Name, "simpleType"):
					top.hasType = true
				case depth == top.depth+2 && isSchemaName(value.Name, "restriction"):
					top.hasRestriction = true
					top.base = attribute(value, "base")
				case depth == top.depth+3 && isSchemaName(value.Name, "pattern"):
					top.pattern = attribute(value, "value")
				case depth == top.depth+3 && isSchemaName(value.Name, "enumeration"):
					top.enums = append(top.enums, attribute(value, "value"))
				}
			}

			if isSchemaName(value.Name, "documentation") && len(stack) > 0 {
				waiters = waiters[:0]
				for index := range stack {
					if !stack[index].hasDoc {
						waiters = append(waiters, index)
					}
				}
				inDoc = true
				docChild = false
				docText.Reset()
			} else if inDoc {
				docChild = true
			}
		case xml.CharData:
			if inDoc && !docChild {
				docText.Write(value)
			}
		case xml.EndElement:
			if isSchemaName(value.Name, "documentation") && inDoc {
				text := normaliseSpace(docText.String())
				for _, index := range waiters {
					stack[index].doc = text
					stack[index].hasDoc = true
				}
				inDoc = false
			}

			if isSchemaName(value.Name, "element") && len(stack) > 0 {
				state := stack[len(stack)-1]
				stack = stack[:len(stack)-1]

				if strings.HasPrefix(state.name, "f_") {
					fields = append(fields, schemaField{
						name: state.name,
						kind: schemaKind(state),
						doc:  state.doc,
					})
				}
			}

			depth--
		}
	}

	return fields, nil
}

// schemaKind tells the kind of a field of a schema.
func schemaKind(state elementState) string {
	if !state.hasType || !state.hasRestriction {
		return "string"
	}

	switch localName(state.base) {
	case "base64Binary":
		return "base64"
	case "date":
		return "date"
	case "dateTime":
		return "datetime"
	}

	if len(state.enums) > 0 {
		if onlyZeroAndOne(state.enums) {
			return "flag"
		}

		return "enum"
	}

	switch {
	case greekAmount.MatchString(state.pattern):
		return "amount"
	case greekHours.MatchString(state.pattern):
		return "hours"
	case timePattern.MatchString(state.pattern):
		return "time"
	case datePattern.MatchString(state.pattern):
		return "date"
	case oneOrTwoDigits.MatchString(state.pattern):
		return "count"
	case oneToThreeDigits.MatchString(state.pattern):
		return "tally"
	case anyDigits.MatchString(state.pattern):
		return "number"
	default:
		return "string"
	}
}

// onlyZeroAndOne tells whether the enumerated values are the two states of
// a flag.
func onlyZeroAndOne(values []string) bool {
	for _, value := range values {
		if value != "0" && value != "1" {
			return false
		}
	}

	return true
}

// isSchemaName tells whether an element of a schema is the one that is
// wanted.
func isSchemaName(name xml.Name, local string) bool {
	return name.Local == local && (name.Space == xsdNamespace || name.Space == "")
}

// attribute returns the value of an attribute of an element.
func attribute(element xml.StartElement, name string) string {
	for _, attribute := range element.Attr {
		if attribute.Name.Local == name {
			return attribute.Value
		}
	}

	return ""
}

// localName returns the name of a type without the prefix of its namespace.
func localName(name string) string {
	if index := strings.LastIndex(name, ":"); index >= 0 {
		return name[index+1:]
	}

	return name
}

// normaliseSpace turns the runs of white space of a text into single
// spaces.
func normaliseSpace(text string) string {
	return strings.Join(strings.Fields(text), " ")
}
