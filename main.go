package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

type Symbol string
type Keyword string

func ReadPrint(in *bufio.Reader) (string, error) {
	val, err := Read(in)
	if err != nil {
		return "", err
	}

	out := Print(val)
	return out, err
}

func ReadEvalPrint(in *bufio.Reader) (val interface{}, out string, err error) {
	val, err = Read(in)
	if err != nil {
		return
	}

	val, err = Eval(val)
	if err != nil {
		return
	}

	out = Print(val)
	return
}

func ReadEvalPrintLoop() {
	r := bufio.NewReader(os.Stdin)
	prompt()
	for {
		val, output, err := ReadEvalPrint(r)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			prompt()
			continue
		}

		fmt.Println(output)
		if val == io.EOF {
			break
		}

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

func main() {
	ReadEvalPrintLoop()
}
