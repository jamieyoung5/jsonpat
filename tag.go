package dynjson

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
)

const (
	dynamicTag          = "dynamic_json"
	dynamicTagSeperator = ","

	prefixType   = "prefix"
	containsType = "contains"
	suffixType   = "suffix"
)

var (
	loadTypes = []string{prefixType, containsType, suffixType}

	jsonTags = map[string]tagHandler{
		dynamicTag: handleDynamicJsonTag,
		"json":     handleJsonTag,
	}
)

type tagHandler func(field reflect.StructField, fieldIndex []int, value string, info *structInfo) error

type taggingData struct {
	knownFields   map[string][]int
	dynamicFields []dynamicFieldInfo
}

type dynamicFieldInfo struct {
	fieldIndices []int
	value        string
	loadType     string
}

func handleDynamicJsonTag(field reflect.StructField, fieldIndex []int, value string, info *structInfo) error {
	if field.Type.Kind() != reflect.Map {
		return fmt.Errorf("field %s with %s must be a map", field.Name, dynamicTag)
	}

	data := strings.Split(value, dynamicTagSeperator)
	if len(data) != 2 {
		return fmt.Errorf(
			"tag %s on field %s must have a value and search type",
			dynamicTag,
			field.Name,
		)
	}

	loadType := strings.TrimSpace(data[1])
	if !slices.Contains(loadTypes, loadType) {
		return fmt.Errorf(
			"tag %s on field %s has invalid load type; must be one of: %s",
			dynamicTag,
			field.Name,
			strings.Join(loadTypes, ", "),
		)
	}

	info.tagging.dynamicFields = append(info.tagging.dynamicFields, dynamicFieldInfo{
		fieldIndices: fieldIndex,
		value:        strings.TrimSpace(data[0]),
		loadType:     loadType,
	})

	return nil
}

func handleJsonTag(field reflect.StructField, fieldIndex []int, value string, info *structInfo) error {
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
