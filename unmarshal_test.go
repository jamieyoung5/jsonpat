package jsonpat

import (
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type EmbeddedStruct struct {
	EmbeddedField string                 `json:"embedded_field"`
	DynamicSuffix map[string]interface{} `jsonpat:"_suffix,suffix"`
}

type TestStruct struct {
	EmbeddedStruct
	KnownField string `json:"known_field"`
	OtherKnown int    `json:"other"`
	Ignored    string `json:"-"`

	DynamicPrefix   map[string]int     `jsonpat:"dyn_,prefix"`
	DynamicContains map[string]float64 `jsonpat:"_val_,contains"`
	DynamicRegex    map[string]string  `jsonpat:"^re_.*$,regex"`

	ScalarPrefix   string `jsonpat:"scalar_pfx_,prefix"`
	ScalarSuffix   string `jsonpat:"_scalar_sfx,suffix"`
	ScalarContains int    `jsonpat:"_scalar_cont_,contains"`
	ScalarRegex    bool   `jsonpat:"^scalar_re_\\d+$,regex"`
}

func TestLoad(t *testing.T) {
	jsonData := []byte(`{
		"known_field": "hello",
		"other": 123,
		"ignored": "should not be Loaded",
		"embedded_field": "i am embedded",
		"dyn_abc": 1,
		"dyn_xyz": 2,
		"field_val_1": 10.5,
		"field_val_2": 20.75,
		"some_suffix": "test",
		"another_suffix": true,
		"re_a123": "regex-A",
		"re_b456": "regex-B",
		"scalar_pfx_data": "scalar-prefix-val",
		"data_scalar_sfx": "scalar-suffix-val",
		"data_scalar_cont_data": 12345,
		"scalar_re_99": true,
		"not_matching": "skip me"
	}`)

	var result TestStruct
	typeCache = sync.Map{}
	err := Unmarshal(jsonData, &result)
	require.NoError(t, err, "UnmarshalJson should not fail")

	// known
	assert.Equal(t, "hello", result.KnownField, "KnownField mismatch")
	assert.Equal(t, 123, result.OtherKnown, "OtherKnown mismatch")
	assert.Empty(t, result.Ignored, "Ignored field should be empty")

	// embedded
	assert.Equal(t, "i am embedded", result.EmbeddedField, "EmbeddedField mismatch")

	// dynamic map
	require.NotNil(t, result.DynamicPrefix, "DynamicPrefix map should not be nil")
	assert.Len(t, result.DynamicPrefix, 2, "DynamicPrefix map length mismatch")
	assert.Equal(t, 1, result.DynamicPrefix["dyn_abc"], "DynamicPrefix 'dyn_abc' value mismatch")
	assert.Equal(t, 2, result.DynamicPrefix["dyn_xyz"], "DynamicPrefix 'dyn_xyz' value mismatch")

	require.NotNil(t, result.DynamicContains, "DynamicContains map should not be nil")
	assert.Len(t, result.DynamicContains, 2, "DynamicContains map length mismatch")
	assert.Equal(t, 10.5, result.DynamicContains["field_val_1"], "DynamicContains 'field_val_1' value mismatch")
	assert.Equal(t, 20.75, result.DynamicContains["field_val_2"], "DynamicContains 'field_val_2' value mismatch")

	require.NotNil(t, result.DynamicSuffix, "DynamicSuffix map (embedded) should not be nil")
	assert.Len(t, result.DynamicSuffix, 2, "DynamicSuffix map length mismatch")
	assert.Equal(t, "test", result.DynamicSuffix["some_suffix"], "DynamicSuffix 'some_suffix' value mismatch")
	assert.Equal(t, true, result.DynamicSuffix["another_suffix"], "DynamicSuffix 'another_suffix' value mismatch")

	require.NotNil(t, result.DynamicRegex, "DynamicRegex map should not be nil")
	assert.Len(t, result.DynamicRegex, 2, "DynamicRegex map length mismatch")
	assert.Equal(t, "regex-A", result.DynamicRegex["re_a123"], "DynamicRegex 're_a123' value mismatch")
	assert.Equal(t, "regex-B", result.DynamicRegex["re_b456"], "DynamicRegex 're_b456' value mismatch")

	// dynamic scalar
	assert.Equal(t, "scalar-prefix-val", result.ScalarPrefix, "ScalarPrefix mismatch")
	assert.Equal(t, "scalar-suffix-val", result.ScalarSuffix, "ScalarSuffix mismatch")
	assert.Equal(t, 12345, result.ScalarContains, "ScalarContains mismatch")
	assert.Equal(t, true, result.ScalarRegex, "ScalarRegex mismatch")

	// non matching
	assert.NotContains(t, result.DynamicPrefix, "not_matching", "DynamicPrefix should not contain 'not_matching'")
	assert.NotContains(t, result.DynamicContains, "not_matching", "DynamicContains should not contain 'not_matching'")
	assert.NotContains(t, result.DynamicSuffix, "not_matching", "DynamicSuffix should not contain 'not_matching'")
	assert.NotContains(t, result.DynamicRegex, "not_matching", "DynamicRegex should not contain 'not_matching'")
}

func TestLoad_ScalarFirstMatchWins(t *testing.T) {
	type ScalarStruct struct {
		Scalar string `jsonpat:"pfx_,prefix"`
	}

	jsonData := []byte(`{
		"pfx_aaa": "first",
		"pfx_bbb": "second"
	}`)

	var result ScalarStruct

	typeCache = sync.Map{}
	err := Unmarshal(jsonData, &result)
	require.NoError(t, err, "UnmarshalJson should not fail")

	assert.NotEmpty(t, result.Scalar, "Scalar field should be populated")
	assert.Contains(t, []string{"first", "second"}, result.Scalar, "Scalar field should be one of the matching values")
}

func TestLoad_Errors(t *testing.T) {
	jsonData := []byte(`{}`)
	var val TestStruct

	assert.Error(t, Unmarshal(jsonData, nil), "Expected error for nil interface")
	assert.Error(t, Unmarshal(jsonData, val), "Expected error for non-pointer")

	var i int
	assert.Error(t, Unmarshal(jsonData, &i), "Expected error for pointer to non-struct")

	badJson := []byte(`{ "known_field": "hello" `)
	assert.Error(t, Unmarshal(badJson, &val), "Expected error for bad JSON")

	badTypeJson := []byte(`{ "other": "not-a-number" }`)
	assert.Error(t, Unmarshal(badTypeJson, &val), "Expected error for known field type mismatch")

	badDynTypeJson := []byte(`{ "dyn_abc": "not-a-number" }`)
	assert.Error(t, Unmarshal(badDynTypeJson, &val), "Expected error for dynamic field type mismatch")

	badScalarTypeJson := []byte(`{ "scalar_re_99": "not-a-bool" }`)
	assert.Error(t, Unmarshal(badScalarTypeJson, &val), "Expected error for dynamic scalar field type mismatch")
}

func Test_getStructInfo_TagErrors(t *testing.T) {
	typeCache = sync.Map{}

	type BadStruct2 struct {
		DynamicField map[string]int `jsonpat:"prefix,invalid_type"`
	}
	_, err := getStructInfo(reflect.TypeOf(BadStruct2{}))
	assert.Error(t, err, "Expected error for invalid UnmarshalJson type in jsonpat tag")
}

func Test_getStructInfo_Cache(t *testing.T) {
	typ := reflect.TypeOf(TestStruct{})

	typeCache = sync.Map{}

	info1, err := getStructInfo(typ)
	require.NoError(t, err, "First call to getStructInfo failed")

	info2, err := getStructInfo(typ)
	require.NoError(t, err, "Second call to getStructInfo failed")

	assert.Same(t, info1, info2, "getStructInfo should return the same cached pointer")
}

func FuzzUnmarshal(f *testing.F) {
	f.Add([]byte(`{
		"known_field": "hello",
		"other": 123,
		"ignored": "should not be Loaded",
		"embedded_field": "i am embedded",
		"dyn_abc": 1,
		"dyn_xyz": 2,
		"field_val_1": 10.5,
		"field_val_2": 20.75,
		"some_suffix": "test",
		"another_suffix": true,
		"re_a123": "regex-A",
		"re_b456": "regex-B",
		"scalar_pfx_data": "scalar-prefix-val",
		"data_scalar_sfx": "scalar-suffix-val",
		"data_scalar_cont_data": 12345,
		"scalar_re_99": true,
		"not_matching": "skip me"
	}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"broken":`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var result TestStruct

		_ = Unmarshal(data, &result)
	})
}
