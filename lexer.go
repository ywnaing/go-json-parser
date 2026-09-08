package main

import (
	"strings"
)

type Lexer struct {
	Source  string
	Current Token
	Pos     int
}

// NewLexer creates a new Lexer instance initialized at the start of the source string.
func NewLexer(src string) *Lexer {
	return &Lexer{
		Source: src,
		Pos:    0,
	}
}

// IsEOF checks whether the lexer has reached the end of the input source.
func (l *Lexer) IsEOF() bool {
	return l.Pos >= len(l.Source)
}

// Peek returns the byte at the current position, or 0 if at EOF without advancing.
func (l *Lexer) Peek() byte {
	if l.IsEOF() {
		return 0
	}
	return l.Source[l.Pos]
}

// Advance increments the current position by 1 if not at EOF.
func (l *Lexer) Advance() {
	if !l.IsEOF() {
		l.Pos++
	}
}

// IsDigit returns true if c is an ASCII decimal digit ('0'-'9').
func IsDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

// IsHex returns true if c is an ASCII hexadecimal digit ('0'-'9', 'a'-'f', 'A'-'F').
func IsHex(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

// IsHex checks if the current character is a valid hexadecimal digit.
func (l *Lexer) IsHex() bool {
	return IsHex(l.Peek())
}

// ScanString parses a JSON string literal starting at the opening quote.
// Handles valid JSON escape sequences: \", \\, \/, \b, \f, \n, \r, \t, and \uXXXX.
// Control characters (ASCII < 0x20) are disallowed inside string literals per RFC 8259.
func (l *Lexer) ScanString() Token {
	startPos := l.Pos
	l.Advance() // Consume opening quote '"'

	for !l.IsEOF() && l.Peek() != '"' {
		c := l.Peek()

		// Disallow raw ASCII control characters (< 0x20) in JSON strings
		if c < 0x20 {
			return Token{
				Type:  TOKEN_INVALID,
				Start: startPos,
				End:   l.Pos,
			}
		}

		if c == '\\' {
			l.Advance() // Consume backslash '\'

			switch l.Peek() {
			case '"', '\\', '/', 'b', 'f', 'n', 'r', 't':
				l.Advance()
			case 'u':
				l.Advance() // Consume 'u'
				for i := 0; i < 4; i++ {
					if l.IsHex() {
						l.Advance()
					} else {
						return Token{
							Type:  TOKEN_INVALID,
							Start: startPos,
							End:   l.Pos,
						}
					}
				}
			default:
				return Token{
					Type:  TOKEN_INVALID,
					Start: startPos,
					End:   l.Pos,
				}
			}
		} else {
			l.Advance()
		}
	}

	if l.Peek() != '"' {
		return Token{
			Type:  TOKEN_INVALID,
			Start: startPos,
			End:   l.Pos,
		}
	}

	l.Advance() // Consume closing quote '"'

	return Token{
		Type:  TOKEN_STRING,
		Start: startPos,
		End:   l.Pos,
	}
}

// ScanKeyword checks if the current slice matches keyword. If matched, it advances
// the lexer by the keyword length and returns tokenType. Otherwise, it returns TOKEN_INVALID.
func (l *Lexer) ScanKeyword(keyword string, tokenType TokenType) Token {
	startPos := l.Pos
	if strings.HasPrefix(l.Source[l.Pos:], keyword) {
		l.Pos += len(keyword)
		return Token{
			Type:  tokenType,
			Start: startPos,
			End:   l.Pos,
		}
	}

	l.Advance()
	return Token{
		Type:  TOKEN_INVALID,
		Start: startPos,
		End:   l.Pos,
	}
}

// number = [ minus ] int [ frac ] [ exp ]
func (l *Lexer) ScanNumber() Token {
	startPos := l.Pos

	// Optional minus sign '-'
	if l.Peek() == '-' {
		l.Advance()
	}

	// Integer part: either '0' or '1'-'9' followed by digits
	if l.Peek() == '0' {
		l.Advance()
		// Leading zero cannot be followed by another digit (e.g. 01, 00 are invalid)
		if IsDigit(l.Peek()) {
			return Token{
				Type:  TOKEN_INVALID,
				Start: startPos,
				End:   l.Pos,
			}
		}
	} else if l.Peek() >= '1' && l.Peek() <= '9' {
		l.Advance()
		for IsDigit(l.Peek()) {
			l.Advance()
		}
	} else {
		// Must have at least one digit ('-' or non-digits)
		return Token{
			Type:  TOKEN_INVALID,
			Start: startPos,
			End:   l.Pos,
		}
	}

	// Fractional part (optional): '.' followed by one or more digits
	if l.Peek() == '.' {
		l.Advance()
		if !IsDigit(l.Peek()) {
			return Token{
				Type:  TOKEN_INVALID,
				Start: startPos,
				End:   l.Pos,
			}
		}
		for IsDigit(l.Peek()) {
			l.Advance()
		}
	}

	// Exponent part (optional): ('e' | 'E') followed by optional ('+' | '-') and one or more digits
	if l.Peek() == 'e' || l.Peek() == 'E' {
		l.Advance()
		if l.Peek() == '+' || l.Peek() == '-' {
			l.Advance()
		}
		if !IsDigit(l.Peek()) {
			return Token{
				Type:  TOKEN_INVALID,
				Start: startPos,
				End:   l.Pos,
			}
		}
		for IsDigit(l.Peek()) {
			l.Advance()
		}
	}

	return Token{
		Type:  TOKEN_NUMBER,
		Start: startPos,
		End:   l.Pos,
	}
}

// SkipWhitespace skips standard JSON whitespace characters (' ', '\n', '\t', '\r').
func (l *Lexer) SkipWhitespace() {
	for {
		switch l.Peek() {
		case ' ', '\n', '\t', '\r':
			l.Advance()
		default:
			return
		}
	}
}

// Next tokenizes and returns the next token from the source.
func (l *Lexer) Next() Token {
	l.SkipWhitespace()

	startPos := l.Pos

	if l.IsEOF() {
		token := Token{
			Type:  TOKEN_EOF,
			Start: startPos,
			End:   startPos,
		}
		l.Current = token
		return token
	}

	var token Token

	switch l.Peek() {
	case '{':
		l.Advance()
		token = Token{
			Type:  TOKEN_LBRACE,
			Start: startPos,
			End:   l.Pos,
		}
	case '}':
		l.Advance()
		token = Token{
			Type:  TOKEN_RBRACE,
			Start: startPos,
			End:   l.Pos,
		}
	case '[':
		l.Advance()
		token = Token{
			Type:  TOKEN_LBRACKET,
			Start: startPos,
			End:   l.Pos,
		}
	case ']':
		l.Advance()
		token = Token{
			Type:  TOKEN_RBRACKET,
			Start: startPos,
			End:   l.Pos,
		}
	case ':':
		l.Advance()
		token = Token{
			Type:  TOKEN_COLON,
			Start: startPos,
			End:   l.Pos,
		}
	case ',':
		l.Advance()
		token = Token{
			Type:  TOKEN_COMMA,
			Start: startPos,
			End:   l.Pos,
		}
	case '"':
		token = l.ScanString()
	case 't':
		token = l.ScanKeyword("true", TOKEN_TRUE)
	case 'f':
		token = l.ScanKeyword("false", TOKEN_FALSE)
	case 'n':
		token = l.ScanKeyword("null", TOKEN_NULL)
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		token = l.ScanNumber()
	default:
		l.Advance()
		token = Token{
			Type:  TOKEN_INVALID,
			Start: startPos,
			End:   l.Pos,
		}
	}

	l.Current = token
	return token
}
