package main

import (
	"fmt"
	"strings"
)

func Print(val interface{}) string {
	switch t := val.(type) {
	case string, rune:
		return fmt.Sprintf("%q", t)
	case nil:
		return "nil"
	case []interface{}:
		arr := make([]string, len(t))
		for i, v := range t {
			arr[i] = Print(v)
		}
		return fmt.Sprintf("(%s)", strings.Join(arr, " "))
	case Vector:
		arr := make([]string, len(t))
		for i, v := range t {
			arr[i] = Print(v)
		}
		return fmt.Sprintf("[%s]", strings.Join(arr, " "))
	case map[interface{}]interface{}:
		arr := make([]string, len(t))
		i := 0
		for key, val := range t {
			arr[i] = Print(key) + " " + Print(val)
			i++
		}
		return fmt.Sprintf("{%s}", strings.Join(arr, ", "))
	default:
		return fmt.Sprintf("%v", val)
	}
}
