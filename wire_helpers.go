package ergani

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/takispanag/ergani-go-sdk/internal/wire"
)

// The helpers of the wire format.

// declarationFields turns a payload into the map of its fields, so that the
// fields a declaration type does not carry can be left out.
func declarationFields(payload any) (map[string]any, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("ergani: failed to marshal the declaration: %w", err)
	}

	fields := map[string]any{}
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil, fmt.Errorf("ergani: failed to decode the declaration: %w", err)
	}

	return fields, nil
}

// dropFields removes the named fields from a payload.
func dropFields(fields map[string]any, names []string) {
	for _, name := range names {
		delete(fields, name)
	}
}

// boolPtr turns the API encoding of an optional boolean into a pointer.
//
// The API leaves a boolean empty when it is not answered, so the three states
// are kept apart.
func boolPtr(value wire.Bool) *bool {
	if value == "" {
		return nil
	}

	answer := value == wire.True

	return &answer
}

// base64Bytes decodes a document of the API.
//
// A document that cannot be decoded yields no bytes rather than an error: the
// API has published a reference payload with a document that is not valid
// Base64, and a client must not fail on a document it only passes through.
func base64Bytes(value wire.Base64) []byte {
	decoded, err := base64.StdEncoding.DecodeString(string(value))
	if err != nil {
		return nil
	}

	return decoded
}
