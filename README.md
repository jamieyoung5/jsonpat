# jsonpat

`jsonpat` extends the standard `encoding/json` package to support unmarshaling JSON objects with dynamic keys into maps within a struct.

This is useful when you have a JSON payload where some fields are known, but others are dynamic and follow a predictable pattern (e.g., a common prefix, suffix, or containing substring). `jsonpat` maps these dynamic keys to a `map[string]T` field in your struct.

## Features

- Unmarshal unknown fields into maps based on filters.
- Supports matching dynamic keys by:
  - `prefix`
  - `contains`
  - `suffix`

## WIP Features

- Regex matching support
- QOL changes
- Different key types

## Installation

```sh
go get [github.com/jamieyoung5/jsonpat](https://github.com/jamieyoung5/jsonpat)
```

## Usage

Define your struct using both standard `json` tags and the `jsonpat` tag.

The `jsonpat` tag format is:
**`jsonpat:"<value>,<type>"`**

-   **`<value>`**: The string value to match against the JSON key.
-   **`<type>`**: The matching logic. Must be one of `prefix`, `contains`, or `suffix`.

### Example

Here is the struct definition from the library's tests:

```go
import "[github.com/jamieyoung5/jsonpat](https://github.com/jamieyoung5/jsonpat)"

// EmbeddedStruct demonstrates support for embedded structs.
type EmbeddedStruct struct {
	EmbeddedField string                 `json:"embedded_field"`
	DynamicSuffix map[string]interface{} `jsonpat:"_suffix,suffix"`
}

type TestStruct struct {
	EmbeddedStruct
	KnownField      string             `json:"known_field"`
	OtherKnown      int                `json:"other"`
	Ignored         string             `json:"-"`
	DynamicPrefix   map[string]int     `jsonpat:"dyn_,prefix"`
	DynamicContains map[string]float64 `jsonpat:"_val_,contains"`
}
```

Now, let's unmarshal some JSON:

```go
package main

import (
	"fmt"
	"log"

	"[github.com/jamieyoung5/jsonpat](https://github.com/jamieyoung5/jsonpat)"
)

// (Struct definitions from above)

func main() {
	jsonData := []byte(`{
		"known_field": "hello",
		"other": 123,
		"ignored": "should not be Loaded",
		"embedded_field": "i am embedded",
		"dyn_abc": 1,
		"dyn_xyz": 2,
		"field_val_1": 10.5,
		"field_val_2": 20.75,
		"some_suffix": "test string",
		"another_suffix": true,
		"not_matching": "skip me"
	}`)

	var result TestStruct
	err := jsonpat.UnmarshalJson(jsonData, &result)
	if err != nil {
		log.Fatalf("Failed to unmarshal: %v", err)
	}

	// --- Known Fields ---
	fmt.Printf("KnownField:     %s\n", result.KnownField)
	fmt.Printf("OtherKnown:     %d\n", result.OtherKnown)
	fmt.Printf("Ignored:        '%s' (should be empty)\n", result.Ignored)

	// --- Embedded Known Field ---
	fmt.Printf("EmbeddedField:  %s\n", result.EmbeddedField)

	// --- Dynamic Fields ---
	fmt.Println("\n--- DynamicPrefix (dyn_,prefix) ---")
	for k, v := range result.DynamicPrefix {
		fmt.Printf("  %s: %d\n", k, v)
	}

	fmt.Println("\n--- DynamicContains (_val_,contains) ---")
	for k, v := range result.DynamicContains {
		fmt.Printf("  %s: %f\n", k, v)
	}

	fmt.Println("\n--- DynamicSuffix (_suffix,suffix) ---")
	for k, v := range result.DynamicSuffix {
		fmt.Printf("  %s: %v\n", k, v)
	}
}
```

### Output:

```
KnownField:     hello
OtherKnown:     123
Ignored:        '' (should be empty)
EmbeddedField:  i am embedded

--- DynamicPrefix (dyn_,prefix) ---
  dyn_abc: 1
  dyn_xyz: 2

--- DynamicContains (_val_,contains) ---
  field_val_1: 10.500000
  field_val_2: 20.750000

--- DynamicSuffix (_suffix,suffix) ---
  some_suffix: test string
  another_suffix: true
```