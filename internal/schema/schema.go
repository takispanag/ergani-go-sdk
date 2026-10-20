// Package schema holds the official schemas of the Ergani API and checks a
// payload against them.
//
// The XSD schemas are the ones the Greek Ministry of Labour publishes in
// the sample archive of the interoperability guide. They describe the fields
// of every declaration type, which field is mandatory, and how the fields
// are nested.
//
// The schemas belong to the Ministry and are reproduced here so that a
// payload can be checked without a network call.
package schema

import (
	"embed"
	"encoding/xml"
	"fmt"
	"path"
	"sort"
	"strings"
	"sync"
)

//go:embed xsds/*.xsd
var files embed.FS

// Field is a field of a declaration, as the schema declares it.
type Field struct {
	// Name is the name of the field, such as "f_afm".
	Name string

	// Required reports whether the declaration cannot do without the field.
	Required bool
}

// Schema is the set of the fields of a declaration type.
type Schema struct {
	// Name is the name of the schema, such as "Card_v1".
	Name string

	// Fields are the fields the schema declares.
	Fields []Field
}

var (
	once    sync.Once
	schemas map[string]*Schema
)

// Names returns the names of the schemas that are embedded, in order.
func Names() []string {
	once.Do(load)

	names := make([]string, 0, len(schemas))
	for name := range schemas {
		names = append(names, name)
	}
	sort.Strings(names)

	return names
}

// Get returns the schema of a declaration type.
//
// The name is the name of the file of the schema, without its extension.
func Get(name string) (*Schema, bool) {
	once.Do(load)

	schema, ok := schemas[name]

	return schema, ok
}

// load reads the embedded schemas.
func load() {
	schemas = map[string]*Schema{}

	entries, err := files.ReadDir("xsds")
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".xsd") {
			continue
		}

		data, err := files.ReadFile(path.Join("xsds", entry.Name()))
		if err != nil {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".xsd")
		schemas[name] = parse(name, data)
	}
}

// parse reads the fields of a schema.
func parse(name string, data []byte) *Schema {
	document := &struct {
		Elements []struct {
			Name      string `xml:"name,attr"`
			MinOccurs string `xml:"minOccurs,attr"`
		} `xml:"element"`
	}{}

	if err := xml.Unmarshal(data, document); err != nil {
		return &Schema{Name: name}
	}

	schema := &Schema{Name: name}
	seen := map[string]bool{}

	for _, element := range document.Elements {
		if !strings.HasPrefix(element.Name, "f_") || seen[element.Name] {
			continue
		}
		seen[element.Name] = true
		schema.Fields = append(schema.Fields, Field{
			Name:     element.Name,
			Required: element.MinOccurs != "0",
		})
	}

	sort.Slice(schema.Fields, func(i, j int) bool {
		return schema.Fields[i].Name < schema.Fields[j].Name
	})

	return schema
}

// Check reports the fields of a payload that the schema does not declare,
// and the fields the schema declares as mandatory and the payload does not
// carry.
//
// The payload is the flat set of the fields of a declaration.
func (s *Schema) Check(fields map[string]any) []error {
	if s == nil {
		return nil
	}

	declared := map[string]bool{}
	required := map[string]bool{}
	for _, field := range s.Fields {
		declared[field.Name] = true
		if field.Required {
			required[field.Name] = true
		}
	}

	var problems []error

	for name := range fields {
		if !declared[name] {
			problems = append(problems, fmt.Errorf("the field %q is not declared by the schema %s", name, s.Name))
		}
	}

	for name := range required {
		if _, ok := fields[name]; !ok {
			problems = append(problems, fmt.Errorf("the field %q is required by the schema %s", name, s.Name))
		}
	}

	sort.Slice(problems, func(i, j int) bool {
		return problems[i].Error() < problems[j].Error()
	})

	return problems
}
