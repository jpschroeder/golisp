package main

import "fmt"

type Env struct {
	symbols map[Symbol]any
	parent  *Env
}

func NewEnv() *Env {
	return &Env{defaultEnv, nil}
}

func ChildEnv(parent *Env) *Env {
	return &Env{make(map[Symbol]any), parent}
}

func (e *Env) Define(s Symbol, val any) {
	e.symbols[s] = val
}

func (e *Env) Find(s Symbol) (any, error) {
	f, exists := e.symbols[s]
	if exists {
		return f, nil
	}
	if e.parent == nil {
		return nil, fmt.Errorf("unable to resolve symbol: %v in this context", s)
	}
	return e.parent.Find(s)
}
