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
	testRead(t, "(+ 1 2)", List{Symbol("+"), 1, 2})
	testRead(t, "()", List{})
	testRead(t, "( )", List{})
	testRead(t, "(nil)", List{nil})
	testRead(t, "((3 4))", List{List{3, 4}})
	testRead(t, "(+ 1 (+ 2 3))", List{Symbol("+"), 1, List{Symbol("+"), 2, 3}})
	testRead(t, "  ( +   1   (+   2 3   )   )  ", List{Symbol("+"), 1, List{Symbol("+"), 2, 3}})
	testRead(t, "(* 1 2)", List{Symbol("*"), 1, 2})
	testRead(t, "(** 1 2)", List{Symbol("**"), 1, 2})
	testRead(t, "(* -3 6)", List{Symbol("*"), -3, 6})
	testRead(t, "(()())", List{List{}, List{}})
}

func TestCommas(t *testing.T) {
	testRead(t, "(1 2, 3,,,,),,", List{1, 2, 3})
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
	testRead(t, "(:kw1 :kw2 :kw3)", List{Keyword("kw1"), Keyword("kw2"), Keyword("kw3")})
}

func TestVectors(t *testing.T) {
	testRead(t, "[+ 1 2]", []any{Symbol("+"), 1, 2})
	testRead(t, "[]", []any{})
	testRead(t, "[ ]", []any{})
	testRead(t, "[[3 4]]", []any{[]any{3, 4}})
	testRead(t, "[+ 1 [+ 2 3]]", []any{Symbol("+"), 1, []any{Symbol("+"), 2, 3}})
	testRead(t, "  [ +   1   [+   2 3   ]   ]  ", []any{Symbol("+"), 1, []any{Symbol("+"), 2, 3}})
	testRead(t, "([])", List{[]any{}})
}

func TestMaps(t *testing.T) {
	testRead(t, "{}", map[any]any{})
	testRead(t, "{ }", map[any]any{})
	testRead(t, "{\"abc\" 1}", map[any]any{"abc": 1})
	testRead(t, "{\"a\" {\"b\" 2}}", map[any]any{"a": map[any]any{"b": 2}})
	testRead(t, "{\"a\" {\"b\" {\"c\" 3}}}", map[any]any{"a": map[any]any{"b": map[any]any{"c": 3}}})
	testRead(t, "{  \"a\"  {\"b\"   {  \"cde\"     3   }  }}", map[any]any{"a": map[any]any{"b": map[any]any{"cde": 3}}})
	testRead(t, "{  :a  {:b   {  :cde     3   }  }}", map[any]any{Keyword("a"): map[any]any{Keyword("b"): map[any]any{Keyword("cde"): 3}}})
	testRead(t, "{\"1\" 1}", map[any]any{"1": 1})
	testRead(t, "({})", List{map[any]any{}})
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

func testRead(t *testing.T, input string, output any) {
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

func read(input string) (any, error) {
	return Read(bufio.NewReader(strings.NewReader(input)))
}
