package main

import (
	"bufio"
	"reflect"
	"strings"
	"testing"
)

func TestArithmetic(t *testing.T) {
	testEval(t, "(+ 1 2)", 3)
	testEval(t, "(+ 5 (* 2 3))", 11)
	testEval(t, "(- (+ 5 (* 2 3)) 3)", 8)
	testEval(t, "(/ (- (+ 5 (* 2 3)) 3) 4)", 2)
	testEval(t, "(/ (- (+ 515 (* 87 311)) 302) 27)", 1010)
	testEval(t, "(* -3 6)", -18)
	testEval(t, "(/ (- (+ 515 (* -87 311)) 296) 27)", -994)
}

func TestError(t *testing.T) {
	testEvalError(t, "(abc 1 2 3)")
}

func TestEmpty(t *testing.T) {
	testEval(t, "()", newList())
	testEval(t, "[]", newVect())
	testEval(t, "{}", newMap{})
}

func TestCollections(t *testing.T) {
	testEval(t, "[1 2 (+ 1 2)]", newVect(1, 2, 3))
	testEval(t, "{\"a\" (+ 7 8)}", newMap{"a": 15})
	testEval(t, "{:a (+ 7 8)}", newMap{Keyword("a"): 15})
}

func testEval(t *testing.T, input string, output interface{}) {
	actual, err := readEval(input, NewEnv())
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

func testEvalError(t *testing.T, input string) {
	actual, err := readEval(input, NewEnv())
	if err == nil {
		t.Errorf("Expected: Error\nActual: %v %v\n", reflect.TypeOf(actual), actual)
	}
}

func readEval(input string, env *Env) (interface{}, error) {
	in := bufio.NewReader(strings.NewReader(input))

	val, err := Read(in)
	if err != nil {
		return "", err
	}

	return Eval(val, env)
}
