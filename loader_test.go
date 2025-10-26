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
	KnownField      string             `json:"known_field"`
	OtherKnown      int                `json:"other"`
	Ignored         string             `json:"-"`
	DynamicPrefix   map[string]int     `jsonpat:"dyn_,prefix"`
	DynamicContains map[string]float64 `jsonpat:"_val_,contains"`
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
		"not_matching": "skip me"
	}`)

	var result TestStruct
	err := UnmarshalJson(jsonData, &result)
	require.NoError(t, err, "UnmarshalJson should not fail")

	assert.Equal(t, "hello", result.KnownField, "KnownField mismatch")
	assert.Equal(t, 123, result.OtherKnown, "OtherKnown mismatch")
	assert.Empty(t, result.Ignored, "Ignored field should be empty")

	assert.Equal(t, "i am embedded", result.EmbeddedField, "EmbeddedField mismatch")

	require.NotNil(t, result.DynamicPrefix, "DynamicPrefix map should not be nil")
	assert.Len(t, result.DynamicPrefix, 2, "DynamicPrefix map length mismatch")
	assert.Equal(t, 1, result.DynamicPrefix["dyn_abc"], "DynamicPrefix 'dyn_abc' value mismatch")
	assert.Equal(t, 2, result.DynamicPrefix["dyn_xyz"], "DynamicPrefix 'dyn_xyz' value mismatch")

	require.NotNil(t, result.DynamicContains, "DynamicContains map should not be nil")
	assert.Len(t, result.DynamicContains, 2, "DynamicContains map length mismatch")
	assert.Equal(t, 10.5, result.DynamicContains["field_val_1"], "DynamicContains 'field_val_1' value mismatch")
	assert.Equal(t, 20.75, result.DynamicContains["field_val_2"], "DynamicContains 'field_val_2' value mismatch")

	require.NotNil(t, result.DynamicSuffix, "DynamicSuffix map should not be nil")
	assert.Len(t, result.DynamicSuffix, 2, "DynamicSuffix map length mismatch")
	assert.Equal(t, "test", result.DynamicSuffix["some_suffix"], "DynamicSuffix 'some_suffix' value mismatch")
	assert.Equal(t, true, result.DynamicSuffix["another_suffix"], "DynamicSuffix 'another_suffix' value mismatch")

	assert.NotContains(t, result.DynamicPrefix, "not_matching", "DynamicPrefix should not contain 'not_matching'")
	assert.NotContains(t, result.DynamicContains, "not_matching", "DynamicContains should not contain 'not_matching'")
	assert.NotContains(t, result.DynamicSuffix, "not_matching", "DynamicSuffix should not contain 'not_matching'")
}

func TestLoad_Errors(t *testing.T) {
	jsonData := []byte(`{}`)
	var val TestStruct

	assert.Error(t, UnmarshalJson(jsonData, nil), "Expected error for nil interface")
	assert.Error(t, UnmarshalJson(jsonData, val), "Expected error for non-pointer")

	var i int
	assert.Error(t, UnmarshalJson(jsonData, &i), "Expected error for pointer to non-struct")

	badJson := []byte(`{ "known_field": "hello" `)
	assert.Error(t, UnmarshalJson(badJson, &val), "Expected error for bad JSON")

	badTypeJson := []byte(`{ "other": "not-a-number" }`)
	assert.Error(t, UnmarshalJson(badTypeJson, &val), "Expected error for known field type mismatch")

	badDynTypeJson := []byte(`{ "dyn_abc": "not-a-number" }`)
	assert.Error(t, UnmarshalJson(badDynTypeJson, &val), "Expected error for dynamic field type mismatch")
}

func Test_getStructInfo_TagErrors(t *testing.T) {
	type BadStruct1 struct {
		DynamicField string `jsonpat:"prefix,prefix"`
	}
	_, err := getStructInfo(reflect.TypeOf(BadStruct1{}))
	assert.Error(t, err, "Expected error for jsonpat on non-map field")

	type BadStruct2 struct {
		DynamicField map[string]int `jsonpat:"prefix"`
	}
	_, err = getStructInfo(reflect.TypeOf(BadStruct2{}))
	assert.Error(t, err, "Expected error for malformed jsonpat tag (not enough parts)")

	type BadStruct3 struct {
		DynamicField map[string]int `jsonpat:"prefix,invalid_type"`
	}
	_, err = getStructInfo(reflect.TypeOf(BadStruct3{}))
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
