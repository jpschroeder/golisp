GO LISP
=======

A simple lisp interpreter written in go.  The syntax is inspired by [clojure](https://clojure.org/).

## Usage

```sh
go build
golisp
```

## Syntax

Recursive Fibonacci Example:

```
user=> (defn fib [n]
    (if (< n 2) 
        n
        (+ (fib (- n 1)) 
           (fib (- n 2)))))
fib
user=> (fib 10)
55
```

Tail call optimized Fibonacci:

```
user=> (defn fib [n]
    (defn fib-iter [curr next n]
        (if (= n 0)
            curr
            (fib-iter next 
                (+ curr next) 
                (- n 1))))
    (fib-iter 0 1 n))
fib
user=> (fib 10)
55
```

## References

* [Make a Lisp](https://github.com/kanaka/mal)
* [Clojure](https://github.com/clojure/clojure)
