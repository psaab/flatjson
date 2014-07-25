// Copyright 2014 The flatjson Authors. All rights reserved.
// Use of this source code is governed by the BSD 2-Clause license,
// which can be found in the LICENSE file.

// Package flatjson implements a means of converting a struct into a flattened
// map suitable for JSON encoding. The values in the map are pointers to the
// original struct fields, so the map can be generated once and encoded whenever
// the underlying values change.
package flatjson

import (
	"reflect"
	"strings"
)

type Map map[string]interface{}

// Flatten returns the Map representation of val.
func Flatten(val interface{}) Map {
	return flattenValue(reflect.ValueOf(val))
}

func keyForField(field reflect.StructField, v reflect.Value) (string, bool) {
	if tag := field.Tag.Get("json"); tag != "" {
		tokens := strings.SplitN(tag, ",", 2)
		name := tokens[0]
		opts := ""

		if len(tokens) > 1 {
			opts = tokens[1]
		}

		if name == "-" || strings.Contains(opts, "omitempty") && isEmptyValue(v) {
			return "", false
		} else if name != "" {
			return name, false
		}
	}

	if field.Anonymous {
		return "", true
	}
	return field.Name, false
}

func extractValue(val, fallback reflect.Value) reflect.Value {
	switch val.Kind() {
	case reflect.Struct:
		return val
	case reflect.Ptr:
		return extractValue(val.Elem(), fallback)
	case reflect.Interface:
		return extractValue(val.Elem(), fallback)
	default:
		return fallback
	}
}

func recursiveFlatten(val reflect.Value, prefix string, output Map) int {
	valType := val.Type()
	added := 0

	for i := 0; i < val.NumField(); i++ {
		child := val.Field(i)
		childType := valType.Field(i)
		childPrefix := ""

		key, anonymous := keyForField(childType, child)

		if childType.PkgPath != "" || (key == "" && !anonymous) {
			continue
		}

		child = extractValue(child, child)
		if !anonymous {
			childPrefix = prefix + key + "."
		}

		if child.Kind() == reflect.Struct {
			childAdded := recursiveFlatten(child, childPrefix, output)
			if childAdded != 0 {
				added += childAdded
				continue
			}
		}

		output[prefix+key] = child.Addr().Interface()
		added++
	}

	return added
}

func flattenValue(val reflect.Value) Map {
	if val.Kind() == reflect.Ptr {
		return flattenValue(val.Elem())
	}

	if val.Kind() != reflect.Struct {
		panic("must be called with a struct type")
	}

	m := Map{}
	recursiveFlatten(val, "", m)
	return m
}

func isEmptyValue(v reflect.Value) bool {
	return v.Interface() == reflect.Zero(v.Type()).Interface()
}
