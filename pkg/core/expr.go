package core

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type TokenType int

const (
	TokenEOF TokenType = iota
	TokenIdent
	TokenString
	TokenNumber
	TokenBool
	TokenDot
	TokenEqual
	TokenNotEqual
	TokenGreater
	TokenGreaterEqual
	TokenLess
	TokenLessEqual
	TokenAnd
	TokenOr
	TokenNot
	TokenLParen
	TokenRParen
)

type Token struct {
	Type  TokenType
	Value string
	Pos   int
}

type Lexer struct {
	input []rune
	pos   int
}

func NewLexer(input string) *Lexer {
	return &Lexer{input: []rune(input)}
}

func (l *Lexer) NextToken() (Token, error) {
	l.skipWhitespace()
	if l.pos >= len(l.input) {
		return Token{Type: TokenEOF, Pos: l.pos}, nil
	}

	ch := l.input[l.pos]

	if ch == '(' {
		l.pos++
		return Token{Type: TokenLParen, Value: "(", Pos: l.pos - 1}, nil
	}
	if ch == ')' {
		l.pos++
		return Token{Type: TokenRParen, Value: ")", Pos: l.pos - 1}, nil
	}
	if ch == '.' {
		l.pos++
		return Token{Type: TokenDot, Value: ".", Pos: l.pos - 1}, nil
	}

	if ch == '=' && l.peek() == '=' {
		l.pos += 2
		return Token{Type: TokenEqual, Value: "==", Pos: l.pos - 2}, nil
	}
	if ch == '!' && l.peek() == '=' {
		l.pos += 2
		return Token{Type: TokenNotEqual, Value: "!=", Pos: l.pos - 2}, nil
	}
	if ch == '!' {
		l.pos++
		return Token{Type: TokenNot, Value: "!", Pos: l.pos - 1}, nil
	}
	if ch == '>' && l.peek() == '=' {
		l.pos += 2
		return Token{Type: TokenGreaterEqual, Value: ">=", Pos: l.pos - 2}, nil
	}
	if ch == '>' {
		l.pos++
		return Token{Type: TokenGreater, Value: ">", Pos: l.pos - 1}, nil
	}
	if ch == '<' && l.peek() == '=' {
		l.pos += 2
		return Token{Type: TokenLessEqual, Value: "<=", Pos: l.pos - 2}, nil
	}
	if ch == '<' {
		l.pos++
		return Token{Type: TokenLess, Value: "<", Pos: l.pos - 1}, nil
	}
	if ch == '&' && l.peek() == '&' {
		l.pos += 2
		return Token{Type: TokenAnd, Value: "&&", Pos: l.pos - 2}, nil
	}
	if ch == '|' && l.peek() == '|' {
		l.pos += 2
		return Token{Type: TokenOr, Value: "||", Pos: l.pos - 2}, nil
	}

	if ch == '"' || ch == ''' {
		return l.lexString(ch)
	}

	if unicode.IsDigit(ch) {
		return l.lexNumber()
	}

	if unicode.IsLetter(ch) || ch == '_' || ch == '$' {
		return l.lexIdent()
	}

	return Token{}, fmt.Errorf("unexpected character %q at position %d", ch, l.pos)
}

func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) && unicode.IsSpace(l.input[l.pos]) {
		l.pos++
	}
}

func (l *Lexer) peek() rune {
	if l.pos+1 >= len(l.input) {
		return 0
	}
	return l.input[l.pos+1]
}

func (l *Lexer) lexString(quote rune) (Token, error) {
	start := l.pos
	l.pos++
	var sb strings.Builder
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch == '\\' && l.pos+1 < len(l.input) {
			l.pos++
			sb.WriteRune(l.input[l.pos])
			l.pos++
			continue
		}
		if ch == quote {
			l.pos++
			return Token{Type: TokenString, Value: sb.String(), Pos: start}, nil
		}
		sb.WriteRune(ch)
		l.pos++
	}
	return Token{}, fmt.Errorf("unterminated string literal starting at position %d", start)
}

func (l *Lexer) lexNumber() (Token, error) {
	start := l.pos
	var sb strings.Builder
	for l.pos < len(l.input) && (unicode.IsDigit(l.input[l.pos]) || l.input[l.pos] == '.') {
		sb.WriteRune(l.input[l.pos])
		l.pos++
	}
	return Token{Type: TokenNumber, Value: sb.String(), Pos: start}, nil
}

func (l *Lexer) lexIdent() (Token, error) {
	start := l.pos
	var sb strings.Builder
	for l.pos < len(l.input) && (unicode.IsLetter(l.input[l.pos]) || unicode.IsDigit(l.input[l.pos]) || l.input[l.pos] == '_' || l.input[l.pos] == '$' || l.input[l.pos] == '-') {
		sb.WriteRune(l.input[l.pos])
		l.pos++
	}
	val := sb.String()
	if val == "true" || val == "false" {
		return Token{Type: TokenBool, Value: val, Pos: start}, nil
	}
	return Token{Type: TokenIdent, Value: val, Pos: start}, nil
}

type Node interface {
	Eval(ctx map[string]interface{}) (interface{}, error)
}

type LiteralNode struct {
	Value interface{}
}

func (n *LiteralNode) Eval(ctx map[string]interface{}) (interface{}, error) {
	return n.Value, nil
}

type VariableNode struct {
	Path []string
}

func (n *VariableNode) Eval(ctx map[string]interface{}) (interface{}, error) {
	if len(n.Path) == 0 {
		return nil, nil
	}

	var current interface{} = ctx
	for _, segment := range n.Path {
		m, ok := current.(map[string]interface{})
		if !ok {
			return nil, nil
		}
		current = m[segment]
	}
	return current, nil
}

type BinaryOpNode struct {
	Op    TokenType
	Left  Node
	Right Node
}

func (n *BinaryOpNode) Eval(ctx map[string]interface{}) (interface{}, error) {
	leftVal, err := n.Left.Eval(ctx)
	if err != nil {
		return nil, err
	}

	// Short-circuit logical ops
	if n.Op == TokenAnd {
		if !toBool(leftVal) {
			return false, nil
		}
		rightVal, err := n.Right.Eval(ctx)
		if err != nil {
			return nil, err
		}
		return toBool(rightVal), nil
	}

	if n.Op == TokenOr {
		if toBool(leftVal) {
			return true, nil
		}
		rightVal, err := n.Right.Eval(ctx)
		if err != nil {
			return nil, err
		}
		return toBool(rightVal), nil
	}

	rightVal, err := n.Right.Eval(ctx)
	if err != nil {
		return nil, err
	}

	switch n.Op {
	case TokenEqual:
		return fmt.Sprintf("%v", leftVal) == fmt.Sprintf("%v", rightVal), nil
	case TokenNotEqual:
		return fmt.Sprintf("%v", leftVal) != fmt.Sprintf("%v", rightVal), nil
	case TokenGreater:
		lf, lok := toFloat(leftVal)
		rf, rok := toFloat(rightVal)
		if lok && rok {
			return lf > rf, nil
		}
		return fmt.Sprintf("%v", leftVal) > fmt.Sprintf("%v", rightVal), nil
	case TokenGreaterEqual:
		lf, lok := toFloat(leftVal)
		rf, rok := toFloat(rightVal)
		if lok && rok {
			return lf >= rf, nil
		}
		return fmt.Sprintf("%v", leftVal) >= fmt.Sprintf("%v", rightVal), nil
	case TokenLess:
		lf, lok := toFloat(leftVal)
		rf, rok := toFloat(rightVal)
		if lok && rok {
			return lf < rf, nil
		}
		return fmt.Sprintf("%v", leftVal) < fmt.Sprintf("%v", rightVal), nil
	case TokenLessEqual:
		lf, lok := toFloat(leftVal)
		rf, rok := toFloat(rightVal)
		if lok && rok {
			return lf <= rf, nil
		}
		return fmt.Sprintf("%v", leftVal) <= fmt.Sprintf("%v", rightVal), nil
	default:
		return nil, fmt.Errorf("unknown binary operator %v", n.Op)
	}
}

func toBool(v interface{}) bool {
	if v == nil {
		return false
	}
	if b, ok := v.(bool); ok {
		return b
	}
	if s, ok := v.(string); ok {
		return s != "" && s != "false" && s != "0"
	}
	if num, ok := toFloat(v); ok {
		return num != 0
	}
	return true
}

func toFloat(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case float64:
		return val, true
	case string:
		f, err := strconv.ParseFloat(val, 64)
		return f, err == nil
	default:
		return 0, false
	}
}

type Parser struct {
	tokens []Token
	pos    int
}

func NewParser(tokens []Token) *Parser {
	return &Parser{tokens: tokens}
}

func (p *Parser) Parse() (Node, error) {
	if len(p.tokens) == 0 {
		return &LiteralNode{Value: true}, nil
	}
	return p.parseOr()
}

func (p *Parser) parseOr() (Node, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}

	for p.current().Type == TokenOr {
		p.pos++
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpNode{Op: TokenOr, Left: left, Right: right}
	}
	return left, nil
}

func (p *Parser) parseAnd() (Node, error) {
	left, err := p.parseComparison()
	if err != nil {
		return nil, err
	}

	for p.current().Type == TokenAnd {
		p.pos++
		right, err := p.parseComparison()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpNode{Op: TokenAnd, Left: left, Right: right}
	}
	return left, nil
}

func (p *Parser) parseComparison() (Node, error) {
	left, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	t := p.current().Type
	if t == TokenEqual || t == TokenNotEqual || t == TokenGreater || t == TokenGreaterEqual || t == TokenLess || t == TokenLessEqual {
		p.pos++
		right, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		return &BinaryOpNode{Op: t, Left: left, Right: right}, nil
	}

	return left, nil
}

func (p *Parser) parsePrimary() (Node, error) {
	tok := p.current()
	if tok.Type == TokenLParen {
		p.pos++
		expr, err := p.parseOr()
		if err != nil {
			return nil, err
		}
		if p.current().Type != TokenRParen {
			return nil, fmt.Errorf("expected closing parenthesis, got %v", p.current())
		}
		p.pos++
		return expr, nil
	}

	if tok.Type == TokenNumber {
		p.pos++
		f, _ := strconv.ParseFloat(tok.Value, 64)
		return &LiteralNode{Value: f}, nil
	}

	if tok.Type == TokenString {
		p.pos++
		return &LiteralNode{Value: tok.Value}, nil
	}

	if tok.Type == TokenBool {
		p.pos++
		return &LiteralNode{Value: tok.Value == "true"}, nil
	}

	if tok.Type == TokenIdent {
		p.pos++
		path := []string{tok.Value}
		for p.current().Type == TokenDot {
			p.pos++
			nextTok := p.current()
			if nextTok.Type != TokenIdent {
				return nil, fmt.Errorf("expected identifier after dot, got %v", nextTok)
			}
			path = append(path, nextTok.Value)
			p.pos++
		}
		return &VariableNode{Path: path}, nil
	}

	return nil, fmt.Errorf("unexpected token %v at position %d", tok, p.pos)
}

func (p *Parser) current() Token {
	if p.pos >= len(p.tokens) {
		return Token{Type: TokenEOF}
	}
	return p.tokens[p.pos]
}

func EvaluateExpression(expr string, ctx map[string]interface{}) (bool, error) {
	trimmed := strings.TrimSpace(expr)
	if trimmed == "" {
		return true, nil
	}

	lexer := NewLexer(trimmed)
	var tokens []Token
	for {
		tok, err := lexer.NextToken()
		if err != nil {
			return false, err
		}
		if tok.Type == TokenEOF {
			break
		}
		tokens = append(tokens, tok)
	}

	parser := NewParser(tokens)
	ast, err := parser.Parse()
	if err != nil {
		return false, err
	}

	res, err := ast.Eval(ctx)
	if err != nil {
		return false, err
	}
	return toBool(res), nil
}
