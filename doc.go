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
  - `regex`: Matches if the JSON key matches <value> (which must be a valid regex pattern)

Fields using this tag can be one of two kinds:

 1. **Map Type (`map[string]T`):** All JSON keys that match the rule will be
    unmarshaled into this map. The map must be initialized (or nil).

 2. **Scalar Type (e.g., `string`, `int`, `bool`):** The *first* JSON key
    that matches the rule will have its value unmarshaled into this field.
    Any subsequent keys matching the same rule will be ignored for this field.

# Example Usage

Given a struct:

	type MyData struct {
		KnownField      string            `json:"known_field"`
		DynamicByPrefix map[string]int    `jsonpat:"dyn_,prefix"`
		DynamicByRegex  map[string]string `jsonpat:"^user_\\d+$,regex"`
		FirstScalar     string            `jsonpat:"_val,contains"`
	}

And JSON data:

	jsonData := []byte(`{
		"known_field": "hello",
		"dyn_abc": 123,
		"dyn_xyz": 456,
		"user_101": "u-1",
		"user_102": "u-2",
		"other_val": "first",
		"another_val": "second"
	}`)

Unmarshaling:

	var data MyData
	err := jsonpat.Unmarshal(jsonData, &data) //
	if err != nil {
		// handle error
	}

	// data.KnownField == "hello"
	//
	// data.DynamicByPrefix["dyn_abc"] == 123
	// data.DynamicByPrefix["dyn_xyz"] == 456
	//
	// data.DynamicByRegex["user_101"] == "u-1"
	// data.DynamicByRegex["user_102"] == "u-2"
	//
	// data.FirstScalar == "second" (Deterministically selected because "another_val" sorts before "other_val")
*/
package jsonpat
