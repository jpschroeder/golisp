package main

import (
	"bufio"
	"errors"
	"io"
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

func TestEmpty(t *testing.T) {
	testEval(t, "()", List{})
	testEval(t, "[]", Vector{})
	testEval(t, "{}", Map{})
}

func TestCollections(t *testing.T) {
	testEval(t, "[1 2 (+ 1 2)]", Vector{1, 2, 3})
	testEval(t, "{\"a\" (+ 7 8)}", Map{"a": 15})
	testEval(t, "{:a (+ 7 8)}", Map{Keyword("a"): 15})
	testEval(t, "(list 1)", List{1})
	testEval(t, "(list 1 2 (+ 1 2))", List{1, 2, 3})
	testEval(t, "(quote (\"a\" (+ 7 8)))", List{"a", List{Symbol("+"), 7, 8}})
}

func TestEq(t *testing.T) {
	testEval(t, "(= 1 2)", false)
	testEval(t, "(= 1 1)", true)
	testEval(t, "(= 1 1.1)", false)
	testEval(t, "(= \"blah\" \"bloo\")", false)
	testEval(t, "(= \"blah\" \"blah\")", true)
	testEval(t, "(= [1 2] [1 2])", true)
	testEval(t, "(= {1 2 3 4} {1 2 3 5})", false)
	testEval(t, "(= {1 2 3 4} {1 2 3 4})", true)
}

func TestOrder(t *testing.T) {
	testEval(t, "(< 1)", true)
	testEval(t, "(< 1 2 3 10)", true)
	testEval(t, "(< 10 20 30 15)", false)
	testEval(t, "(< 10 50 30 40)", false)
	testEval(t, "(< 50 20 30 40)", false)
	testEval(t, "(>= 1)", true)
	testEval(t, "(>= 10 9 8 8.0 -1 -2.5)", true)
	testEval(t, "(>= 0 9 8 8 -1 -2.5)", false)
	testEval(t, "(>= 10 9 8 8.5 -1 -2.5)", false)
}

func TestDo(t *testing.T) {
	testEval(t, "(do)", nil)
	testEval(t, "(do 1)", 1)
	testEval(t, "(do (+ 1 2) (+ 3 4))", 7)
}

func TestDef(t *testing.T) {
	testEval(t, "(def x 10)", Symbol("x"))
	testEval(t, "(do (def x 10) x)", 10)
	testEval(t, "(do (def x (+ 10 5)) (+ x 7))", 22)
}

func TestFn(t *testing.T) {
	testEval(t, "((fn [x] (+ 1 x)) 10)", 11)
	testEval(t, "((fn [x y] (+ y x)) 10 7)", 17)
	testEval(t, `(( 
		(fn [x] (fn [y] (+ x y))) 
		10) 7)`, 17)
}

func TestDefn(t *testing.T) {
	testEval(t, `(do
		(defn add1 [x] (+ 1 x))
		(add1 5))`, 6)
	testEval(t, `(do
		(defn addxy [x y] (+ y x))
		(addxy 10 7))`, 17)
}

func TestIf(t *testing.T) {
	testEval(t, "(if true 1)", 1)
	testEval(t, "(if false 1)", nil)
	testEval(t, "(if true 1 2)", 1)
	testEval(t, "(if false 1 2)", 2)
	testEval(t, "(if (= 1 1) \"trueval\" 2)", "trueval")
	testEval(t, "(if (= 1 2) \"trueval\" 2)", 2)
	testEval(t, "(if \"blah\" 1 2)", 1)
	testEval(t, "(if nil 1 2)", 2)
}

func TestMap(t *testing.T) {
	testEval(t, `({ :a "blah", :b 42} :b)`, 42)
	testEval(t, `(do
		(def nested { :a "blah" :nmap { :na 42 :nb 43}})
		(nested :nmap :nb))`, 43)
}

func TestFib(t *testing.T) {
	testEval(t, `
		(defn fib [n]
			(if (< n 2) 
				n
				(+ (fib (- n 1)) (fib (- n 2)))))
		(fib 10)`, 55)
	testEval(t, `
		(defn fib [n]
			(defn fib-iter [curr next n]
				(if (= n 0)
					curr
					(fib-iter next (+ curr next) (- n 1))))
			(fib-iter 0 1 n))
		(fib 10)`, 55)
}

func TestReflection(t *testing.T) {
	testEvalError(t, "(testfunc)")
	testEvalError(t, "(testfunc 1)")
	testEvalError(t, "(testfunc \"blah\" \"bloo\")")
	testEvalError(t, "(testfunc 1 2)")
	testEvalError(t, "(testfunc 1 \"blah\" 2)")
	testEval(t, "(testfunc 1 \"blah\")", List{1, "blah"})
}

func TestReflectionVariadic(t *testing.T) {
	testEvalError(t, "(testvar)")
	testEvalError(t, "(testvar \"blah\" \"bloo\")")
	testEvalError(t, "(testvar 1 2)")
	testEvalError(t, "(testvar 1 \"blah\" 2)")
	testEval(t, "(testvar 1)", List{1, List{}})
	testEval(t, "(testvar 1 \"blah\")", List{1, List{"blah"}})
	testEval(t, "(testvar 1 \"blah\" \"bloo\")", List{1, List{"blah", "bloo"}})
}

func TestReflectionError(t *testing.T) {
	testEvalError(t, "(testerr1 \"err\")")
	testEval(t, "(testerr1 \"\")", nil)
	testEvalError(t, "(testerr2 1 \"err\")")
	testEval(t, "(testerr2 1 \"\")", 1)
	testEvalError(t, "(testerr3 1 2 \"err\")")
	testEval(t, "(testerr3 1 2 \"\")", List{1, 2})
}

func TestError(t *testing.T) {
	testEvalError(t, "(abc 1 2 3)")
	testEvalError(t, "((fn [x y] (+ y x)) 10 7 8)")
	testEvalError(t, "((fn [x y] (+ y x)) 10)")
	testEvalError(t, "(def)")
	testEvalError(t, "(def x)")
	testEvalError(t, "(def \"x\" 1)")
	testEvalError(t, "(if)")
	testEvalError(t, "(if true)")
	testEvalError(t, "(if true 1 2 3)")
}

func testEval(t *testing.T, input string, output Expr) {
	actual, err := readEval(input, NewEnv())
	if err != nil {
		t.Errorf("\nExpected: %v - %v\nActual: Error - %s\n",
			reflect.TypeOf(output), Print(output),
			err)
		return
	}
	if !Equals(actual, output) {
		t.Errorf("\nExpr: %s\nExpected: %v - %v\nActual: %v - %v\n",
			input,
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

func readEval(input string, env *Env) (val Expr, err error) {
	in := bufio.NewReader(strings.NewReader(input))

	for {
		rval, rerr := Read(in)
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			return "", err
		}

		val, err = Eval(rval, env)
	}

	return
}

// Test functions to call via reflection

func init() {
	testfuncs := map[Symbol]Expr{
		Symbol("testfunc"): gofunc(testfunc),
		Symbol("testvar"):  gofunc(testvar),
		Symbol("testerr1"): gofunc(testerr1),
		Symbol("testerr2"): gofunc(testerr2),
		Symbol("testerr3"): gofunc(testerr3),
	}
	for k, v := range testfuncs {
		defaultEnv[k] = v
	}
}

func testvar(i int, s ...string) (int, []string) {
	return i, s
}

func testfunc(i int, s string) (int, string) {
	return i, s
}

func testerr1(s string) error {
	if len(s) > 0 {
		return errors.New(s)
	}
	return nil
}

func testerr2(i int, s string) (int, error) {
	if len(s) > 0 {
		return i, errors.New(s)
	}
	return i, nil
}

func testerr3(i, j int, s string) (int, int, error) {
	if len(s) > 0 {
		return i, j, errors.New(s)
	}
	return i, j, nil
}
