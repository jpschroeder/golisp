package main

import (
	"encoding/json"
	"fmt"
	"reflect"
)

var defaultEnv map[Symbol]Expr

func init() {
	defaultEnv = map[Symbol]Expr{
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
		Symbol("marshal"):     gofunc(marshal),
	}
}

// primitives take pre-evaluated arguments
type primitive func(args []Expr) (Expr, error)

// special forms take unevaluated arguments and the env
type specialform func(args []Expr, env *Env) (Expr, error)

// a wrapper around a go function
type gofunc Expr

// store a user defined function that can be applied later
type procedure struct {
	params []Symbol
	body   []Expr
	env    *Env
}

// a return value that indicates that we should perform tail call optimization
type tailcall struct {
	nextVal Expr
	env     *Env
}

// Evaluate an expression using tail call optimization
func Eval(val Expr, env *Env) (Expr, error) {
	var err error
	for {
		val, err = performEval(val, env)
		if err != nil {
			return nil, err
		}

		tail, isTail := val.(tailcall)
		if !isTail {
			return val, err
		}
		val = tail.nextVal
		env = tail.env
	}
}

// Evaluate an expression (returns tailcall if tco is needed)
func performEval(val Expr, env *Env) (Expr, error) {
	switch t := val.(type) {
	case Symbol:
		return env.Find(t)
	case Vector:
		return evalVector(t, env)
	case Map:
		return evalMap(t, env)
	case List:
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

		args, err := evalSlice(t[1:], env)
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

		mp, isMap := front.(Map)
		if isMap {
			return accessMap(mp, args)
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

// eval all elements in a vector
func evalVector(val Vector, env *Env) (Vector, error) {
	return evalSlice(val, env)
}

// eval all elements in a slice
func evalSlice(val []Expr, env *Env) ([]Expr, error) {
	arr := make([]Expr, len(val))
	for i, v := range val {
		res, err := Eval(v, env)
		if err != nil {
			return nil, err
		}
		arr[i] = res
	}
	return arr, nil
}

// eval all elements in a map
func evalMap(val Map, env *Env) (Map, error) {
	ret := make(Map, len(val))
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

// access values in a (potentially nested) map
func accessMap(val Map, args []Expr) (Expr, error) {
	var ret Expr
	ret = val
	for _, arg := range args {
		asmap, ismap := ret.(Map)
		if !ismap {
			return nil, fmt.Errorf("Trying to access nested value that isn't a map: %v", arg)
		}

		access, exists := asmap[arg]
		if !exists {
			return nil, fmt.Errorf("Value does not exist in map: %v ", arg)
		}
		ret = access
	}
	return ret, nil
}

func apply(proc procedure, args []Expr) (Expr, error) {
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

// call a go function using reflection
func call(fun gofunc, args []Expr) (Expr, error) {
	f := reflect.ValueOf(fun)

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

	out := make(List, 0)
	var err error
	for _, res := range result {
		if res.Type() == errorType {
			if !res.IsNil() {
				err = res.Interface().(error)
			}
		} else if res.Kind() == reflect.Slice {
			// If the procedure returns a slice, convert it to a List
			arr := make(List, res.Len())
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

func add(args []Expr) (Expr, error) {
	return agg(args,
		func(r, x int) int {
			return r + x
		},
		func(r, x float64) float64 {
			return r + x
		})
}

func sub(args []Expr) (Expr, error) {
	return agg(args,
		func(r, x int) int {
			return r - x
		},
		func(r, x float64) float64 {
			return r - x
		})
}

func mul(args []Expr) (Expr, error) {
	return agg(args,
		func(r, x int) int {
			return r * x
		},
		func(r, x float64) float64 {
			return r * x
		})
}

func div(args []Expr) (Expr, error) {
	return agg(args,
		func(r, x int) int {
			return r / x
		},
		func(r, x float64) float64 {
			return r / x
		})
}

func agg(args []Expr, accumInt func(int, int) int, accumFloat func(float64, float64) float64) (Expr, error) {
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

func lt(args []Expr) (Expr, error) {
	return order(args,
		func(r, x int) bool {
			return r < x
		},
		func(r, x float64) bool {
			return r < x
		})
}

func lte(args []Expr) (Expr, error) {
	return order(args,
		func(r, x int) bool {
			return r <= x
		},
		func(r, x float64) bool {
			return r <= x
		})
}

func gt(args []Expr) (Expr, error) {
	return order(args,
		func(r, x int) bool {
			return r > x
		},
		func(r, x float64) bool {
			return r > x
		})
}

func gte(args []Expr) (Expr, error) {
	return order(args,
		func(r, x int) bool {
			return r >= x
		},
		func(r, x float64) bool {
			return r >= x
		})
}

func order(args []Expr, orderInt func(int, int) bool, orderFloat func(float64, float64) bool) (Expr, error) {
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

func eq(args []Expr) (Expr, error) {
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

func list(args []Expr) (Expr, error) {
	return List(args), nil
}

// Special Forms

func quote(args []Expr, env *Env) (Expr, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("Wrong number of args (%d) passed to quote", len(args))
	}
	return args[0], nil
}

func do(args []Expr, env *Env) (Expr, error) {
	if len(args) == 0 {
		return nil, nil
	}

	// evaluate all args except for the last one
	_, err := evalSlice(args[:len(args)-1], env)
	if err != nil {
		return nil, err
	}

	// return an indicator that we should eval the
	// last argument using tail call optimization
	return tailcall{
		nextVal: args[len(args)-1],
		env:     env,
	}, nil
}

func def(args []Expr, env *Env) (Expr, error) {
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

func fn(args []Expr, env *Env) (Expr, error) {
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

func defn(args []Expr, env *Env) (Expr, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("Too few arguments to defn")
	}

	proc, err := fn(args[1:], env)
	if err != nil {
		return nil, err
	}
	return def([]Expr{args[0], proc}, env)
}

func ifprim(args []Expr, env *Env) (Expr, error) {
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
		return tailcall{args[1], env}, nil
	}

	if len(args) == 3 {
		return tailcall{args[2], env}, nil
	}

	return nil, nil
}

func isTruthy(val Expr) bool {
	isTrue, isBoolean := val.(bool)
	if isBoolean {
		return isTrue
	}
	return val != nil
}

func marshal(val interface{}) (string, error) {
	barr, err := json.Marshal(val)
	if err != nil {
		return "", err
	}
	return string(barr), err
}
