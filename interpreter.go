package main

import (
	"fmt"
)

// todo: less duplication when iterating lists vs slices

var defaultEnv map[Symbol]interface{}

func init() {
	defaultEnv = map[Symbol]interface{}{
		Symbol("+"):     primitive(add),
		Symbol("-"):     primitive(sub),
		Symbol("*"):     primitive(mul),
		Symbol("/"):     primitive(div),
		Symbol("list"):  primitive(list),
		Symbol("def"):   specialform(def),
		Symbol("quote"): specialform(quote),
		Symbol("fn"):    specialform(fn),
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

func agg(args []interface{}, accumInt func(int, int) int, accumFloat func(float64, float64) float64) (interface{}, error) {
	var ret interface{}
	for idx, arg := range args {
		if idx == 0 {
			ret = arg
			continue
		}

		ri, rIsInt := ret.(int)
		rf, rIsFloat := ret.(float64)

		ai, aIsInt := arg.(int)
		af, aIsFloat := arg.(float64)

		if rIsInt && aIsInt {
			ret = accumInt(ri, ai)
		} else if rIsInt && aIsFloat {
			ret = accumFloat(float64(ri), af)
		} else if rIsFloat && aIsInt {
			ret = accumFloat(rf, float64(ai))
		} else if rIsFloat && aIsFloat {
			ret = accumFloat(rf, af)
		} else {
			return nil, fmt.Errorf("Invalid operand: %v", arg)
		}
	}
	return ret, nil
}

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

func list(args []interface{}) (interface{}, error) {
	if len(args) == 1 {
		return args[0], nil
	}
	return args, nil
}

func quote(args []interface{}, env *Env) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("Wrong number of args (%d) passed to quote", len(args))
	}
	return args[0], nil
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

func apply(proc procedure, args []interface{}) (interface{}, error) {
	if len(args) > len(proc.params) {
		return nil, fmt.Errorf("Too many parameters to procedure")
	}

	child := ChildEnv(proc.env)
	for i, arg := range args {
		child.Define(proc.params[i], arg)
	}

	if len(args) < len(proc.params) {
		return procedure{
			params: proc.params[len(args):],
			body:   proc.body,
			env:    child,
		}, nil
	}

	ret, err := evalList(proc.body, child)
	if err != nil {
		return nil, err
	}
	return ret[len(ret)-1], nil
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
