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
	"encoding/base64"
	"encoding/json"
	"fmt"
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

// UnmarshalJSON implements the json.Unmarshaler interface.
func (d *Date) UnmarshalJSON(data []byte) error {
	var text string
	if err := json.Unmarshal(data, &text); err != nil {
		return err
	}
	if text == "" {
		d.Time = time.Time{}
		return nil
	}
	// The API writes a day and a month with and without their leading zero.
	parsed, err := time.Parse("2/1/2006", text)
	if err != nil {
		return fmt.Errorf("wire: %q is not a date of the form DD/MM/YYYY: %w", text, err)
	}
	d.Time = parsed

	return nil
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

// UnmarshalJSON implements the json.Unmarshaler interface.
func (t *Time) UnmarshalJSON(data []byte) error {
	var text string
	if err := json.Unmarshal(data, &text); err != nil {
		return err
	}
	if text == "" {
		t.Time = time.Time{}
		return nil
	}
	parsed, err := time.Parse("15:04", text)
	if err != nil {
		return fmt.Errorf("wire: %q is not a time of the form HH:MM: %w", text, err)
	}
	t.Time = parsed

	return nil
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

// Base64 is a document that the API carries Base64 encoded.
//
// The decoder is lenient on purpose: one of the reference payloads of the
// Ministry carries a document that is not valid Base64, and a client must
// not fail on a document it only has to pass through.
type Base64 string

// MarshalJSON implements the json.Marshaler interface.
func (b Base64) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(b))
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (b *Base64) UnmarshalJSON(data []byte) error {
	var text string
	if err := json.Unmarshal(data, &text); err != nil {
		return err
	}
	*b = Base64(text)

	return nil
}

// Base64FromBytes encodes a document the way the API expects it.
func Base64FromBytes(data []byte) Base64 {
	return Base64(base64.StdEncoding.EncodeToString(data))
}

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

// FromBoolPtr returns the API encoding of an optional boolean.
//
// The API leaves a boolean empty when it is not answered, so the three states
// are kept apart: a nil pointer is the empty string, a pointer to false is
// "0" and a pointer to true is "1".
func FromBoolPtr(b *bool) Bool {
	if b == nil {
		return ""
	}
	return FromBool(*b)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (b *Bool) UnmarshalJSON(data []byte) error {
	var text string
	if err := json.Unmarshal(data, &text); err != nil {
		return err
	}
	switch text {
	case "", "0", "1":
		*b = Bool(text)
		return nil
	default:
		return fmt.Errorf("wire: %q is not a boolean of the API, which uses \"0\" and \"1\"", text)
	}
}
