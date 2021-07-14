package main

import (
	"bufio"
	"reflect"
	"strings"
	"testing"
)

func TestNumbers(t *testing.T) {
	testRead(t, "1", 1)
	testRead(t, "7", 7)
	testRead(t, "  7   ", 7)
	testRead(t, "-123", -123)
}

func TestSymbols(t *testing.T) {
	testRead(t, "+", Symbol("+"))
	testRead(t, "abc", Symbol("abc"))
	testRead(t, "   abc   ", Symbol("abc"))
	testRead(t, "abc5", Symbol("abc5"))
	testRead(t, "abc-def", Symbol("abc-def"))
}

func TestDashes(t *testing.T) {
	testRead(t, "-", Symbol("-"))
	testRead(t, "-abc", Symbol("-abc"))
	testRead(t, "->>", Symbol("->>"))
}

func TestLists(t *testing.T) {
	testRead(t, "(+ 1 2)", newList(Symbol("+"), 1, 2))
	testRead(t, "()", newList())
	testRead(t, "( )", newList())
	testRead(t, "(nil)", newList(nil))
	testRead(t, "((3 4))", newList(newList(3, 4)))
	testRead(t, "(+ 1 (+ 2 3))", newList(Symbol("+"), 1, newList(Symbol("+"), 2, 3)))
	testRead(t, "  ( +   1   (+   2 3   )   )  ", newList(Symbol("+"), 1, newList(Symbol("+"), 2, 3)))
	testRead(t, "(* 1 2)", newList(Symbol("*"), 1, 2))
	testRead(t, "(** 1 2)", newList(Symbol("**"), 1, 2))
	testRead(t, "(* -3 6)", newList(Symbol("*"), -3, 6))
	testRead(t, "(()())", newList(newList(), newList()))
}

func TestCommas(t *testing.T) {
	testRead(t, "(1 2, 3,,,,),,", newList(1, 2, 3))
}

func TestNilTrueFalse(t *testing.T) {
	testRead(t, "nil", nil)
	testRead(t, "true", true)
	testRead(t, "false", false)
}

func TestStrings(t *testing.T) {
	testRead(t, `"abc"`, "abc")
	testRead(t, `   "abc"   `, "abc")
	testRead(t, `"abc (with parens)"`, "abc (with parens)")
	testRead(t, `"abc\"def"`, `abc"def`)
	testRead(t, `""`, "")
	testRead(t, `"\\"`, `\`)
	testRead(t, `"\\\\\\\\\\\\\\\\\\"`, `\\\\\\\\\`)
	testRead(t, `"&"`, "&")
	testRead(t, `"'"`, "'")
	testRead(t, `"("`, "(")
	testRead(t, `")"`, ")")
	testRead(t, `"*"`, "*")
	testRead(t, `"+"`, "+")
	testRead(t, `","`, ",")
	testRead(t, `"-"`, "-")
	testRead(t, `"/"`, "/")
	testRead(t, `":"`, ":")
	testRead(t, `";"`, ";")
	testRead(t, `"<"`, "<")
	testRead(t, `"="`, "=")
	testRead(t, `">"`, ">")
	testRead(t, `"?"`, "?")
	testRead(t, `"@"`, "@")
	testRead(t, `"["`, "[")
	testRead(t, `"]"`, "]")
	testRead(t, `"^"`, "^")
	testRead(t, `"_"`, "_")
	testRead(t, "\"`\"", "`")
	testRead(t, `"{"`, "{")
	testRead(t, `"}"`, "}")
	testRead(t, `"~"`, "~")

	testRead(t, `"\n"`, "\n")
	testRead(t, `"#"`, "#")
	testRead(t, `"$"`, "$")
	testRead(t, `"%"`, "%")
	testRead(t, `"."`, ".")
	testRead(t, `"\\"`, `\`)
	testRead(t, `"|"`, "|")
}

func TestCharacters(t *testing.T) {
	testRead(t, `\a`, 'a')
	testRead(t, `\8`, '8')
	testRead(t, `\newline`, '\n')
	testRead(t, `\tab`, '\t')
	testRead(t, `\space`, ' ')
}

func TestErrors(t *testing.T) {
	testReadError(t, "[1 2")
	testReadError(t, `"abc`)
	testReadError(t, `"`)
	testReadError(t, `"\"`)
	testReadError(t, "(1 \"abc")
	testReadError(t, "(1 \"abc\"")
}

func TestKeywords(t *testing.T) {
	testRead(t, ":kw", Keyword("kw"))
	testRead(t, "(:kw1 :kw2 :kw3)", newList(Keyword("kw1"), Keyword("kw2"), Keyword("kw3")))
}

func TestVectors(t *testing.T) {
	testRead(t, "[+ 1 2]", newVect(Symbol("+"), 1, 2))
	testRead(t, "[]", newVect())
	testRead(t, "[ ]", newVect())
	testRead(t, "[[3 4]]", newVect(newVect(3, 4)))
	testRead(t, "[+ 1 [+ 2 3]]", newVect(Symbol("+"), 1, newVect(Symbol("+"), 2, 3)))
	testRead(t, "  [ +   1   [+   2 3   ]   ]  ", newVect(Symbol("+"), 1, newVect(Symbol("+"), 2, 3)))
	testRead(t, "([])", newList(newVect()))
}

func TestMaps(t *testing.T) {
	testRead(t, "{}", newMap{})
	testRead(t, "{ }", newMap{})
	testRead(t, "{\"abc\" 1}", newMap{"abc": 1})
	testRead(t, "{\"a\" {\"b\" 2}}", newMap{"a": newMap{"b": 2}})
	testRead(t, "{\"a\" {\"b\" {\"c\" 3}}}", newMap{"a": newMap{"b": newMap{"c": 3}}})
	testRead(t, "{  \"a\"  {\"b\"   {  \"cde\"     3   }  }}", newMap{"a": newMap{"b": newMap{"cde": 3}}})
	testRead(t, "{  :a  {:b   {  :cde     3   }  }}", newMap{Keyword("a"): newMap{Keyword("b"): newMap{Keyword("cde"): 3}}})
	testRead(t, "{\"1\" 1}", newMap{"1": 1})
	testRead(t, "({})", newList(newMap{}))
}

func TestComments(t *testing.T) {
	testRead(t, "1 ; comment after expression", 1)
	testRead(t, "1; comment after expression", 1)
	testRead(t, "1;!", 1)
	testRead(t, "1;\"", 1)
	testRead(t, "1;#", 1)
	testRead(t, "1;$", 1)
	testRead(t, "1;%", 1)
	testRead(t, "1;'", 1)
	testRead(t, "1;\\", 1)
	testRead(t, "1;\\\\", 1)
	testRead(t, "1;\\\\\\", 1)
	testRead(t, "1;`", 1)
}

func testRead(t *testing.T, input string, output interface{}) {
	actual, err := read(input)
	if err != nil {
		t.Errorf("\nExpected: %v - %v\nActual: Error - %s\n",
			reflect.TypeOf(output), Print(output),
			err)
		return
	}
	if !Equals(actual, output) {
		t.Errorf("\nExpected: %v - %v\nActual: %v - %v\n",
			reflect.TypeOf(output), Print(output),
			reflect.TypeOf(actual), Print(actual))
	}
}

func testReadError(t *testing.T, input string) {
	actual, err := read(input)
	if err == nil {
		t.Errorf("Expected: Error\nActual: %v %v\n", reflect.TypeOf(actual), actual)
	}
}

func read(input string) (interface{}, error) {
	return Read(bufio.NewReader(strings.NewReader(input)))
}

func newList(vals ...interface{}) []interface{} {
	return vals
}

func newVect(vals ...interface{}) Vector {
	return vals
}

type newMap = map[interface{}]interface{}
