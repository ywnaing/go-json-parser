package main

import (
	"fmt"
)

func runTestCase(title string, input string) {
	fmt.Printf("==================================================\n")
	fmt.Printf("TEST: %s\n", title)
	fmt.Printf("Raw JSON: %s\n", input)
	fmt.Printf("--------------------------------------------------\n")

	// 1. Lexer tokens demo
	l := NewLexer(input)
	for {
		token := l.Next()
		val := token.Val(input)
		fmt.Printf("  %-16s -> %-20q [Start: %d, End: %d]\n", token.AsString(), val, token.Start, token.End)
		if token.Type == TOKEN_INVALID || token.Type == TOKEN_EOF {
			break
		}
	}

	// 2. Parser & Serializer demo
	p := NewParser(input)
	parsed, err := p.Parse()
	if err != nil {
		fmt.Printf("  [Parse Error]: %v\n", err)
	} else {
		fmt.Printf("  [Minified]   : %s\n", parsed.String())
		fmt.Printf("  [Pretty]\n%s\n", parsed.PrettyPrint("  "))
	}
	fmt.Printf("\n")
}

func main() {
	// Case 1: Standard JSON (Objects, Arrays, Booleans, Null)
	runTestCase(
		"Standard JSON",
		`{"name": "Tom", "age": 30, "is_active": true, "data": null}`,
	)

	// Case 2: String with Escaped Quotes and Newlines
	runTestCase(
		"Escaped Quotes & Newlines in String",
		`{"quote": "He said \"Hello\"", "multiline": "Line1\nLine2"}`,
	)

	// Case 3: URL with Escaped Forward Slash (\/)
	runTestCase(
		"Escaped Slash in String (\\/)",
		`{"url": "https:\/\/example.com\/api"}`,
	)

	// Case 4: Backslash in File Path (\\)
	runTestCase(
		"Escaped Backslash in String (\\\\)",
		`{"path": "C:\\Windows\\System32"}`,
	)

	// Case 5: Scientific Notation / Exponent Numbers (e.g. 1e10)
	runTestCase(
		"Scientific Notation Numbers (1e10)",
		`{"count": 1e10}`,
	)

	// Case 6: Unicode Escape Sequences (\u0041 for 'A')
	runTestCase(
		"Unicode Escape Sequence (\\u0041)",
		`{"letter": "\u0041\u0042"}`,
	)

	// Case 7: Nested Arrays & Objects
	runTestCase(
		"Nested Structures",
		`{"project": "JSON Parser", "scores": [100, 95.5], "tags": ["go", "streaming"]}`,
	)

	// Case 8: Lone Minus Sign (Expected Error)
	runTestCase(
		"Lone Minus Sign ('-')",
		`-`,
	)
}
