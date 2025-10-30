package jsonpat

import (
	"reflect"
	"regexp"
	"slices"
	"strings"
	"sync"
)

const (
	jsonPatTag          = "jsonpat"
	jsonPatTagSeperator = ","
)

// structInfo holds cached reflection data for a struct.
type structInfo struct {
	tagging *taggingData
}

// typeCache caches structInfo for seen types to avoid costly re-calculations
var typeCache sync.Map

type dynamicFieldInfo struct {
	fieldIndices []int
	value        string
	loadType     string
	re           *regexp.Regexp
}

type taggingData struct {
	knownFields         map[string][]int
	dynamicMapFields    []dynamicFieldInfo
	dynamicScalarFields []dynamicFieldInfo
}

// analyseStruct analyses a struct for relevant tagging related info
func analyseStruct(typ reflect.Type, info *structInfo, baseIndex []int) error {
	// decided to use a c style loop here rather than 'range' to support older go versions
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		currentIndex := append(slices.Clone(baseIndex), i)

		if !field.IsExported() {
			continue // skip unexported field
		}

		// support embedded structs
		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			if err := analyseStruct(field.Type, info, currentIndex); err != nil {
				return err
			}
			continue
		}

		err := analyseFieldTag(field, currentIndex, info)
		if err != nil {
			return err
		}
	}

	return nil
}

// analyseFieldTag routes the analysis of a fields tag
func analyseFieldTag(field reflect.StructField, fieldIndex []int, info *structInfo) error {
	if value, ok := field.Tag.Lookup(jsonPatTag); ok {
		return analyseJsonPatTag(field, fieldIndex, value, info)
	} else if _, ok = field.Tag.Lookup("json"); ok {
		return analyseJsonTag(field, fieldIndex, info)
	}

	// default to field name since there are no tags
	info.tagging.knownFields[field.Name] = fieldIndex
	return nil
}

// analyseJsonPatTag parses and stores a jsonpat tags info
func analyseJsonPatTag(field reflect.StructField, fieldIndex []int, value string, info *structInfo) error {

	data := strings.Split(value, jsonPatTagSeperator)
	matcher, err := extractMatcher(data)
	if err != nil {
		return err
	}

	fieldInfo := dynamicFieldInfo{
		fieldIndices: fieldIndex,
		value:        strings.TrimSpace(data[0]),
		loadType:     matcher,
	}

	// compile/validate regex once
	if matcher == "regex" {
		fieldInfo.re = regexp.MustCompile(fieldInfo.value)
	}

	if field.Type.Kind() == reflect.Map {
		info.tagging.dynamicMapFields = append(info.tagging.dynamicMapFields, fieldInfo)
	} else {
		info.tagging.dynamicScalarFields = append(info.tagging.dynamicScalarFields, fieldInfo)
	}

	return nil
}

// analyseJsonTag parses and handles a json field tag
func analyseJsonTag(field reflect.StructField, fieldIndex []int, info *structInfo) error {
	if jsonTag, ok := field.Tag.Lookup("json"); ok {
		jsonName := strings.Split(jsonTag, ",")[0]
		if jsonName != "-" && jsonName != "" {
			info.tagging.knownFields[jsonName] = fieldIndex
		} else if jsonName == "" {
			info.tagging.knownFields[field.Name] = fieldIndex
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
			knownFields:         make(map[string][]int),
			dynamicMapFields:    make([]dynamicFieldInfo, 0),
			dynamicScalarFields: make([]dynamicFieldInfo, 0),
		},
	}

	if err := analyseStruct(typ, info, nil); err != nil {
		return nil, err
	}

	// protect against race conditions
	v, _ := typeCache.LoadOrStore(typ, info)
	return v.(*structInfo), nil
}
