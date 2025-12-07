package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
)

var variablePattern = regexp.MustCompile(`\$\{([a-zA-Z0-9_\-]+)\}`)

// Substitute walks through arbitrary data structures and replaces ${var}
// placeholders using context variables.
func Substitute(value any, vars map[string]string) any {
	switch v := value.(type) {
	case string:
		return substituteString(v, vars)
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, val := range v {
			out[key] = Substitute(val, vars)
		}
		return out
	case []any:
		out := make([]any, len(v))
		for i, val := range v {
			out[i] = Substitute(val, vars)
		}
		return out
	default:
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Map {
			out := map[string]any{}
			iter := rv.MapRange()
			for iter.Next() {
				key := fmt.Sprint(iter.Key().Interface())
				out[key] = Substitute(iter.Value().Interface(), vars)
			}
			return out
		}
		return value
	}
}

func substituteString(in string, vars map[string]string) string {
	return variablePattern.ReplaceAllStringFunc(in, func(match string) string {
		groups := variablePattern.FindStringSubmatch(match)
		if len(groups) != 2 {
			return match
		}
		if val, ok := vars[groups[1]]; ok {
			return val
		}
		return match
	})
}

// CloneMap returns a deep copy of an arbitrary map; used for building
// request payloads where we must avoid mutating original suite definitions.
func CloneMap(in map[string]any) map[string]any {
	if in == nil {
		return nil
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = CloneValue(v)
	}
	return out
}

// CloneValue clones supported container types.
func CloneValue(v any) any {
	switch val := v.(type) {
	case map[string]any:
		return CloneMap(val)
	case []any:
		dup := make([]any, len(val))
		for i := range val {
			dup[i] = CloneValue(val[i])
		}
		return dup
	default:
		data, err := json.Marshal(val)
		if err != nil {
			return val
		}
		var cloned any
		if err := json.Unmarshal(data, &cloned); err != nil {
			return val
		}
		return cloned
	}
}
