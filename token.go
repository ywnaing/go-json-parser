package main

// {"name" : "yewintnaing" }
// {
// "name"
// }
// ":"
// ,
// [
// ]
// true
// false
// null
// number

type TokenType int

const (
	TOKEN_NULL TokenType = iota
	TOKEN_TRUE
	TOKEN_FALSE
	TOKEN_COLON
	TOKEN_COMMA
	TOKEN_LBRACE
	TOKEN_RBRACE
	TOKEN_LBRACKET
	TOKEN_RBRACKET
	TOKEN_NUMBER
	TOKEN_STRING
	TOKEN_INVALID
	TOKEN_EOF
)

func (t TokenType) AsString() string {
	switch t {
	case TOKEN_LBRACE:
		return "TOKEN_LBRACE"
	case TOKEN_RBRACE:
		return "TOKEN_RBRACE"
	case TOKEN_LBRACKET:
		return "TOKEN_LBRACKET"
	case TOKEN_RBRACKET:
		return "TOKEN_RBRACKET"
	case TOKEN_COLON:
		return "TOKEN_COLON"
	case TOKEN_COMMA:
		return "TOKEN_COMMA"
	case TOKEN_STRING:
		return "TOKEN_STRING"
	case TOKEN_NUMBER:
		return "TOKEN_NUMBER"
	case TOKEN_TRUE:
		return "TOKEN_TRUE"
	case TOKEN_FALSE:
		return "TOKEN_FALSE"
	case TOKEN_NULL:
		return "TOKEN_NULL"
	case TOKEN_INVALID:
		return "TOKEN_INVALID"
	case TOKEN_EOF:
		return "TOKEN_EOF"
	default:
		return "UNKNOWN"
	}
}

type Token struct {
	Type  TokenType
	Start int
	End   int
}

func (t Token) AsString() string {
	switch t.Type {
	case TOKEN_LBRACE:
		return "TOKEN_LBRACE"
	case TOKEN_RBRACE:
		return "TOKEN_RBRACE"
	case TOKEN_LBRACKET:
		return "TOKEN_LBRACKET"
	case TOKEN_RBRACKET:
		return "TOKEN_RBRACKET"
	case TOKEN_COLON:
		return "TOKEN_COLON"
	case TOKEN_COMMA:
		return "TOKEN_COMMA"
	case TOKEN_STRING:
		return "TOKEN_STRING"
	case TOKEN_NUMBER:
		return "TOKEN_NUMBER"
	case TOKEN_TRUE:
		return "TOKEN_TRUE"
	case TOKEN_FALSE:
		return "TOKEN_FALSE"
	case TOKEN_NULL:
		return "TOKEN_NULL"
	case TOKEN_INVALID:
		return "TOKEN_INVALID"
	case TOKEN_EOF:
		return "TOKEN_EOF"
	default:
		return ""
	}
}

func (t Token) Val(src string) string {
	if t.Start < 0 || t.End > len(src) || t.Start > t.End {
		return ""
	}
	return src[t.Start:t.End]
}
