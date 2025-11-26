package core

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// DynamicTokenKind identifies lexical tokens.
type DynamicTokenKind int

const (
	TokEnd DynamicTokenKind = iota
	TokNumber
	TokIdent
	TokPlus
	TokMinus
	TokMul
	TokDiv
	TokLParen
	TokRParen
	TokEq
	TokNeq
	TokLt
	TokLte
	TokGt
	TokGte
	TokAnd
	TokOr
	TokNot
)

// DynamicToken represents a parsed lexical token.
type DynamicToken struct {
	Kind DynamicTokenKind
	Val  string
}

// InfixExprParser parses infix arithmetic and comparison expressions into an evaluatable AST.
type InfixExprParser struct {
	tokens []DynamicToken
	pos    int
}

// NewInfixExprParser tokenizes an expression string.
func NewInfixExprParser(expr string) (*InfixExprParser, error) {
	tokens, err := tokenize(expr)
	if err != nil {
		return nil, err
	}
	return &InfixExprParser{tokens: tokens}, nil
}

func tokenize(s string) ([]DynamicToken, error) {
	var tokens []DynamicToken
	runes := []rune(s)
	n := len(runes)
	i := 0

	for i < n {
		r := runes[i]
		if unicode.IsSpace(r) {
			i++
			continue
		}

		if unicode.IsDigit(r) || r == '.' {
			start := i
			for i < n && (unicode.IsDigit(runes[i]) || runes[i] == '.') {
				i++
			}
			tokens = append(tokens, DynamicToken{Kind: TokNumber, Val: string(runes[start:i])})
			continue
		}

		if unicode.IsLetter(r) || r == '_' {
			start := i
			for i < n && (unicode.IsLetter(runes[i]) || unicode.IsDigit(runes[i]) || runes[i] == '_' || runes[i] == '.') {
				i++
			}
			val := string(runes[start:i])
			switch strings.ToLower(val) {
			case "and":
				tokens = append(tokens, DynamicToken{Kind: TokAnd, Val: "and"})
			case "or":
				tokens = append(tokens, DynamicToken{Kind: TokOr, Val: "or"})
			case "not":
				tokens = append(tokens, DynamicToken{Kind: TokNot, Val: "not"})
			default:
				tokens = append(tokens, DynamicToken{Kind: TokIdent, Val: val})
			}
			continue
		}

		// Two-character operators
		if i+1 < n {
			two := string(runes[i : i+2])
			switch two {
			case "==":
				tokens = append(tokens, DynamicToken{Kind: TokEq, Val: "=="})
				i += 2
				continue
			case "!=":
				tokens = append(tokens, DynamicToken{Kind: TokNeq, Val: "!="})
				i += 2
				continue
			case "<=":
				tokens = append(tokens, DynamicToken{Kind: TokLte, Val: "<="})
				i += 2
				continue
			case ">=":
				tokens = append(tokens, DynamicToken{Kind: TokGte, Val: ">="})
				i += 2
				continue
			case "&&":
				tokens = append(tokens, DynamicToken{Kind: TokAnd, Val: "&&"})
				i += 2
				continue
			case "||":
				tokens = append(tokens, DynamicToken{Kind: TokOr, Val: "||"})
				i += 2
				continue
			}
		}

		// Single-character operators
		switch r {
		case '+':
			tokens = append(tokens, DynamicToken{Kind: TokPlus, Val: "+"})
		case '-':
			tokens = append(tokens, DynamicToken{Kind: TokMinus, Val: "-"})
		case '*':
			tokens = append(tokens, DynamicToken{Kind: TokMul, Val: "*"})
		case '/':
			tokens = append(tokens, DynamicToken{Kind: TokDiv, Val: "/"})
		case '(':
			tokens = append(tokens, DynamicToken{Kind: TokLParen, Val: "("})
		case ')':
			tokens = append(tokens, DynamicToken{Kind: TokRParen, Val: ")"})
		case '<':
			tokens = append(tokens, DynamicToken{Kind: TokLt, Val: "<"})
		case '>':
			tokens = append(tokens, DynamicToken{Kind: TokGt, Val: ">"})
		case '!':
			tokens = append(tokens, DynamicToken{Kind: TokNot, Val: "!"})
		default:
			return nil, fmt.Errorf("unexpected character: %c", r)
		}
		i++
	}

	tokens = append(tokens, DynamicToken{Kind: TokEnd, Val: ""})
	return tokens, nil
}

// Parse evaluates and returns computed float64 value given context variables.
func (p *InfixExprParser) Evaluate(vars map[string]float64) (float64, error) {
	return p.parseOr(vars)
}

func (p *InfixExprParser) parseOr(vars map[string]float64) (float64, error) {
	left, err := p.parseAnd(vars)
	if err != nil {
		return 0, err
	}

	for p.peek().Kind == TokOr {
		p.next()
		right, err := p.parseAnd(vars)
		if err != nil {
			return 0, err
		}
		if (left != 0) || (right != 0) {
			left = 1
		} else {
			left = 0
		}
	}
	return left, nil
}

func (p *InfixExprParser) parseAnd(vars map[string]float64) (float64, error) {
	left, err := p.parseComparison(vars)
	if err != nil {
		return 0, err
	}

	for p.peek().Kind == TokAnd {
		p.next()
		right, err := p.parseComparison(vars)
		if err != nil {
			return 0, err
		}
		if (left != 0) && (right != 0) {
			left = 1
		} else {
			left = 0
		}
	}
	return left, nil
}

func (p *InfixExprParser) parseComparison(vars map[string]float64) (float64, error) {
	left, err := p.parseAddSub(vars)
	if err != nil {
		return 0, err
	}

	switch p.peek().Kind {
	case TokEq:
		p.next()
		right, err := p.parseAddSub(vars)
		if err != nil {
			return 0, err
		}
		if left == right {
			return 1, nil
		}
		return 0, nil
	case TokNeq:
		p.next()
		right, err := p.parseAddSub(vars)
		if err != nil {
			return 0, err
		}
		if left != right {
			return 1, nil
		}
		return 0, nil
	case TokLt:
		p.next()
		right, err := p.parseAddSub(vars)
		if err != nil {
			return 0, err
		}
		if left < right {
			return 1, nil
		}
		return 0, nil
	case TokLte:
		p.next()
		right, err := p.parseAddSub(vars)
		if err != nil {
			return 0, err
		}
		if left <= right {
			return 1, nil
		}
		return 0, nil
	case TokGt:
		p.next()
		right, err := p.parseAddSub(vars)
		if err != nil {
			return 0, err
		}
		if left > right {
			return 1, nil
		}
		return 0, nil
	case TokGte:
		p.next()
		right, err := p.parseAddSub(vars)
		if err != nil {
			return 0, err
		}
		if left >= right {
			return 1, nil
		}
		return 0, nil
	}

	return left, nil
}

func (p *InfixExprParser) parseAddSub(vars map[string]float64) (float64, error) {
	left, err := p.parseMulDiv(vars)
	if err != nil {
		return 0, err
	}

	for {
		tok := p.peek()
		if tok.Kind == TokPlus {
			p.next()
			right, err := p.parseMulDiv(vars)
			if err != nil {
				return 0, err
			}
			left += right
		} else if tok.Kind == TokMinus {
			p.next()
			right, err := p.parseMulDiv(vars)
			if err != nil {
				return 0, err
			}
			left -= right
		} else {
			break
		}
	}
	return left, nil
}

func (p *InfixExprParser) parseMulDiv(vars map[string]float64) (float64, error) {
	left, err := p.parsePrimary(vars)
	if err != nil {
		return 0, err
	}

	for {
		tok := p.peek()
		if tok.Kind == TokMul {
			p.next()
			right, err := p.parsePrimary(vars)
			if err != nil {
				return 0, err
			}
			left *= right
		} else if tok.Kind == TokDiv {
			p.next()
			right, err := p.parsePrimary(vars)
			if err != nil {
				return 0, err
			}
			if right == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			left /= right
		} else {
			break
		}
	}
	return left, nil
}

func (p *InfixExprParser) parsePrimary(vars map[string]float64) (float64, error) {
	tok := p.peek()

	if tok.Kind == TokNumber {
		p.next()
		val, err := strconv.ParseFloat(tok.Val, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid number: %s", tok.Val)
		}
		return val, nil
	}

	if tok.Kind == TokIdent {
		p.next()
		val, exists := vars[tok.Val]
		if !exists {
			return 0, fmt.Errorf("unknown variable %q", tok.Val)
		}
		return val, nil
	}

	if tok.Kind == TokMinus {
		p.next()
		prim, err := p.parsePrimary(vars)
		if err != nil {
			return 0, err
		}
		return -prim, nil
	}

	if tok.Kind == TokNot {
		p.next()
		prim, err := p.parsePrimary(vars)
		if err != nil {
			return 0, err
		}
		if prim == 0 {
			return 1, nil
		}
		return 0, nil
	}

	if tok.Kind == TokLParen {
		p.next()
		val, err := p.parseOr(vars)
		if err != nil {
			return 0, err
		}
		if p.peek().Kind != TokRParen {
			return 0, fmt.Errorf("missing closing parenthesis")
		}
		p.next()
		return val, nil
	}

	return 0, fmt.Errorf("unexpected token: %s", tok.Val)
}

func (p *InfixExprParser) peek() DynamicToken {
	if p.pos < len(p.tokens) {
		return p.tokens[p.pos]
	}
	return DynamicToken{Kind: TokEnd}
}

func (p *InfixExprParser) next() DynamicToken {
	tok := p.peek()
	p.pos++
	return tok
}
