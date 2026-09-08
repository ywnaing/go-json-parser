package main

import (
	"fmt"
)

func runTestCase(title string, input string) {
	fmt.Printf("==================================================\n")
	fmt.Printf("TEST: %s\n", title)
	fmt.Printf("Raw JSON: %s\n", input)
	fmt.Printf("--------------------------------------------------\n")

	l := NewLexer(input)

	for {
		token := l.Next()
		val := token.Val(input)

		fmt.Printf("  %-16s -> %-20q [Start: %d, End: %d]\n", token.AsString(), val, token.Start, token.End)

		if token.Type == TOKEN_INVALID || token.Type == TOKEN_EOF {
			break
		}
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
	// Observe: token.Val() contains literal quotes and literal '\' 'n' characters
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
	// Verified behavior: Correctly handled as a valid TOKEN_STRING
	runTestCase(
		"Escaped Backslash in String (\\\\)",
		`{"path": "C:\\Windows\\System32"}`,
	)

	// Case 5: Scientific Notation / Exponent Numbers (e.g. 1e10, 2.5e-3)
	// Verified behavior: Correctly parsed as a valid TOKEN_NUMBER
	runTestCase(
		"Scientific Notation Numbers (1e10)",
		`{"count": 1e10}`,
	)

	// Case 6: Unicode Escape Sequences (\u0041 for 'A')
	// Observe: token.Val() contains the raw "\u0041" characters, not the decoded rune
	runTestCase(
		"Unicode Escape Sequence (\\u0041)",
		`{"letter": "\u0041\u0042"}`,
	)

	// Case 7: Lone Minus Sign
	// Verified behavior: Correctly rejected as TOKEN_INVALID
	runTestCase(
		"Lone Minus Sign ('-')",
		`-`,
	)
}
