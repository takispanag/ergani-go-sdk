package ergani

import (
	"github.com/takispanag/ergani-go-sdk/internal/schema"
)

// Schemas returns the names of the official schemas of the API that the SDK
// embeds.
//
// The schemas are the XSD files the Greek Ministry of Labour publishes in
// the sample archive of the interoperability guide. They describe the fields
// of every declaration type.
func Schemas() []string {
	return schema.Names()
}

// CheckSchema checks the fields of a declaration against the official schema
// of its type.
//
// It reports the fields the schema does not declare, and the fields the
// schema declares as mandatory and the declaration does not carry. The
// schema is the one the SDK embeds, named after the file of the schema, such
// as "Card_v1".
func CheckSchema(schemaName string, fields map[string]any) []error {
	definition, ok := schema.Get(schemaName)
	if !ok {
		return []error{&ValidationError{
			Field:  "schema",
			Value:  schemaName,
			Reason: "the SDK does not embed a schema with that name",
		}}
	}

	return definition.Check(fields)
}
