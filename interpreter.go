package main

import (
	"fmt"
	"reflect"
)

var defaultEnv map[Symbol]interface{}

func init() {
	defaultEnv = map[Symbol]interface{}{
		Symbol("+"):     primitive(add),
		Symbol("-"):     primitive(sub),
		Symbol("*"):     primitive(mul),
		Symbol("/"):     primitive(div),
		Symbol("="):     primitive(eq),
		Symbol("<"):     primitive(lt),
		Symbol("<="):    primitive(lte),
		Symbol(">"):     primitive(gt),
		Symbol(">="):    primitive(gte),
		Symbol("list"):  primitive(list),
		Symbol("quote"): specialform(quote),
		Symbol("do"):    specialform(do),
		Symbol("def"):   specialform(def),
		Symbol("fn"):    specialform(fn),
		Symbol("defn"):  specialform(defn),
		Symbol("if"):    specialform(ifprim),
	}
}

// primitives take pre-evaluated arguments
type primitive func(args []interface{}) (interface{}, error)

// special forms take unevaluated arguments and the env
type specialform func(args []interface{}, env *Env) (interface{}, error)

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
