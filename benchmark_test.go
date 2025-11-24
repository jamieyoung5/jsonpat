package jsonpat

import (
	"encoding/json"
	"testing"
)

type BenchStatic struct {
	Name        string `json:"name"`
	Value       int    `json:"value"`
	Description string `json:"description"`
}

type BenchDynamic struct {
	Name    string         `json:"name"`
	Dynamic map[string]int `jsonpat:"val_,prefix"`
}

var (
	staticJSON = []byte(`{"name": "benchmark", "value": 12345, "description": "testing overhead"}`)

	dynamicJSON = []byte(`{
		"name": "benchmark",
		"val_1": 100, "val_2": 200, "val_3": 300, "val_4": 400, "val_5": 500,
		"val_6": 600, "val_7": 700, "val_8": 800, "val_9": 900, "val_10": 1000
	}`)
)

// BenchmarkOverhead_JsonPat measures the cost of using jsonpat on a regular struct.
// Expect this to be slower than StdLib due to the intermediate map[string]RawMessage step.
func BenchmarkOverhead_JsonPat(b *testing.B) {
	var v BenchStatic
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := Unmarshal(staticJSON, &v); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkOverhead_StdLib provides the baseline performance of standard encoding/json.
func BenchmarkOverhead_StdLib(b *testing.B) {
	var v BenchStatic
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := json.Unmarshal(staticJSON, &v); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkDynamic_JsonPat measures unmarshaling into a struct + map using jsonpat.
func BenchmarkDynamic_JsonPat(b *testing.B) {
	var v BenchDynamic
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := Unmarshal(dynamicJSON, &v); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkDynamic_MapInterface measures the alternative "loose" approach:
// unmarshaling everything into map[string]interface{}.
func BenchmarkDynamic_MapInterface(b *testing.B) {
	var v map[string]interface{}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := json.Unmarshal(dynamicJSON, &v); err != nil {
			b.Fatal(err)
		}
	}
}
