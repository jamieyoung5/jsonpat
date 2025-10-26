package dynjson

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// UnmarshalDynJson parses json data into a struct, supporting `dynamic_json` tags
// while preserving existing `json` tagging functionality. Dynamic json tags allow
// for filtered matching of json keys into a struct field.
//
// The 'v' argument must be a non-nil pointer to a struct.
func UnmarshalDynJson(data []byte, v interface{}) error {
	// validata struct pointer
	ptrVal := reflect.ValueOf(v)
	if ptrVal.Kind() != reflect.Ptr || ptrVal.IsNil() {
		return fmt.Errorf("v must be a non-nil pointer to a struct")
	}
	structVal := ptrVal.Elem()
	if structVal.Kind() != reflect.Struct {
		return fmt.Errorf("v must be a non-nil pointer to a struct")
	}
	structType := structVal.Type()

	// retrieve struct analysis
	info, err := getStructInfo(structType)
	if err != nil {
		return fmt.Errorf("failed to analyze struct %s: %w", structType.Name(), err)
	}

	// parse all json data
	var raw map[string]json.RawMessage
	if err = json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("failed to unmarshal raw json: %w", err)
	}

	dynamicMaps := buildDynamicMaps(info.tagging.dynamicFields, structVal)

	for key, rawValue := range raw {
		if fieldIndices, ok := info.tagging.knownFields[key]; ok {
			field := structVal.FieldByIndex(fieldIndices)

			if err = json.Unmarshal(rawValue, field.Addr().Interface()); err != nil {
				return fmt.Errorf("failed to unmarshal known key %s: %w", key, err)
			}
			continue
		}

		for _, dynInfo := range info.tagging.dynamicFields {
			matches := false
			switch dynInfo.loadType {
			case "prefix":
				matches = strings.HasPrefix(key, dynInfo.value)
			case "contains":
				matches = strings.Contains(key, dynInfo.value)
			case "suffix":
				matches = strings.HasSuffix(key, dynInfo.value)
			}

			if matches {
				pathKey := fmt.Sprint(dynInfo.fieldIndices)
				if err = unmarshalDynamic(
					dynamicMaps[pathKey],
					key,
					rawValue,
				); err != nil {
					return err
				}
			}
		}

	}

	return nil
}

func unmarshalDynamic(dynMap reflect.Value, key string, jsonRaw json.RawMessage) error {
	mapValType := dynMap.Type().Elem()

	newVal := reflect.New(mapValType)

	if errUnmarshal := json.Unmarshal(jsonRaw, newVal.Interface()); errUnmarshal != nil {
		return fmt.Errorf("failed to unmarshal dynamic key %s: %w", key, errUnmarshal)
	}

	dynMap.SetMapIndex(reflect.ValueOf(key), newVal.Elem())
	return nil
}

func buildDynamicMaps(dynFields []dynamicFieldInfo, structVal reflect.Value) map[string]reflect.Value {
	dynamicMaps := make(map[string]reflect.Value)
	for _, dynInfo := range dynFields {
		fieldVal := structVal.FieldByIndex(dynInfo.fieldIndices)
		if fieldVal.IsNil() {
			fieldVal.Set(reflect.MakeMap(fieldVal.Type()))
		}
		pathKey := fmt.Sprint(dynInfo.fieldIndices)
		dynamicMaps[pathKey] = fieldVal
	}

	return dynamicMaps
}
