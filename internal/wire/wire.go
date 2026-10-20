// Package wire holds the JSON representation of the Ergani payloads.
//
// The types in this package are the only place where the field names of the
// API are written down. They are unexported to the rest of the SDK on
// purpose: the public models of the ergani package carry no JSON tags, so the
// wire format can be corrected or extended without breaking the public API.
//
// The field names, the nesting and the value formats follow the official XSD
// schemas of the Greek Ministry of Labour, as published in the sample archive
// of the interoperability guide.
package wire

import (
	"encoding/json"
	"time"
)

// Date is a calendar date in the DD/MM/YYYY format.
//
// The working time organisation declarations, the E8 overtime declaration and
// the E3 family use this format. The zero time is encoded as an empty
// string, which the schemas accept for the optional fields.
type Date struct{ time.Time }

// MarshalJSON implements the json.Marshaler interface.
func (d Date) MarshalJSON() ([]byte, error) {
	if d.IsZero() {
		return json.Marshal("")
	}
	return json.Marshal(d.Format("02/01/2006"))
}

// ISODate is a calendar date in the YYYY-MM-DD format.
//
// The work card declaration uses this format for its reference date. The
// schema declares it as an xs:date.
type ISODate struct{ time.Time }

// MarshalJSON implements the json.Marshaler interface.
func (d ISODate) MarshalJSON() ([]byte, error) {
	if d.IsZero() {
		return json.Marshal("")
	}
	return json.Marshal(d.Format("2006-01-02"))
}

// Time is a wall clock time in the HH:MM format.
type Time struct{ time.Time }

// MarshalJSON implements the json.Marshaler interface.
func (t Time) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return json.Marshal("")
	}
	return json.Marshal(t.Format("15:04"))
}

// DateTime is a timestamp in ISO 8601, with the offset of its location.
//
// The work card declaration uses this format for the moment of the movement.
// The schema declares it as an xs:dateTime.
type DateTime struct{ time.Time }

// MarshalJSON implements the json.Marshaler interface.
func (t DateTime) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return json.Marshal("")
	}
	return json.Marshal(t.Format("2006-01-02T15:04:05.999999999Z07:00"))
}

// Day is a day of the week, from 0 for Sunday to 6 for Saturday.
//
// The schema declares it as a string, not as a number.
type Day string

// Bool is a boolean that the API encodes as "0" and "1".
type Bool string

// False and True are the two values of a Bool.
const (
	False Bool = "0"
	True  Bool = "1"
)

// FromBool returns the API encoding of a boolean.
func FromBool(b bool) Bool {
	if b {
		return True
	}
	return False
}
