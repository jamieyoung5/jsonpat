<div align="center">
    <h1>jsonpat</h1>

[![Go Reference](https://pkg.go.dev/badge/github.com/jamieyoung5/jsonpat.svg)](https://pkg.go.dev/github.com/jamieyoung5/jsonpat)
[![CI](https://github.com/jamieyoung5/jsonpat/actions/workflows/ci.yml/badge.svg)](https://github.com/jamieyoung5/jsonpat/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/jamieyoung5/jsonpat/graph/badge.svg)](https://codecov.io/gh/jamieyoung5/jsonpat)
[![Go Report Card](https://goreportcard.com/badge/github.com/jamieyoung5/jsonpat)](https://goreportcard.com/report/github.com/jamieyoung5/jsonpat)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

</div>

`jsonpat` extends the standard `encoding/json` package to support unmarshaling JSON objects with dynamic keys into maps within a struct.

This is useful when you have a JSON payload where some fields are known, but others are dynamic and follow a predictable pattern (e.g., a common prefix, suffix, or containing substring). `jsonpat` maps these dynamic keys to a `map[string]T` field in your struct.

## Features

- Unmarshal unknown fields into maps based on filters.
- Supports both map fields (`map[string]T`) to capture all matching keys and scalar fields (e.g., `string`, `int`) to capture the first matching key.
- Supports matching dynamic keys by:
    - `prefix`
    - `contains`
    - `suffix`
    - `regex`
- Works alongside standard `json` tags and supports embedded structs

## Installation

```sh
go get github.com/jamieyoung5/jsonpat
```

## Usage

Define your struct using both standard `json` tags and the `jsonpat` tag.

The `jsonpat` tag format is:
**`jsonpat:"<value>,<type>"`**

- **`<value>`**: The string value to match (e.g, a prefix, a substring, suffix, or regex pattern).
- **`<type>`**: The matching logic. Must be one of `prefix`, `contains`, `suffix`, or `regex`.

### Field Types

- **Map Fields (`map[string]T`):** All JSON keys matching the rule will be unmarshaled into this map.

- **Scalar Fields (e.g., `string`, `int`, `bool`):** The value of the first JSON key that matches the rule will be unmarshaled into this field. Subsequent matches for the same rule are ignored.

### Example

Here is a struct definition demonstrating various features:

```go
import "github.com/jamieyoung5/jsonpat"

// EmbeddedStruct demonstrates support for embedded structs.
type EmbeddedStruct struct {
    EmbeddedField string                 `json:"embedded_field"`
    DynamicSuffix map[string]interface{} `jsonpat:"_suffix,suffix"`
}

type TestStruct struct {
    EmbeddedStruct
    KnownField string `json:"known_field"`
    OtherKnown int    `json:"other"`
    Ignored    string `json:"-"`

    // Dynamic Map Fields (Collect all matches)
    DynamicPrefix   map[string]int     `jsonpat:"dyn_,prefix"`
    DynamicContains map[string]float64 `jsonpat:"_val_,contains"`
    DynamicRegex    map[string]string  `jsonpat:"^re_.*$,regex"`

    // Dynamic Scalar Fields (Collect first match)
    ScalarPrefix   string `jsonpat:"scalar_pfx_,prefix"`
    ScalarSuffix   string `jsonpat:"_scalar_sfx,suffix"`
    ScalarContains int    `jsonpat:"_scalar_cont_,contains"`
    ScalarRegex    bool   `jsonpat:"^scalar_re_\\d+$,regex"`
}
```

Now, let's unmarshal some JSON:

```go
package main

import (
  "fmt"
  "log"

  "github.com/jamieyoung5/jsonpat"
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

		"re_a123": "regex-A",
		"re_b456": "regex-B",

		"scalar_pfx_data": "scalar-prefix-val",
		"other_scalar_pfx_field": "ignored, scalar already set",

		"data_scalar_sfx": "scalar-suffix-val",
		"data_scalar_cont_data": 12345,
		"scalar_re_99": true,

		"not_matching": "skip me"
	}`)
  // (adapted from test)

  var result TestStruct
  err := jsonpat.Unmarshal(jsonData, &result) //
  if err != nil {
    log.Fatalf("Failed to unmarshal: %v", err)
  }

  // --- Known Fields ---
  fmt.Printf("KnownField:     %s\n", result.KnownField)
  fmt.Printf("OtherKnown:     %d\n", result.OtherKnown)
  fmt.Printf("Ignored:        '%s' (should be empty)\n", result.Ignored)

  // --- Embedded Known Field ---
  fmt.Printf("EmbeddedField:  %s\n", result.EmbeddedField)

  // --- Dynamic Map Fields ---
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

  fmt.Println("\n--- DynamicRegex (^re_.*$,regex) ---")
  for k, v := range result.DynamicRegex {
    fmt.Printf("  %s: %s\n", k, v)
  }

  // --- Dynamic Scalar Fields ---
  fmt.Println("\n--- Dynamic Scalar Fields (First Match Wins) ---")
  fmt.Printf("ScalarPrefix:   %s\n", result.ScalarPrefix)
  fmt.Printf("ScalarSuffix:   %s\n", result.ScalarSuffix)
  fmt.Printf("ScalarContains: %d\n", result.ScalarContains)
  fmt.Printf("ScalarRegex:    %t\n", result.ScalarRegex)
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

--- DynamicRegex (^re_.*$,regex) ---
  re_a123: regex-A
  re_b456: regex-B

--- Dynamic Scalar Fields (First Match Wins) ---
ScalarPrefix:   scalar-prefix-val
ScalarSuffix:   scalar-suffix-val
ScalarContains: 12345
ScalarRegex:    true
```

## Benchmarks

For dynamically matched fields, `jsonpat` does introduce slightly more overhead versus manually parsing into `map[string]interface{}`, however it handles the complexity of iteration, type assertion, and regex matching automatically, saving you from writing potentially brittle, boilerplate-heavy code.

`jsonpat` incurs a negligible overhead (~2.7%) for standard struct fields compared to the standard library.

**Results on Apple M4 Pro:**

```text
BenchmarkOverhead_JsonPat-12          2821650        408.7 ns/op       248 B/op        6 allocs/op
BenchmarkOverhead_StdLib-12           3000922        397.9 ns/op       248 B/op        6 allocs/op
BenchmarkDynamic_JsonPat-12            284892       4119 ns/op        4546 B/op      110 allocs/op
BenchmarkDynamic_MapInterface-12       755284       1590 ns/op         552 B/op       47 allocs/op
```
