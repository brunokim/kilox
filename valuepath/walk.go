package valuepath

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Walk walks the path on the object provided and returns the value found.
func Walk(path string, obj any) (any, error) {
	steps := strings.Split(path, ".")
	if (steps[0] != "") && (steps[0] != "$") {
		return nil, fmt.Errorf("%q: at 0: invalid root object %q", path, steps[0])
	}
	value := reflect.ValueOf(obj)
	for i, step := range steps[1:] {
		// Dereference value until hitting a concrete type.
		for value.Kind() == reflect.Pointer || value.Kind() == reflect.Interface {
			value = value.Elem()
		}
		// If step is a valid integer, consider it an index.
		if idx, err := strconv.Atoi(step); err == nil {
			value, err = getIndex(value, idx)
			if err != nil {
				return nil, fmt.Errorf("%q at %d: %w", path, i, err)
			}
			continue
		}
		var err error
		value, err = getKey(value, step)
		if err != nil {
			return nil, fmt.Errorf("%q at %d: %w", path, i, err)
		}
	}
	return value.Interface(), nil
}

func getIndex(value reflect.Value, idx int) (reflect.Value, error) {
	switch value.Kind() {
	case reflect.Map:
		v := value.MapIndex(reflect.ValueOf(idx))
		if !v.IsValid() {
			return reflect.Value{}, fmt.Errorf("key %d not found in map", idx)
		}
		return v, nil
	case reflect.Array, reflect.Slice, reflect.String:
		if idx < 0 || idx >= value.Len() {
			return reflect.Value{}, fmt.Errorf("index %d is out of range [0,%d)", idx, value.Len())
		}
		return value.Index(idx), nil
	default:
		return reflect.Value{}, fmt.Errorf("can't access member %d of %v", idx, value.Type())
	}
}

func getKey(value reflect.Value, key string) (reflect.Value, error) {
	switch value.Kind() {
	case reflect.Map:
		v := value.MapIndex(reflect.ValueOf(key))
		if !v.IsValid() {
			return reflect.Value{}, fmt.Errorf("key %q not found in map with type %v", key, value.Type())
		}
		return v, nil
	case reflect.Struct:
		v := value.FieldByName(key)
		if !v.IsValid() {
			return reflect.Value{}, fmt.Errorf("field %q not found in struct with type %v", key, value.Type())
		}
		return v, nil
	default:
		return reflect.Value{}, fmt.Errorf("can't access member %q of %v", key, value.Type())
	}
}
