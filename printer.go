package main

import (
	"fmt"
	"strings"
)

func Print(val any) string {
	switch t := val.(type) {
	case string, rune:
		return fmt.Sprintf("%q", t)
	case nil:
		return "nil"
	case List:
		return fmt.Sprintf("(%s)", printSlice(t))
	case []any:
		return fmt.Sprintf("[%s]", printSlice(t))
	case map[any]any:
		return fmt.Sprintf("{%s}", printMap(t))
	case Keyword:
		return fmt.Sprintf(":%s", t)
	default:
		return fmt.Sprintf("%v", val)
	}
}

func printSlice(val []any) string {
	if len(val) < 1 {
		return ""
	}
	var ret strings.Builder
	fmt.Fprint(&ret, Print(val[0]))
	for _, v := range val[1:] {
		fmt.Fprintf(&ret, " %s", Print(v))
	}
	return ret.String()
}

func printMap(val map[any]any) string {
	if len(val) < 1 {
		return ""
	}
	var ret strings.Builder
	i := 0
	for k, v := range val {
		if i != 0 {
			fmt.Fprintf(&ret, ", ")
		}
		fmt.Fprintf(&ret, "%s %s", Print(k), Print(v))
		i++
	}
	return ret.String()
}
