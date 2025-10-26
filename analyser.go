package dynjson

import (
	"reflect"
	"slices"
	"sync"
)

// structInfo holds cached reflection data for a struct.
type structInfo struct {
	tagging *taggingData
}

// typeCache caches structInfo for seen types to avoid costly re-calculations
var typeCache sync.Map

// analyseStruct analyses a struct for relevant tagging related info
func analyseStruct(typ reflect.Type, info *structInfo, baseIndex []int) error {
	// decided to use a c style loop here rather than range to support older go versions
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		currentIndex := append(slices.Clone(baseIndex), i)

		if !field.IsExported() {
			continue // skip unexported field
		}

		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			if err := analyseStruct(field.Type, info, currentIndex); err != nil {
				return err
			}
			continue
		}

		isDynamic := false
		for tag, handler := range jsonTags {
			if value, ok := field.Tag.Lookup(tag); ok {
				if tag == dynamicTag {
					isDynamic = true
				}

				err := handler(field, currentIndex, value, info)
				if err != nil {
					return err
				}
			}
		}

		// default to field name
		if _, ok := field.Tag.Lookup("json"); !ok {
			if !isDynamic {
				info.tagging.knownFields[field.Name] = currentIndex
			}
		}
	}

	return nil
}

// getStructInfo retrieves cached struct info or analyzes the type if not cached.
func getStructInfo(typ reflect.Type) (*structInfo, error) {
	// check if struct info is already cached
	if v, ok := typeCache.Load(typ); ok {
		return v.(*structInfo), nil
	}

	// analyse struct since it isn't cached
	info := &structInfo{
		tagging: &taggingData{
			knownFields:   make(map[string][]int),
			dynamicFields: make([]dynamicFieldInfo, 0),
		},
	}

	if err := analyseStruct(typ, info, nil); err != nil {
		return nil, err
	}

	// protect against race conditions
	v, _ := typeCache.LoadOrStore(typ, info)
	return v.(*structInfo), nil
}
