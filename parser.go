package main

import (
	"fmt"
	"strconv"
	"strings"
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

// GetString extracts the string value if Type == TypeString.
func (val Value) GetString() (string, bool) {
	if val.Type != TypeString {
		return "", false
	}
	return val.Str, true
}

// GetFloat extracts the float64 number if Type == TypeNumber.
func (val Value) GetFloat() (float64, bool) {
	if val.Type != TypeNumber {
		return 0.0, false
	}
	return val.Num, true
}

// GetInt extracts the integer representation if Type == TypeNumber.
func (val Value) GetInt() (int, bool) {
	if val.Type != TypeNumber {
		return 0, false
	}
	return int(val.Num), true
}

// GetBool extracts the boolean value if Type == TypeBool.
func (val Value) GetBool() (bool, bool) {
	if val.Type != TypeBool {
		return false, false
	}
	return val.Bol, true
}

// IsNull returns true if Type == TypeNull.
func (val Value) IsNull() bool {
	return val.Type == TypeNull
}

// Get finds a member by key in a JSON object.
func (val Value) Get(key string) (Value, bool) {
	if val.Type != TypeObject {
		return Value{}, false
	}

	for i := 0; i < len(val.Obj); i++ {
		if val.Obj[i].Key == key {
			return val.Obj[i].Val, true
		}
	}

	return Value{}, false
}

// Index retrieves the element at the specified index in a JSON array.
func (val Value) Index(index int) (Value, bool) {
	if val.Type != TypeArray {
		return Value{}, false
	}

	if index < 0 || index >= len(val.Arr) {
		return Value{}, false
	}

	return val.Arr[index], true
}

// String returns the minified compact JSON representation of the Value.
func (val Value) String() string {
	var sb strings.Builder
	formatValue(val, &sb, "", 0, false)
	return sb.String()
}

// PrettyPrint returns a formatted multi-line JSON string with the specified indentation (e.g. "  ").
func (val Value) PrettyPrint(indent string) string {
	var sb strings.Builder
	formatValue(val, &sb, indent, 0, false)
	return sb.String()
}

// ColorPrint returns a formatted JSON string with ANSI terminal color highlighting.
func (val Value) ColorPrint(indent string) string {
	var sb strings.Builder
	formatValue(val, &sb, indent, 0, true)
	return sb.String()
}

// ToNative converts the Value AST into native Go standard types (map[string]any, []any, float64, etc.).
func (val Value) ToNative() any {
	switch val.Type {
	case TypeNull:
		return nil
	case TypeBool:
		return val.Bol
	case TypeNumber:
		return val.Num
	case TypeString:
		return val.Str
	case TypeArray:
		result := make([]any, len(val.Arr))
		for i, item := range val.Arr {
			result[i] = item.ToNative()
		}
		return result
	case TypeObject:
		result := make(map[string]any, len(val.Obj))
		for _, member := range val.Obj {
			result[member.Key] = member.Val.ToNative()
		}
		return result
	default:
		return nil
	}
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

		unescapeStr, err := unescapeString(currentToken.Val(p.source))
		if err != nil {
			return Value{}, err
		}
		return Value{
			Type: TypeString,
			Str:  unescapeStr,
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

		key, err := unescapeString(current.Val(p.source))
		if err != nil {
			return Value{}, err
		}

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

// unescapeString decodes a quoted JSON string literal into raw Go string content.
func unescapeString(raw string) (string, error) {
	var sb strings.Builder

	// Strip surrounding double quotes if present
	if len(raw) >= 2 && (raw[0] == '"' && raw[len(raw)-1] == '"') {
		raw = raw[1 : len(raw)-1]
	}

	for i := 0; i < len(raw); i++ {
		currChar := raw[i]

		if currChar == '\\' {
			i++

			if i >= len(raw) {
				return "", fmt.Errorf("unexpected end of string after '\\'")
			}

			switch raw[i] {
			case 'n':
				sb.WriteByte('\n')
			case 't':
				sb.WriteByte('\t')
			case 'r':
				sb.WriteByte('\r')
			case 'b':
				sb.WriteByte('\b')
			case 'f':
				sb.WriteByte('\f')
			case '/':
				sb.WriteByte('/')
			case '\\':
				sb.WriteByte('\\')
			case '"':
				sb.WriteByte('"')
			case 'u':
				if i+5 > len(raw) {
					return "", fmt.Errorf("invalid unicode character!")
				}
				hexDigit := raw[i+1 : i+5]
				hexInt, err := strconv.ParseInt(hexDigit, 16, 16)
				if err != nil {
					return "", err
				}

				sb.WriteRune(rune(hexInt))
				i += 4

			default:
				return "", fmt.Errorf("invalid escape character: \\%c", raw[i])
			}
		} else {
			sb.WriteByte(currChar)
		}
	}

	return sb.String(), nil
}

// escapeString encodes a raw Go string into a quoted JSON string with proper escapes.
func escapeString(s string) string {
	var sb strings.Builder
	sb.WriteByte('"')

	for i := 0; i < len(s); i++ {
		c := s[i]

		// Handle control characters (< 0x20) and standard escapes
		if c < 0x20 {
			switch c {
			case '\n':
				sb.WriteString(`\n`)
			case '\t':
				sb.WriteString(`\t`)
			case '\r':
				sb.WriteString(`\r`)
			case '\b':
				sb.WriteString(`\b`)
			case '\f':
				sb.WriteString(`\f`)
			default:
				fmt.Fprintf(&sb, "\\u%04x", c)
			}
		} else {
			switch c {
			case '"':
				sb.WriteString(`\"`)
			case '\\':
				sb.WriteString(`\\`)
			default:
				sb.WriteByte(c)
			}
		}
	}

	sb.WriteByte('"')
	return sb.String()
}

// ANSI escape sequences for terminal syntax highlighting
const (
	colorReset  = "\033[0m"
	colorKey    = "\033[36m" // Cyan for object keys
	colorString = "\033[32m" // Green for string literals
	colorNumber = "\033[33m" // Yellow for numeric values
	colorBool   = "\033[35m" // Magenta for booleans
	colorNull   = "\033[35m" // Magenta for null
	colorPunct  = "\033[90m" // Dark Gray for structural symbols ({ } [ ] : ,)
)

// formatValue recursively renders a Value into valid JSON with optional indentation and colorization.
func formatValue(v Value, sb *strings.Builder, indent string, depth int, colorize bool) {
	isPretty := indent != ""

	switch v.Type {
	case TypeNull:
		if colorize {
			sb.WriteString(colorNull)
		}
		sb.WriteString("null")
		if colorize {
			sb.WriteString(colorReset)
		}

	case TypeBool:
		if colorize {
			sb.WriteString(colorBool)
		}
		if v.Bol {
			sb.WriteString("true")
		} else {
			sb.WriteString("false")
		}
		if colorize {
			sb.WriteString(colorReset)
		}

	case TypeNumber:
		if colorize {
			sb.WriteString(colorNumber)
		}
		sb.WriteString(strconv.FormatFloat(v.Num, 'f', -1, 64))
		if colorize {
			sb.WriteString(colorReset)
		}

	case TypeString:
		if colorize {
			sb.WriteString(colorString)
		}
		sb.WriteString(escapeString(v.Str))
		if colorize {
			sb.WriteString(colorReset)
		}

	case TypeArray:
		if len(v.Arr) == 0 {
			if colorize {
				sb.WriteString(colorPunct)
			}
			sb.WriteString("[]")
			if colorize {
				sb.WriteString(colorReset)
			}
			return
		}

		if colorize {
			sb.WriteString(colorPunct)
		}
		sb.WriteString("[")
		if colorize {
			sb.WriteString(colorReset)
		}

		if isPretty {
			sb.WriteString("\n")
		}

		for i := 0; i < len(v.Arr); i++ {
			if isPretty {
				sb.WriteString(strings.Repeat(indent, depth+1))
			}

			formatValue(v.Arr[i], sb, indent, depth+1, colorize)

			if i < len(v.Arr)-1 {
				if colorize {
					sb.WriteString(colorPunct)
				}
				sb.WriteString(",")
				if colorize {
					sb.WriteString(colorReset)
				}
			}

			if isPretty {
				sb.WriteString("\n")
			}
		}

		if isPretty {
			sb.WriteString(strings.Repeat(indent, depth))
		}

		if colorize {
			sb.WriteString(colorPunct)
		}
		sb.WriteString("]")
		if colorize {
			sb.WriteString(colorReset)
		}

	case TypeObject:
		if len(v.Obj) == 0 {
			if colorize {
				sb.WriteString(colorPunct)
			}
			sb.WriteString("{}")
			if colorize {
				sb.WriteString(colorReset)
			}
			return
		}

		if colorize {
			sb.WriteString(colorPunct)
		}
		sb.WriteString("{")
		if colorize {
			sb.WriteString(colorReset)
		}

		if isPretty {
			sb.WriteString("\n")
		}

		for i := 0; i < len(v.Obj); i++ {
			member := v.Obj[i]

			if isPretty {
				sb.WriteString(strings.Repeat(indent, depth+1))
			}

			// Key
			if colorize {
				sb.WriteString(colorKey)
			}
			sb.WriteString(escapeString(member.Key))
			if colorize {
				sb.WriteString(colorReset)
			}

			// Colon
			if colorize {
				sb.WriteString(colorPunct)
			}
			sb.WriteString(":")
			if colorize {
				sb.WriteString(colorReset)
			}
			if isPretty {
				sb.WriteString(" ")
			}

			// Value
			formatValue(member.Val, sb, indent, depth+1, colorize)

			// Comma
			if i < len(v.Obj)-1 {
				if colorize {
					sb.WriteString(colorPunct)
				}
				sb.WriteString(",")
				if colorize {
					sb.WriteString(colorReset)
				}
			}

			if isPretty {
				sb.WriteString("\n")
			}
		}

		if isPretty {
			sb.WriteString(strings.Repeat(indent, depth))
		}

		if colorize {
			sb.WriteString(colorPunct)
		}
		sb.WriteString("}")
		if colorize {
			sb.WriteString(colorReset)
		}
	}
}
