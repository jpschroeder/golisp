package main

type Env struct {
	symbols map[Symbol]interface{}
	parent  *Env
}

func NewEnv(s map[Symbol]interface{}) *Env {
	return &Env{s, nil}
}
