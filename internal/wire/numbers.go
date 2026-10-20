package wire

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// The API encodes the numbers of the employment declarations the way they are
// written in Greece: the thousands are separated by a dot, the decimals by a
// comma, and the hours carry a single decimal.

// Amount is a monetary amount, encoded with a dot for the thousands and a
// comma for the decimals, such as "1.500,00".
type Amount float64

// MarshalJSON implements the json.Marshaler interface.
//
// A value that is not given is left empty, the way the API leaves it.
func (a Amount) MarshalJSON() ([]byte, error) {
	if a == 0 {
		return json.Marshal("")
	}
	return json.Marshal(formatAmount(float64(a)))
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (a *Amount) UnmarshalJSON(data []byte) error {
	text, err := unmarshalText(data)
	if err != nil {
		return err
	}
	if text == "" {
		*a = 0
		return nil
	}
	value, err := parseGreekNumber(text)
	if err != nil {
		return err
	}
	*a = Amount(value)

	return nil
}

// Hours is a number of hours, encoded with a comma for the decimals, such as
// "38,0".
type Hours float64

// MarshalJSON implements the json.Marshaler interface.
//
// A value that is not given is left empty, the way the API leaves it.
func (h Hours) MarshalJSON() ([]byte, error) {
	if h == 0 {
		return json.Marshal("")
	}
	return json.Marshal(formatHours(float64(h)))
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (h *Hours) UnmarshalJSON(data []byte) error {
	text, err := unmarshalText(data)
	if err != nil {
		return err
	}
	if text == "" {
		*h = 0
		return nil
	}
	value, err := parseGreekNumber(text)
	if err != nil {
		return err
	}
	*h = Hours(value)

	return nil
}

// Tally is a number of three digits, such as "050".
type Tally int

// MarshalJSON implements the json.Marshaler interface.
func (t Tally) MarshalJSON() ([]byte, error) {
	return json.Marshal(fmt.Sprintf("%03d", int(t)))
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (t *Tally) UnmarshalJSON(data []byte) error {
	text, err := unmarshalNumber(data)
	if err != nil {
		return err
	}
	if text == "" {
		*t = 0
		return nil
	}
	value, err := strconv.Atoi(text)
	if err != nil {
		return fmt.Errorf("wire: %q is not a number: %w", text, err)
	}
	*t = Tally(value)

	return nil
}

// Plain is a number the API writes without padding, such as "15".
type Plain int

// MarshalJSON implements the json.Marshaler interface.
func (p Plain) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.Itoa(int(p)))
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (p *Plain) UnmarshalJSON(data []byte) error {
	text, err := unmarshalNumber(data)
	if err != nil {
		return err
	}
	if text == "" {
		*p = 0
		return nil
	}
	value, err := strconv.Atoi(text)
	if err != nil {
		return fmt.Errorf("wire: %q is not a number: %w", text, err)
	}
	*p = Plain(value)

	return nil
}

// unmarshalNumber reads a number the API writes as a string.
func unmarshalNumber(data []byte) (string, error) {
	var text string
	if err := json.Unmarshal(data, &text); err != nil {
		return "", fmt.Errorf("wire: the value is not a number: %w", err)
	}

	return text, nil
}

// Count is a small number, encoded with two digits, such as "02".
type Count int

// MarshalJSON implements the json.Marshaler interface.
func (c Count) MarshalJSON() ([]byte, error) {
	return json.Marshal(fmt.Sprintf("%02d", int(c)))
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (c *Count) UnmarshalJSON(data []byte) error {
	text, err := unmarshalText(data)
	if err != nil {
		return err
	}
	if text == "" {
		*c = 0
		return nil
	}
	value, err := strconv.Atoi(text)
	if err != nil {
		return fmt.Errorf("wire: %q is not a number: %w", text, err)
	}
	*c = Count(value)

	return nil
}

// Minutes is a number of minutes, encoded with three digits, such as "030".
type Minutes int

// MarshalJSON implements the json.Marshaler interface.
func (m Minutes) MarshalJSON() ([]byte, error) {
	return json.Marshal(fmt.Sprintf("%03d", int(m)))
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (m *Minutes) UnmarshalJSON(data []byte) error {
	text, err := unmarshalText(data)
	if err != nil {
		return err
	}
	if text == "" {
		*m = 0
		return nil
	}
	value, err := strconv.Atoi(text)
	if err != nil {
		return fmt.Errorf("wire: %q is not a number: %w", text, err)
	}
	*m = Minutes(value)

	return nil
}

// unmarshalText reads a JSON string.
func unmarshalText(data []byte) (string, error) {
	var text string
	if err := json.Unmarshal(data, &text); err != nil {
		return "", fmt.Errorf("wire: the value is not a string: %w", err)
	}

	return text, nil
}

// parseGreekNumber reads a number the way it is written in Greece: the
// thousands are separated by a dot and the decimals by a comma.
func parseGreekNumber(text string) (float64, error) {
	normalised := strings.ReplaceAll(text, ".", "")
	normalised = strings.ReplaceAll(normalised, ",", ".")

	value, err := strconv.ParseFloat(normalised, 64)
	if err != nil {
		return 0, fmt.Errorf("wire: %q is not a number: %w", text, err)
	}

	return value, nil
}

// formatAmount writes a monetary amount the way the API expects it: the
// thousands separated by a dot, the decimals by a comma, and always two
// decimals.
func formatAmount(value float64) string {
	whole, decimals := splitDecimal(value, 2)
	return groupThousands(whole) + "," + decimals
}

// formatHours writes a number of hours the way the API expects it: the
// thousands separated by a dot, the decimals by a comma, and always one
// decimal.
func formatHours(value float64) string {
	whole, decimals := splitDecimal(value, 1)
	return groupThousands(whole) + "," + decimals
}

// splitDecimal splits a number into its integer part and its decimals.
func splitDecimal(value float64, digits int) (string, string) {
	text := strconv.FormatFloat(value, 'f', digits, 64)
	parts := strings.SplitN(text, ".", 2)
	decimals := "0"
	if len(parts) == 2 {
		decimals = parts[1]
	}
	return parts[0], decimals
}

// groupThousands separates the thousands of an integer part with a dot.
func groupThousands(whole string) string {
	sign := ""
	if strings.HasPrefix(whole, "-") {
		sign = "-"
		whole = whole[1:]
	}
	if len(whole) <= 3 {
		return sign + whole
	}

	var groups []string
	for len(whole) > 3 {
		groups = append([]string{whole[len(whole)-3:]}, groups...)
		whole = whole[:len(whole)-3]
	}
	groups = append([]string{whole}, groups...)

	return sign + strings.Join(groups, ".")
}
