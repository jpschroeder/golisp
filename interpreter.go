package main

import (
	"errors"
	"fmt"
	"reflect"
)

var defaultEnv map[Symbol]interface{}

func init() {
	defaultEnv = map[Symbol]interface{}{
		Symbol("+"):           primitive(add),
		Symbol("-"):           primitive(sub),
		Symbol("*"):           primitive(mul),
		Symbol("/"):           primitive(div),
		Symbol("="):           primitive(eq),
		Symbol("<"):           primitive(lt),
		Symbol("<="):          primitive(lte),
		Symbol(">"):           primitive(gt),
		Symbol(">="):          primitive(gte),
		Symbol("list"):        primitive(list),
		Symbol("quote"):       specialform(quote),
		Symbol("do"):          specialform(do),
		Symbol("def"):         specialform(def),
		Symbol("fn"):          specialform(fn),
		Symbol("defn"):        specialform(defn),
		Symbol("if"):          specialform(ifprim),
		Symbol("fmt.Println"): gofunc(fmt.Println),
		Symbol("fmt.Printf"):  gofunc(fmt.Printf),
		Symbol("testfunc"):    gofunc(testfunc),
		Symbol("testvar"):     gofunc(testvar),
		Symbol("testerr1"):    gofunc(testerr1),
		Symbol("testerr2"):    gofunc(testerr2),
		Symbol("testerr3"):    gofunc(testerr3),
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

// primitives take pre-evaluated arguments
type primitive func(args []interface{}) (interface{}, error)

// special forms take unevaluated arguments and the env
type specialform func(args []interface{}, env *Env) (interface{}, error)

// a wrapper around a go function
type gofunc interface{}

type procedure struct {
	params []Symbol
	body   []interface{}
	env    *Env
}

func Eval(val interface{}, env *Env) (interface{}, error) {
	switch t := val.(type) {
	case Symbol:
		return env.Find(t)
	case Vector:
		return evalVector(t, env)
	case map[interface{}]interface{}:
		return evalMap(t, env)
	case []interface{}:
		if len(t) == 0 {
			return t, nil
		}

		front, err := Eval(t[0], env)
		if err != nil {
			return nil, err
		}

		spec, isSpec := front.(specialform)
		if isSpec {
			return spec(t[1:], env)
		}

		args, err := evalList(t[1:], env)
		if err != nil {
			return nil, err
		}

		prim, isPrim := front.(primitive)
		if isPrim {
			return prim(args)
		}

		proc, isProc := front.(procedure)
		if isProc {
			return apply(proc, args)
		}

		fun, isFun := front.(gofunc)
		if isFun {
			return call(fun, args)
		}

		return nil, fmt.Errorf("Invalid proc: %v", front)
	default:
		return t, nil
	}
}

func evalVector(val []interface{}, env *Env) (Vector, error) {
	return evalList(val, env)
}

func evalList(val []interface{}, env *Env) ([]interface{}, error) {
	arr := make([]interface{}, len(val))
	for i, v := range val {
		res, err := Eval(v, env)
		if err != nil {
			return nil, err
		}
		arr[i] = res
	}
	return arr, nil
}

func evalMap(val map[interface{}]interface{}, env *Env) (map[interface{}]interface{}, error) {
	ret := make(map[interface{}]interface{}, len(val))
	for k, v := range val {
		evalK, err := Eval(k, env)
		if err != nil {
			return nil, err
		}
		evalV, err := Eval(v, env)
		if err != nil {
			return nil, err
		}
		ret[evalK] = evalV
	}
	return ret, nil
}

func apply(proc procedure, args []interface{}) (interface{}, error) {
	if len(args) != len(proc.params) {
		return nil, fmt.Errorf("Wrong number of args (%d) passed to procedure", len(args))
	}

	child := ChildEnv(proc.env)
	for i, arg := range args {
		child.Define(proc.params[i], arg)
	}

	return do(proc.body, child)
}

var errorType = reflect.TypeOf((*error)(nil)).Elem()

func call(fun gofunc, args []interface{}) (interface{}, error) {
	f := reflect.ValueOf(interface{}(fun))

	if !isArgLenValid(f.Type(), len(args)) {
		return nil, fmt.Errorf("Wrong number of args (%d) passed to procedure", len(args))
	}

	in := make([]reflect.Value, len(args))
	for i, arg := range args {
		if !isArgTypeValid(f.Type(), reflect.TypeOf(arg), i) {
			return nil, fmt.Errorf("Wrong arg type (%v) passed to procedure", reflect.TypeOf(arg))
		}

		in[i] = reflect.ValueOf(arg)
	}

	result := f.Call(in)

	if len(result) < 1 {
		return nil, nil
	}

	out := make([]interface{}, 0)
	var err error
	for _, res := range result {
		if res.Type() == errorType {
			if !res.IsNil() {
				err = res.Interface().(error)
			}
		} else if res.Kind() == reflect.Slice {
			// If the procedure returns a slice, convert it to a []interface{}
			arr := make([]interface{}, res.Len())
			for i := 0; i < res.Len(); i++ {
				arr[i] = res.Index(i).Interface()
			}
			out = append(out, arr)
		} else {
			out = append(out, res.Interface())
		}
	}

	if len(out) == 0 {
		return nil, err
	}
	if len(out) == 1 {
		return out[0], err
	}
	return out, err
}

func isArgLenValid(funcT reflect.Type, length int) bool {
	if !funcT.IsVariadic() {
		// Non variadic functions need matching argument lengths
		return length == funcT.NumIn()
	} else {
		// Variadic functions need to have at least the non-variadic parts
		// func(x, y int, s ...string) -> NumIn == 3
		return length >= funcT.NumIn()-1
	}
}

func isArgTypeValid(funcT, argT reflect.Type, argIdx int) bool {
	if !funcT.IsVariadic() {
		// Non variadic, check that arg is assignable to param at idx
		return argT.AssignableTo(funcT.In(argIdx))
	} else {
		if argIdx < funcT.NumIn()-1 {
			// Variadic, but before the last param
			return argT.AssignableTo(funcT.In(argIdx))
		} else {
			// Variadic, a parameter in the last array
			return argT.AssignableTo(funcT.In(funcT.NumIn() - 1).Elem())
		}
	}
}

// Primitives

func add(args []interface{}) (interface{}, error) {
	return agg(args,
		func(r, x int) int {
			return r + x
		},
		func(r, x float64) float64 {
			return r + x
		})
}

func sub(args []interface{}) (interface{}, error) {
	return agg(args,
		func(r, x int) int {
			return r - x
		},
		func(r, x float64) float64 {
			return r - x
		})
}

func mul(args []interface{}) (interface{}, error) {
	return agg(args,
		func(r, x int) int {
			return r * x
		},
		func(r, x float64) float64 {
			return r * x
		})
}

func div(args []interface{}) (interface{}, error) {
	return agg(args,
		func(r, x int) int {
			return r / x
		},
		func(r, x float64) float64 {
			return r / x
		})
}

func agg(args []interface{}, accumInt func(int, int) int, accumFloat func(float64, float64) float64) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("Wrong number of args (%d) passed to procedure", len(args))
	}

	ret := args[0]
	for i := 1; i < len(args); i++ {
		ri, rIsInt := ret.(int)
		rf, rIsFloat := ret.(float64)

		ai, aIsInt := args[i].(int)
		af, aIsFloat := args[i].(float64)

		if rIsInt && aIsInt {
			ret = accumInt(ri, ai)
		} else if rIsInt && aIsFloat {
			ret = accumFloat(float64(ri), af)
		} else if rIsFloat && aIsInt {
			ret = accumFloat(rf, float64(ai))
		} else if rIsFloat && aIsFloat {
			ret = accumFloat(rf, af)
		} else {
			return nil, fmt.Errorf("Invalid operand: %v", args[i])
		}
	}
	return ret, nil
}

func lt(args []interface{}) (interface{}, error) {
	return order(args,
		func(r, x int) bool {
			return r < x
		},
		func(r, x float64) bool {
			return r < x
		})
}

func lte(args []interface{}) (interface{}, error) {
	return order(args,
		func(r, x int) bool {
			return r <= x
		},
		func(r, x float64) bool {
			return r <= x
		})
}

func gt(args []interface{}) (interface{}, error) {
	return order(args,
		func(r, x int) bool {
			return r > x
		},
		func(r, x float64) bool {
			return r > x
		})
}

func gte(args []interface{}) (interface{}, error) {
	return order(args,
		func(r, x int) bool {
			return r >= x
		},
		func(r, x float64) bool {
			return r >= x
		})
}

func order(args []interface{}, orderInt func(int, int) bool, orderFloat func(float64, float64) bool) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("Wrong number of args (%d) passed to procedure", len(args))
	}

	for i := 1; i < len(args); i++ {
		ri, rIsInt := args[i-1].(int)
		rf, rIsFloat := args[i-1].(float64)

		ai, aIsInt := args[i].(int)
		af, aIsFloat := args[i].(float64)

		var ret bool
		if rIsInt && aIsInt {
			ret = orderInt(ri, ai)
		} else if rIsInt && aIsFloat {
			ret = orderFloat(float64(ri), af)
		} else if rIsFloat && aIsInt {
			ret = orderFloat(rf, float64(ai))
		} else if rIsFloat && aIsFloat {
			ret = orderFloat(rf, af)
		} else {
			return nil, fmt.Errorf("Invalid operand: %v", args[i])
		}

		if !ret {
			return false, nil
		}
	}
	return true, nil
}

func eq(args []interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("Wrong number of args (%d) passed to: =", len(args))
	}

	compare := args[0]
	for i := 1; i < len(args); i++ {
		if reflect.TypeOf(compare) != reflect.TypeOf(args[i]) {
			return false, nil
		}
		if !Equals(compare, args[i]) {
			return false, nil
		}
	}
	return true, nil
}

func list(args []interface{}) (interface{}, error) {
	return args, nil
}

// Special Forms

func quote(args []interface{}, env *Env) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("Wrong number of args (%d) passed to quote", len(args))
	}
	return args[0], nil
}

func do(args []interface{}, env *Env) (interface{}, error) {
	if len(args) == 0 {
		return nil, nil
	}

	ret, err := evalList(args, env)
	if err != nil {
		return nil, err
	}
	return ret[len(ret)-1], nil
}

func def(args []interface{}, env *Env) (interface{}, error) {
	if len(args) > 2 {
		return nil, fmt.Errorf("Too many arguments to def")
	}
	if len(args) < 2 {
		return nil, fmt.Errorf("Too few arguments to def")
	}

	sym, isSym := args[0].(Symbol)

	if !isSym {
		return nil, fmt.Errorf("First argument to def must be a Symbol")
	}

	evaled, err := Eval(args[1], env)
	if err != nil {
		return nil, err
	}

	env.Define(sym, evaled)
	return sym, nil
}

func fn(args []interface{}, env *Env) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("Too few arguments to fn")
	}

	vect, isVect := args[0].(Vector)
	if !isVect {
		return nil, fmt.Errorf("First argument to fn must be a Vector")
	}

	symbols := make([]Symbol, len(vect))
	for i, v := range vect {
		sym, isSym := v.(Symbol)
		if !isSym {
			return nil, fmt.Errorf("First argument to fn must be a Vector of Symbols")
		}
		symbols[i] = sym
	}

	return procedure{
		params: symbols,
		body:   args[1:],
		env:    env,
	}, nil
}

func defn(args []interface{}, env *Env) (interface{}, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("Too few arguments to defn")
	}

	proc, err := fn(args[1:], env)
	if err != nil {
		return nil, err
	}
	return def([]interface{}{args[0], proc}, env)
}

func ifprim(args []interface{}, env *Env) (interface{}, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("Too few arguments to if")
	}
	if len(args) > 3 {
		return nil, fmt.Errorf("Too many arguments to if")
	}

	cond, err := Eval(args[0], env)
	if err != nil {
		return nil, err
	}

	if isTruthy(cond) {
		return Eval(args[1], env)
	}

	if len(args) == 3 {
		return Eval(args[2], env)
	}

	return nil, nil
}

func isTruthy(val interface{}) bool {
	isTrue, isBoolean := val.(bool)
	if isBoolean {
		return isTrue
	}
	return val != nil
}
