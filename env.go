package main

import "fmt"

type Env struct {
	symbols map[Symbol]interface{}
	parent  *Env
}

func NewEnv() *Env {
	return &Env{defaultEnv, nil}
}

func ChildEnv(parent *Env) *Env {
	return &Env{make(map[Symbol]interface{}), parent}
}

func (e *Env) Define(s Symbol, val interface{}) {
	e.symbols[s] = val
}

func (e *Env) Find(s Symbol) (interface{}, error) {
	f, exists := e.symbols[s]
	if exists {
		return f, nil
	}
	if e.parent == nil {
		return nil, fmt.Errorf("Unable to resolve symbol: %v in this context", s)
	}
	return e.parent.Find(s)
}
