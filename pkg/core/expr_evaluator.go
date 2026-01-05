package core

import (
	"errors"
	"fmt"
	"strconv"
	"unicode"
)

// MathTokenType represents lexical token categories.
type MathTokenType int

const (
	MathTokenNumber MathTokenType = iota
	MathTokenIdent
	MathTokenPlus
	MathTokenMinus
	MathTokenMultiply
	MathTokenDivide
	MathTokenLParen
	MathTokenRParen
	MathTokenEOF
)

// Token represents a single lexical token.
type MathToken struct {
	Type    MathTokenType
	Literal string
}

// Tokenizer scans mathematical expressions into token streams.
type Tokenizer struct {
	input []rune
	pos   int
}

// NewTokenizer creates a tokenizer for expression string.
func NewTokenizer(input string) *Tokenizer {
	return &Tokenizer{
		input: []rune(input),
		pos:   0,
	}
}

// NextToken returns next parsed token.
func (t *Tokenizer) NextToken() (MathToken, error) {
	for t.pos < len(t.input) && unicode.IsSpace(t.input[t.pos]) {
		t.pos++
	}

	if t.pos >= len(t.input) {
		return MathToken{Type: MathTokenEOF}, nil
	}

	ch := t.input[t.pos]
	switch ch {
	case '+':
		t.pos++
		return MathToken{Type: MathTokenPlus, Literal: "+"}, nil
	case '-':
		t.pos++
		return MathToken{Type: MathTokenMinus, Literal: "-"}, nil
	case '*':
		t.pos++
		return MathToken{Type: MathTokenMultiply, Literal: "*"}, nil
	case '/':
		t.pos++
		return MathToken{Type: MathTokenDivide, Literal: "/"}, nil
	case '(':
		t.pos++
		return MathToken{Type: MathTokenLParen, Literal: "("}, nil
	case ')':
		t.pos++
		return MathToken{Type: MathTokenRParen, Literal: ")"}, nil
	}

	if unicode.IsDigit(ch) || ch == '.' {
		start := t.pos
		for t.pos < len(t.input) && (unicode.IsDigit(t.input[t.pos]) || t.input[t.pos] == '.') {
			t.pos++
		}
		return MathToken{Type: MathTokenNumber, Literal: string(t.input[start:t.pos])}, nil
	}

	if unicode.IsLetter(ch) || ch == '_' {
		start := t.pos
		for t.pos < len(t.input) && (unicode.IsLetter(t.input[t.pos]) || unicode.IsDigit(t.input[t.pos]) || t.input[t.pos] == '_') {
			t.pos++
		}
		return MathToken{Type: MathTokenIdent, Literal: string(t.input[start:t.pos])}, nil
	}

	return MathToken{}, fmt.Errorf("unexpected character at position %d: %c", t.pos, ch)
}

// MathExprEvaluator evaluates simple arithmetic expressions with variable environment resolution.
type MathExprEvaluator struct {
	tokens []MathToken
	pos    int
	env    map[string]float64
}

// NewMathExprEvaluator creates an expression evaluator.
func NewMathExprEvaluator(expr string, env map[string]float64) (*MathExprEvaluator, error) {
	tok := NewTokenizer(expr)
	var tokens []MathToken
	for {
		t, err := tok.NextToken()
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, t)
		if t.Type == MathTokenEOF {
			break
		}
	}
	if env == nil {
		env = make(map[string]float64)
	}
	return &MathExprEvaluator{
		tokens: tokens,
		pos:    0,
		env:    env,
	}, nil
}

// Evaluate computes arithmetic expression value.
func (e *MathExprEvaluator) Evaluate() (float64, error) {
	val, err := e.parseExpression()
	if err != nil {
		return 0, err
	}
	if e.peek().Type != MathTokenEOF {
		return 0, fmt.Errorf("unexpected trailing token: %s", e.peek().Literal)
	}
	return val, nil
}

func (e *MathExprEvaluator) peek() MathToken {
	if e.pos >= len(e.tokens) {
		return MathToken{Type: MathTokenEOF}
	}
	return e.tokens[e.pos]
}

func (e *MathExprEvaluator) next() MathToken {
	t := e.peek()
	e.pos++
	return t
}

func (e *MathExprEvaluator) parseExpression() (float64, error) {
	left, err := e.parseTerm()
	if err != nil {
		return 0, err
	}

	for {
		switch e.peek().Type {
		case MathTokenPlus:
			e.next()
			right, err := e.parseTerm()
			if err != nil {
				return 0, err
			}
			left += right
		case MathTokenMinus:
			e.next()
			right, err := e.parseTerm()
			if err != nil {
				return 0, err
			}
			left -= right
		default:
			return left, nil
		}
	}
}

func (e *MathExprEvaluator) parseTerm() (float64, error) {
	left, err := e.parseFactor()
	if err != nil {
		return 0, err
	}

	for {
		switch e.peek().Type {
		case MathTokenMultiply:
			e.next()
			right, err := e.parseFactor()
			if err != nil {
				return 0, err
			}
			left *= right
		case MathTokenDivide:
			e.next()
			right, err := e.parseFactor()
			if err != nil {
				return 0, err
			}
			if right == 0 {
				return 0, errors.New("division by zero")
			}
			left /= right
		default:
			return left, nil
		}
	}
}

func (e *MathExprEvaluator) parseFactor() (float64, error) {
	tok := e.next()
	switch tok.Type {
	case MathTokenNumber:
		return strconv.ParseFloat(tok.Literal, 64)
	case MathTokenIdent:
		val, exists := e.env[tok.Literal]
		if !exists {
			return 0, fmt.Errorf("undefined variable: %s", tok.Literal)
		}
		return val, nil
	case MathTokenLParen:
		val, err := e.parseExpression()
		if err != nil {
			return 0, err
		}
		if e.next().Type != MathTokenRParen {
			return 0, errors.New("missing closing parenthesis")
		}
		return val, nil
	case MathTokenMinus:
		// Unary minus
		val, err := e.parseFactor()
		if err != nil {
			return 0, err
		}
		return -val, nil
	default:
		return 0, fmt.Errorf("unexpected factor token: %s", tok.Literal)
	}
}
