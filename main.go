package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
)

type Symbol string
type Keyword string
type Expr interface{}
type Vector []Expr
type List []Expr
type Map map[Expr]Expr

func ReadEvalPrint(in *bufio.Reader, env *Env) (string, error) {
	val, err := Read(in)
	if err != nil {
		return "", err
	}

	val, err = Eval(val, env)
	if err != nil {
		return "", err
	}

	out := Print(val)
	return out, nil
}

func ReadEvalPrintLoop() {
	r := bufio.NewReader(os.Stdin)
	prompt()
	env := NewEnv()
	for {
		output, err := ReadEvalPrint(r, env)
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println(err)
			prompt()
			continue
		}

		fmt.Println(output)

		peeked, err := r.Peek(1)
		if err == nil && (peeked[0] == '\r' || peeked[0] == '\n') {
			prompt()
		}
	}
}

func prompt() {
	if !isInputRedirected() {
		fmt.Printf("user=> ")
	}
}

func isInputRedirected() bool {
	fi, _ := os.Stdin.Stat()
	return (fi.Mode() & os.ModeCharDevice) == 0
}

func setupCloseHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Exit(0)
	}()
}

func main() {
	setupCloseHandler()
	ReadEvalPrintLoop()
}
