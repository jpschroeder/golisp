package main

import (
	"fmt"
	"strings"
)

func Print(val Expr) string {
	switch t := val.(type) {
	case string, rune:
		return fmt.Sprintf("%q", t)
	case nil:
		return "nil"
	case List:
		return fmt.Sprintf("(%s)", printSlice(t))
	case Vector:
		return fmt.Sprintf("[%s]", printSlice(t))
	case Map:
		return fmt.Sprintf("{%s}", printMap(t))
	case Keyword:
		return fmt.Sprintf(":%s", t)
	default:
		return fmt.Sprintf("%v", val)
	}
}

func printSlice(val []Expr) string {
	if len(val) < 1 {
		return ""
	}
	var ret strings.Builder
	fmt.Fprintf(&ret, Print(val[0]))
	for _, v := range val[1:] {
		fmt.Fprintf(&ret, " %s", Print(v))
	}
	return ret.String()
}

func printMap(val Map) string {
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
