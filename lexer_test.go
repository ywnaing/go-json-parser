package main

import (
	"testing"
)

func TestLexerBasicTokens(t *testing.T) {
	input := `{} [] : ,`
	expectedTypes := []TokenType{
		TOKEN_LBRACE,
		TOKEN_RBRACE,
		TOKEN_LBRACKET,
		TOKEN_RBRACKET,
		TOKEN_COLON,
		TOKEN_COMMA,
		TOKEN_EOF,
	}

	l := NewLexer(input)
	for i, expectedType := range expectedTypes {
		token := l.Next()
		if token.Type != expectedType {
			t.Fatalf("step %d: expected type %v (%s), got %v (%s)",
				i, expectedType, expectedTypeString(expectedType), token.Type, token.AsString())
		}
	}
}

func TestLexerKeywords(t *testing.T) {
	input := `true false null`
	expected := []struct {
		tokenType TokenType
		val       string
	}{
		{TOKEN_TRUE, "true"},
		{TOKEN_FALSE, "false"},
		{TOKEN_NULL, "null"},
		{TOKEN_EOF, ""},
	}

	l := NewLexer(input)
	for i, exp := range expected {
		token := l.Next()
		if token.Type != exp.tokenType {
			t.Fatalf("step %d: expected %v, got %v", i, exp.tokenType, token.Type)
		}
		if val := token.Val(input); val != exp.val {
			t.Fatalf("step %d: expected val %q, got %q", i, exp.val, val)
		}
	}
}

func TestLexerStrings(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantType  TokenType
		wantVal   string
	}{
		{"empty string", `""`, TOKEN_STRING, `""`},
		{"simple string", `"hello world"`, TOKEN_STRING, `"hello world"`},
		{"escaped quotes", `"hello \"world\""`, TOKEN_STRING, `"hello \"world\""`},
		{"escaped backslash", `"C:\\path\\file"`, TOKEN_STRING, `"C:\\path\\file"`},
		{"escaped slash", `"https:\/\/example.com"`, TOKEN_STRING, `"https:\/\/example.com"`},
		{"control escapes", `"\b\f\n\r\t"`, TOKEN_STRING, `"\b\f\n\r\t"`},
		{"unicode escape lowercase", `"\u0041\u0066"`, TOKEN_STRING, `"\u0041\u0066"`},
		{"unicode escape uppercase", `"\u004F\uABCD"`, TOKEN_STRING, `"\u004F\uABCD"`},
		{"unclosed string", `"unclosed`, TOKEN_INVALID, ""},
		{"invalid escape", `"\a"`, TOKEN_INVALID, ""},
		{"invalid unicode short", `"\u12"`, TOKEN_INVALID, ""},
		{"invalid unicode non-hex", `"\u004G"`, TOKEN_INVALID, ""},
		{"unescaped control char", "\"\n\"", TOKEN_INVALID, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLexer(tt.input)
			tok := l.Next()
			if tok.Type != tt.wantType {
				t.Errorf("input %q: expected type %v, got %v", tt.input, tt.wantType, tok.Type)
			}
			if tt.wantType != TOKEN_INVALID && tok.Val(tt.input) != tt.wantVal {
				t.Errorf("input %q: expected val %q, got %q", tt.input, tt.wantVal, tok.Val(tt.input))
			}
		})
	}
}

func TestLexerNumbers(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantType TokenType
		wantVal  string
	}{
		{"zero", `0`, TOKEN_NUMBER, `0`},
		{"negative zero", `-0`, TOKEN_NUMBER, `-0`},
		{"positive integer", `12345`, TOKEN_NUMBER, `12345`},
		{"negative integer", `-987`, TOKEN_NUMBER, `-987`},
		{"float", `3.14159`, TOKEN_NUMBER, `3.14159`},
		{"negative float", `-0.123`, TOKEN_NUMBER, `-0.123`},
		{"exponent lowercase", `1e10`, TOKEN_NUMBER, `1e10`},
		{"exponent uppercase", `1E10`, TOKEN_NUMBER, `1E10`},
		{"exponent plus", `2.5e+3`, TOKEN_NUMBER, `2.5e+3`},
		{"exponent minus", `2.5e-3`, TOKEN_NUMBER, `2.5e-3`},
		{"lone minus", `-`, TOKEN_INVALID, ""},
		{"leading zero with digits", `0123`, TOKEN_INVALID, ""},
		{"trailing dot", `1.`, TOKEN_INVALID, ""},
		{"incomplete exponent", `1e`, TOKEN_INVALID, ""},
		{"incomplete exponent with sign", `1e+`, TOKEN_INVALID, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLexer(tt.input)
			tok := l.Next()
			if tok.Type != tt.wantType {
				t.Errorf("input %q: expected type %v, got %v", tt.input, tt.wantType, tok.Type)
			}
			if tt.wantType != TOKEN_INVALID && tok.Val(tt.input) != tt.wantVal {
				t.Errorf("input %q: expected val %q, got %q", tt.input, tt.wantVal, tok.Val(tt.input))
			}
		})
	}
}

func TestLexerComplexJSON(t *testing.T) {
	input := `{
		"id": 101,
		"name": "Antigravity",
		"active": true,
		"tags": ["ai", "golang"],
		"rating": 4.95,
		"meta": null
	}`

	expectedSequence := []struct {
		tokenType TokenType
		val       string
	}{
		{TOKEN_LBRACE, "{"},
		{TOKEN_STRING, `"id"`},
		{TOKEN_COLON, ":"},
		{TOKEN_NUMBER, "101"},
		{TOKEN_COMMA, ","},
		{TOKEN_STRING, `"name"`},
		{TOKEN_COLON, ":"},
		{TOKEN_STRING, `"Antigravity"`},
		{TOKEN_COMMA, ","},
		{TOKEN_STRING, `"active"`},
		{TOKEN_COLON, ":"},
		{TOKEN_TRUE, "true"},
		{TOKEN_COMMA, ","},
		{TOKEN_STRING, `"tags"`},
		{TOKEN_COLON, ":"},
		{TOKEN_LBRACKET, "["},
		{TOKEN_STRING, `"ai"`},
		{TOKEN_COMMA, ","},
		{TOKEN_STRING, `"golang"`},
		{TOKEN_RBRACKET, "]"},
		{TOKEN_COMMA, ","},
		{TOKEN_STRING, `"rating"`},
		{TOKEN_COLON, ":"},
		{TOKEN_NUMBER, "4.95"},
		{TOKEN_COMMA, ","},
		{TOKEN_STRING, `"meta"`},
		{TOKEN_COLON, ":"},
		{TOKEN_NULL, "null"},
		{TOKEN_RBRACE, "}"},
		{TOKEN_EOF, ""},
	}

	l := NewLexer(input)
	for i, exp := range expectedSequence {
		tok := l.Next()
		if tok.Type != exp.tokenType {
			t.Fatalf("step %d: expected type %v (%s), got %v (%s)",
				i, exp.tokenType, expectedTypeString(exp.tokenType), tok.Type, tok.AsString())
		}
		if tok.Type != TOKEN_EOF && tok.Val(input) != exp.val {
			t.Fatalf("step %d: expected val %q, got %q", i, exp.val, tok.Val(input))
		}
	}
}

func expectedTypeString(t TokenType) string {
	tok := Token{Type: t}
	return tok.AsString()
}
