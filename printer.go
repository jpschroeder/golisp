package main

import (
	"container/list"
	"fmt"
	"strings"
)

func Print(val interface{}) string {
	switch t := val.(type) {
	case string, rune:
		return fmt.Sprintf("%q", t)
	case nil:
		return "nil"
	case *list.List:
		arr := make([]string, t.Len())
		i := 0
		for v := t.Front(); v != nil; v = v.Next() {
			arr[i] = Print(v.Value)
			i++
		}
		return fmt.Sprintf("(%s)", strings.Join(arr, " "))
	case []interface{}:
		arr := make([]string, len(t))
		for i, v := range t {
			arr[i] = Print(v)
		}
		return fmt.Sprintf("[%s]", strings.Join(arr, " "))
	case map[interface{}]interface{}:
		arr := make([]string, len(t)*2)
		i := 0
		for key, val := range t {
			arr[i] = Print(key)
			i++
			arr[i] = Print(val)
			i++
		}
		return fmt.Sprintf("{%s}", strings.Join(arr, " "))
	default:
		return fmt.Sprintf("%v", val)
	}
}
