package main

import (
	"container/list"
	"fmt"
	"strconv"
)

// todo: less duplication when iterating lists vs slices
// todo: symplified arithmetic ops

var defaultEnv map[Symbol]interface{}

func init() {
	defaultEnv = map[Symbol]interface{}{
		Symbol("+"): add,
		Symbol("-"): sub,
		Symbol("*"): mul,
		Symbol("/"): div,
	}
}

func Eval(val interface{}) (interface{}, error) {
	switch t := val.(type) {
	case *list.List:
		if t.Len() == 0 {
			return t, nil
		}

		arr := make([]interface{}, t.Len())
		i := 0
		for v := t.Front(); v != nil; v = v.Next() {
			res, err := Eval(v.Value)
			if err != nil {
				return nil, err
			}
			arr[i] = res
			i++
		}

		return evalFunc(arr)

	case []interface{}:
		if len(t) == 0 {
			return t, nil
		}

		arr := make([]interface{}, len(t))
		for i, v := range t {
			res, err := Eval(v)
			if err != nil {
				return nil, err
			}
			arr[i] = res
		}

		return evalFunc(arr)

	default:
		return t, nil
	}
}

func evalFunc(arr []interface{}) (interface{}, error) {
	sym, isSym := arr[0].(Symbol)
	if !isSym {
		return nil, fmt.Errorf("Function call is not a symbol: %v", arr[0])
	}

	val, hasVal := defaultEnv[sym]
	if !hasVal {
		return nil, fmt.Errorf("Function does not exist: %v", sym)
	}

	fun, isFun := val.(func(args []interface{}) (interface{}, error))
	if !isFun {
		return nil, fmt.Errorf("Value is not a function: %v", sym)
	}

	return fun(arr[1:])
}

func add(args []interface{}) (interface{}, error) {
	var ret interface{}
	for idx, arg := range args {
		if idx == 0 {
			ret = arg
			continue
		}

		ri, rIsInt := ret.(int)
		rf, rIsFloat := ret.(float64)
		rs, rIsString := ret.(string)

		ai, aIsInt := arg.(int)
		af, aIsFloat := arg.(float64)
		as, aIsString := arg.(string)

		if rIsInt && aIsInt {
			ret = ri + ai
		} else if rIsInt && aIsFloat {
			ret = float64(ri) + af
		} else if rIsInt && aIsString {
			ret = strconv.Itoa(ri) + as
		} else if rIsFloat && aIsInt {
			ret = rf + float64(ai)
		} else if rIsFloat && aIsFloat {
			ret = rf + af
		} else if rIsFloat && aIsString {
			ret = fmt.Sprintf("%v", rf) + as
		} else if rIsString && aIsInt {
			ret = rs + strconv.Itoa(ai)
		} else if rIsString && aIsFloat {
			ret = rs + fmt.Sprintf("%v", af)
		} else if rIsString && aIsString {
			ret = rs + as
		} else {
			return nil, fmt.Errorf("Invalid operand: %v", arg)
		}
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
