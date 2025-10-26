/*
Package jsonpat provides a custom JSON unmarshaler for Go.
It extends the standard `encoding/json` package to support unmarshaling
JSON objects with dynamic keys (keys not explicitly defined in the struct)
into maps within the struct.

This is useful for handling JSON payloads where some fields are known,
but others follow a predictable pattern (e.g., prefixes, suffixes) and
should be collected into a map.

# Tag Format

The package introduces the `jsonpat` struct tag.
Its format is: `jsonpat:"<value>,<type>"`

- <value>: The string value to match against the JSON key (e.g., "dyn_", "_suffix").
- <type>:  The matching logic to use. Must be one of:
  - `prefix`: Matches if the JSON key starts with <value>.
  - `contains`: Matches if the JSON key contains <value>.
  - `suffix`: Matches if the JSON key ends with <value>.

Fields using this tag must be of a `map[string]T` type, where `T` is
the type you expect the dynamic value to be.

# Example Usage

Given a struct:

	type MyData struct {
		KnownField      string            `json:"known_field"`
		DynamicByPrefix map[string]int    `jsonpat:"dyn_,prefix"`
		DynamicBySuffix map[string]string `jsonpat:"_id,suffix"`
	}

And JSON data:

	jsonData := []byte(`{
		"known_field": "hello",
		"dyn_abc": 123,
		"dyn_xyz": 456,
		"user_id": "u-1",
		"item_id": "i-9"
	}`)

Unmarshaling:

	var data MyData
	err := jsonpat.UnmarshalJson(jsonData, &data)
	if err != nil {
		// handle error
	}

	// data.KnownField == "hello"
	// data.DynamicByPrefix["dyn_abc"] == 123
	// data.DynamicByPrefix["dyn_xyz"] == 456
	// data.DynamicBySuffix["user_id"] == "u-1"
	// data.DynamicBySuffix["item_id"] == "i-9"
*/
package jsonpat
