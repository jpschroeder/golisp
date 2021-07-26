GO LISP
=======

A simple lisp interpreter written in go.  The syntax is inspired by [clojure](https://clojure.org/).

## Usage

```sh
go build
golisp
```

## Syntax

Fibonacci Example:

```sh
user=> (defn fib [n]
    (if (< n 2) 
		n
		(+ (fib (- n 1)) (fib (- n 2)))))
fib
user=> (fib 10)
55
```

## References

* [Make a Lisp](https://github.com/kanaka/mal)
* [Clojure](https://github.com/clojure/clojure)
