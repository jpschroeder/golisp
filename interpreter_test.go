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

func testEval(t *testing.T, input string, output interface{}) {
	actual, _, err := ReadEvalPrint(bufio.NewReader(strings.NewReader(input)))
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
