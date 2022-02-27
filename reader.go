package main

import (
	"bufio"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

var macros map[rune]func(r *bufio.Reader) (Expr, error)

func init() {
	macros = map[rune]func(r *bufio.Reader) (Expr, error){
		'"':  stringReader,
		';':  commentReader,
		'(':  listReader,
		')':  unmatchedDelimiterReader,
		'[':  vectorReader,
		']':  unmatchedDelimiterReader,
		'{':  mapReader,
		'}':  unmatchedDelimiterReader,
		'\\': characterReader,
	}
}

func isWhitespace(ch rune) bool {
	return unicode.IsSpace(ch) || ch == ','
}

func Read(r *bufio.Reader) (Expr, error) {
	for {
		ch, _, err := r.ReadRune()

		for isWhitespace(ch) {
			ch, _, err = r.ReadRune()
		}

		if err != nil {
			return nil, err
		}

		if unicode.IsDigit(ch) {
			return readNumber(r, ch)
		}

		macroFn, isMacro := macros[ch]
		if isMacro {
			ret, err := macroFn(r)
			if ret == r { //no op macros return the reader
				continue
			}
			return ret, err
		}

		if ch == '+' || ch == '-' {
			ch2, _, _ := r.ReadRune()
			r.UnreadRune()
			if unicode.IsDigit(ch2) {
				return readNumber(r, ch)
			}
		}

		token, err := readToken(r, ch)
		if err != nil {
			return nil, err
		}

		return interpretToken(token)
	}
}

func readToken(r *bufio.Reader, initch rune) (string, error) {
	var sb strings.Builder
	sb.WriteRune(initch)

	for {
		ch, _, err := r.ReadRune()

		if err != nil || isWhitespace(ch) || isMacro(ch) {
			r.UnreadRune()
			return sb.String(), nil
		}

		sb.WriteRune(ch)
	}
}

func readNumber(r *bufio.Reader, initch rune) (Expr, error) {
	var sb strings.Builder
	sb.WriteRune(initch)

	for {
		ch, _, err := r.ReadRune()
		if err != nil || isWhitespace(ch) || isMacro(ch) {
			r.UnreadRune()
			break
		}
		sb.WriteRune(ch)
	}

	return matchNumber(sb.String())
}

func interpretToken(s string) (Expr, error) {
	if s == "nil" {
		return nil, nil
	}
	if s == "true" {
		return true, nil
	}
	if s == "false" {
		return false, nil
	}
	if s[0] == ':' {
		return Keyword(s[1:]), nil
	} else {
		return Symbol(s), nil
	}
}

func matchNumber(s string) (Expr, error) {
	i, erri := strconv.Atoi(s)
	if erri == nil {
		return i, nil
	}
	f, errf := strconv.ParseFloat(s, 64)
	if errf == nil {
		return f, nil
	}
	return nil, fmt.Errorf("invalid number: %s", s)
}

func isMacro(ch rune) bool {
	_, ismacro := macros[ch]
	return ismacro
}

func stringReader(r *bufio.Reader) (Expr, error) {
	var sb strings.Builder

	for ch, _, err := r.ReadRune(); ch != '"'; ch, _, err = r.ReadRune() {
		if err != nil {
			return nil, fmt.Errorf("error while reading string: %v", err)
		}
		if ch == '\\' {
			ch, _, err = r.ReadRune()
			if err != nil {
				return nil, err
			}
			switch ch {
			case 't':
				ch = '\t'
			case 'r':
				ch = '\r'
			case 'n':
				ch = '\n'
			case 'b':
				ch = '\b'
			case 'f':
				ch = '\f'
			case '\\':
			case '"':
			default:
				return nil, fmt.Errorf("unsupported escape character: \\%s", string(ch))
			}
		}
		sb.WriteRune(ch)
	}

	return sb.String(), nil
}

func commentReader(r *bufio.Reader) (Expr, error) {
	ch, _, err := r.ReadRune()
	for err != nil && ch != '\n' && ch != '\r' {
		ch, _, err = r.ReadRune()
	}
	return r, nil
}

func characterReader(r *bufio.Reader) (Expr, error) {
	ch, _, err := r.ReadRune()
	if err != nil {
		return nil, err
	}

	token, err := readToken(r, ch)
	if err != nil {
		return nil, err
	}

	if len(token) == 1 {
		return []rune(token)[0], nil
	}

	switch token {
	case "newline":
		return '\n', nil
	case "space":
		return ' ', nil
	case "tab":
		return '\t', nil
	case "backspace":
		return '\b', nil
	case "formfeed":
		return '\f', nil
	case "return":
		return '\r', nil
	default:
		return nil, fmt.Errorf("unsupported character: \\%s", token)
	}
}

func listReader(r *bufio.Reader) (Expr, error) {
	var l []Expr
	err := readDelimitedList(r, ')', func(item Expr) {
		l = append(l, item)
	})
	return List(l), err
}

func vectorReader(r *bufio.Reader) (Expr, error) {
	var l []Expr
	err := readDelimitedList(r, ']', func(item Expr) {
		l = append(l, item)
	})
	return Vector(l), err
}

func mapReader(r *bufio.Reader) (Expr, error) {
	m := make(Map)
	var key Expr
	err := readDelimitedList(r, '}', func(item Expr) {
		if key == nil {
			key = item
		} else {
			m[key] = item
			key = nil
		}
	})
	if key != nil {
		return nil, fmt.Errorf("Map literal must contain an even number of forms")
	}
	return m, err
}

func unmatchedDelimiterReader(r *bufio.Reader) (Expr, error) {
	return nil, errors.New("unmatched delimter")
}

func readDelimitedList(r *bufio.Reader, delim rune, add func(Expr)) error {
	for {
		ch, _, err := r.ReadRune()

		for isWhitespace(ch) {
			ch, _, err = r.ReadRune()
		}

		if err != nil {
			return err
		}

		if ch == delim {
			break
		}

		macroFn, isMacro := macros[ch]
		if isMacro {
			mret, err := macroFn(r)
			if err != nil {
				return err
			}
			if mret != r {
				add(mret)
			}
		} else {
			r.UnreadRune()
			o, err := Read(r)
			if err != nil {
				return err
			}
			if o != r {
				add(o)
			}
		}
	}

	return nil
}
