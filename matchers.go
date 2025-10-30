package jsonpat

import (
	"fmt"
	"slices"
	"strings"
)

const (
	prefixLoadType   = "prefix"
	containsLoadType = "contains"
	suffixLoadType   = "suffix"
	regexLoadType    = "regex"

	defaultMatcher = prefixLoadType
)

var loadTypes = []string{prefixLoadType, containsLoadType, suffixLoadType, regexLoadType}

func extractMatcher(tagValues []string) (string, error) {
	if len(tagValues) < 1 || len(tagValues) > 2 {
		return "", fmt.Errorf(
			"tag %s must have a value and optional search type",
			jsonPatTag,
		)
	}

	if len(tagValues) == 2 {
		matcher := strings.TrimSpace(tagValues[1])

		if !slices.Contains(loadTypes, matcher) {
			return "", fmt.Errorf(
				"tag %s has invalid matcher; must be one of %s",
				jsonPatTag,
				strings.Join(loadTypes, ", "),
			)
		}

		return matcher, nil
	}

	return defaultMatcher, nil
}

func match(key string, fieldInfo dynamicFieldInfo) bool {
	switch fieldInfo.loadType {
	case prefixLoadType:
		return strings.HasPrefix(key, fieldInfo.value)
	case containsLoadType:
		return strings.Contains(key, fieldInfo.value)
	case suffixLoadType:
		return strings.HasSuffix(key, fieldInfo.value)
	case regexLoadType:
		return fieldInfo.re.Match([]byte(key))
	}
	return false
}
