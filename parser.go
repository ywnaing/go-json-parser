package main

import (
	"fmt"
	"strconv"
)

type ValueType int

const (
	TypeNull ValueType = iota
	TypeBool
	TypeNumber
	TypeString
	TypeObject
	TypeArray
)

func (t ValueType) String() string {
	switch t {
	case TypeNull:
		return "null"
	case TypeBool:
		return "bool"
	case TypeNumber:
		return "number"
	case TypeString:
		return "string"
	case TypeObject:
		return "object"
	case TypeArray:
		return "array"
	default:
		return "unknown"
	}
}

type Member struct {
	Key string
	Val Value
}

type Value struct {
	Type ValueType
	Bol  bool
	Str  string
	Num  float64
	Arr  []Value
	Obj  []Member
}

type Parser struct {
	lexer   *Lexer
	current Token
	source  string
}

func NewParser(source string) *Parser {
	l := NewLexer(source)

	p := &Parser{
		lexer:  l,
		source: source,
	}

	p.advance()
	return p
}

func (parser *Parser) advance() Token {
	prev := parser.current
	parser.current = parser.lexer.Next()
	return prev
}

func (p *Parser) expect(expected TokenType) (Token, error) {
	if p.current.Type != expected {
		return Token{}, fmt.Errorf(
			"expected %s, got %s (%q)", expected.AsString(),
			p.current.Type.AsString(), p.current.Val(p.source),
		)
	}

	return p.advance(), nil
}

func (p *Parser) Parse() (Value, error) {
	val, err := p.parseValue()

	if err != nil {
		return Value{}, err
	}

	if p.current.Type != TOKEN_EOF {
		return Value{}, fmt.Errorf("unexpected trailing tokens after valid JSON at position %d", p.current.Start)
	}

	return val, nil
}

func (p *Parser) parseValue() (Value, error) {
	switch p.current.Type {
	case TOKEN_LBRACE:
		return p.parseObject()
	case TOKEN_LBRACKET:
		return p.parseArray()
	case TOKEN_TRUE:
		p.advance()
		return Value{
			Type: TypeBool,
			Bol:  true,
		}, nil
	case TOKEN_FALSE:
		p.advance()
		return Value{
			Type: TypeBool,
			Bol:  false,
		}, nil
	case TOKEN_NULL:
		p.advance()
		return Value{
			Type: TypeNull,
		}, nil
	case TOKEN_STRING:
		currentToken := p.advance()
		return Value{
			Type: TypeString,
			Str:  currentToken.Val(p.source),
		}, nil
	case TOKEN_NUMBER:
		currentToken := p.advance()
		floatVal, err := strconv.ParseFloat(currentToken.Val(p.source), 64)

		if err != nil {
			return Value{}, fmt.Errorf("invalid number %q", currentToken.Val(p.source))
		}

		return Value{
			Type: TypeNumber,
			Num:  floatVal,
		}, nil
	default:
		return Value{}, fmt.Errorf("unexpected token %s (%q) when expecting JSON value", p.current.AsString(), p.current.Val(p.source))
	}
}

func (p *Parser) parseObject() (Value, error) {
	// Consume '{'
	p.advance()
	root := Value{
		Type: TypeObject,
		Obj:  []Member{},
	}

	if p.current.Type == TOKEN_RBRACE {
		p.advance()
		return root, nil
	}

	for p.current.Type != TOKEN_RBRACE && p.current.Type != TOKEN_EOF {
		// Key must be string
		current, err := p.expect(TOKEN_STRING)
		if err != nil {
			return Value{}, err
		}

		key := current.Val(p.source)

		// Expect ':'
		if _, err := p.expect(TOKEN_COLON); err != nil {
			return Value{}, err
		}

		// Parse Value
		value, parseErr := p.parseValue()
		if parseErr != nil {
			return Value{}, parseErr
		}

		root.Obj = append(root.Obj, Member{
			Key: key,
			Val: value,
		})

		current = p.advance()

		if current.Type == TOKEN_COMMA {
			continue
		} else if current.Type == TOKEN_RBRACE {
			return root, nil
		} else {
			return Value{}, fmt.Errorf("expected ',' or '}' after object value, got %s (%q)", current.AsString(), current.Val(p.source))
		}
	}

	return Value{}, fmt.Errorf("unexpected end of file inside object")
}

func (p *Parser) parseArray() (Value, error) {
	// Consume '['
	p.advance()

	if p.current.Type == TOKEN_RBRACKET {
		p.advance()
		return Value{
			Type: TypeArray,
			Arr:  []Value{},
		}, nil
	}

	arr := Value{
		Type: TypeArray,
		Arr:  []Value{},
	}

	for p.current.Type != TOKEN_RBRACKET && p.current.Type != TOKEN_EOF {
		val, err := p.parseValue()
		if err != nil {
			return Value{}, err
		}

		arr.Arr = append(arr.Arr, val)

		token := p.advance()

		if token.Type == TOKEN_COMMA {
			continue
		} else if token.Type == TOKEN_RBRACKET {
			return arr, nil
		} else {
			return Value{}, fmt.Errorf("expected ',' or ']' after array value, got %s (%q)", token.AsString(), token.Val(p.source))
		}
	}

	return Value{}, fmt.Errorf("unexpected end of file inside array")
}
