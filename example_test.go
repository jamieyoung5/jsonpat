package jsonpat_test

import (
	"fmt"
	"log"
	"sort"

	"github.com/jamieyoung5/jsonpat"
)

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

func ExampleUnmarshal() {
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

	var result TestStruct
	err := jsonpat.Unmarshal(jsonData, &result)
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
	keysPrefix := make([]string, 0, len(result.DynamicPrefix))
	for k := range result.DynamicPrefix {
		keysPrefix = append(keysPrefix, k)
	}
	sort.Strings(keysPrefix)
	for _, k := range keysPrefix {
		fmt.Printf("  %s: %d\n", k, result.DynamicPrefix[k])
	}

	fmt.Println("\n--- DynamicContains (_val_,contains) ---")
	keysContains := make([]string, 0, len(result.DynamicContains))
	for k := range result.DynamicContains {
		keysContains = append(keysContains, k)
	}
	sort.Strings(keysContains)
	for _, k := range keysContains {
		fmt.Printf("  %s: %f\n", k, result.DynamicContains[k])
	}

	fmt.Println("\n--- DynamicSuffix (_suffix,suffix) ---")
	keysSuffix := make([]string, 0, len(result.DynamicSuffix))
	for k := range result.DynamicSuffix {
		keysSuffix = append(keysSuffix, k)
	}
	sort.Strings(keysSuffix)
	for _, k := range keysSuffix {
		fmt.Printf("  %s: %v\n", k, result.DynamicSuffix[k])
	}

	fmt.Println("\n--- DynamicRegex (^re_.*$,regex) ---")
	keysRegex := make([]string, 0, len(result.DynamicRegex))
	for k := range result.DynamicRegex {
		keysRegex = append(keysRegex, k)
	}
	sort.Strings(keysRegex)
	for _, k := range keysRegex {
		fmt.Printf("  %s: %s\n", k, result.DynamicRegex[k])
	}

	// --- Dynamic Scalar Fields ---
	fmt.Println("\n--- Dynamic Scalar Fields (First Match Wins) ---")
	fmt.Printf("ScalarPrefix:   %s\n", result.ScalarPrefix)
	fmt.Printf("ScalarSuffix:   %s\n", result.ScalarSuffix)
	fmt.Printf("ScalarContains: %d\n", result.ScalarContains)
	fmt.Printf("ScalarRegex:    %t\n", result.ScalarRegex)

	// Output:
	// KnownField:     hello
	// OtherKnown:     123
	// Ignored:        '' (should be empty)
	// EmbeddedField:  i am embedded
	//
	// --- DynamicPrefix (dyn_,prefix) ---
	//   dyn_abc: 1
	//   dyn_xyz: 2
	//
	// --- DynamicContains (_val_,contains) ---
	//   field_val_1: 10.500000
	//   field_val_2: 20.750000
	//
	// --- DynamicSuffix (_suffix,suffix) ---
	//   another_suffix: true
	//   some_suffix: test string
	//
	// --- DynamicRegex (^re_.*$,regex) ---
	//   re_a123: regex-A
	//   re_b456: regex-B
	//
	// --- Dynamic Scalar Fields (First Match Wins) ---
	// ScalarPrefix:   scalar-prefix-val
	// ScalarSuffix:   scalar-suffix-val
	// ScalarContains: 12345
	// ScalarRegex:    true
}
