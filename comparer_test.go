package main

import "testing"

func TestEqual(t *testing.T) {
	shouldEqual(t, 1, 1)
	shouldEqual(t, 2.5, 2.5)
	shouldEqual(t, "blah", "blah")
	shouldEqual(t, Symbol("+"), Symbol("+"))
	shouldEqual(t, Keyword("blahk"), Keyword("blahk"))
	shouldEqual(t, List{1, 2, "blah", true}, List{1, 2, "blah", true})
	shouldEqual(t, Vector{1, 2, "blah", true}, Vector{1, 2, "blah", true})
	shouldEqual(t, Map{1: 2, "blah": List{1, 2}}, Map{1: 2, "blah": List{1, 2}})
}

func TestNotEqual(t *testing.T) {
	shouldNotEqual(t, 1, 2)
	shouldNotEqual(t, 2.5, 3.6)
	shouldNotEqual(t, "blah", "bloo")
	shouldNotEqual(t, Symbol("+"), Symbol("-"))
	shouldNotEqual(t, Keyword("blahk"), Keyword("blook"))
	shouldNotEqual(t, List{1, 2, "blah", true}, List{1, 3, "blah", false})
	shouldNotEqual(t, Vector{1, 2, "blah", true}, Vector{1, 3, "blah", false})
	shouldNotEqual(t, Map{1: 2, "blah": List{1, 2}}, Map{1: 2, "blah": List{1, 3}})
}

func TestTypeMismatch(t *testing.T) {
	shouldNotEqual(t, 1, "blah")
	shouldNotEqual(t, Symbol("blah"), "blah")
	shouldNotEqual(t, Symbol("blah"), Keyword("blah"))
	shouldNotEqual(t, List{1, 2, "blah", true}, Vector{1, 2, "blah", true})
}

func shouldEqual(t *testing.T, val1, val2 Expr) {
	if !Equals(val1, val2) {
		t.Errorf("\n%v | %v - Expected: equal", val1, val2)
	}
}

func shouldNotEqual(t *testing.T, val1, val2 Expr) {
	if Equals(val1, val2) {
		t.Errorf("\n%v | %v - Expected: not equal", val1, val2)
	}
}
