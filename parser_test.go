package main

import (
	"testing"
)

func TestParsePrimitives(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantType ValueType
		check    func(t *testing.T, v Value)
	}{
		{
			name:     "string",
			input:    `"hello world"`,
			wantType: TypeString,
			check: func(t *testing.T, v Value) {
				if v.Str != "hello world" {
					t.Errorf("expected %q, got %q", "hello world", v.Str)
				}
			},
		},
		{
			name:     "integer",
			input:    `42`,
			wantType: TypeNumber,
			check: func(t *testing.T, v Value) {
				if v.Num != 42 {
					t.Errorf("expected 42, got %v", v.Num)
				}
			},
		},
		{
			name:     "float",
			input:    `3.14159`,
			wantType: TypeNumber,
			check: func(t *testing.T, v Value) {
				if v.Num != 3.14159 {
					t.Errorf("expected 3.14159, got %v", v.Num)
				}
			},
		},
		{
			name:     "negative float",
			input:    `-12.5`,
			wantType: TypeNumber,
			check: func(t *testing.T, v Value) {
				if v.Num != -12.5 {
					t.Errorf("expected -12.5, got %v", v.Num)
				}
			},
		},
		{
			name:     "scientific notation",
			input:    `1e4`,
			wantType: TypeNumber,
			check: func(t *testing.T, v Value) {
				if v.Num != 10000 {
					t.Errorf("expected 10000, got %v", v.Num)
				}
			},
		},
		{
			name:     "boolean true",
			input:    `true`,
			wantType: TypeBool,
			check: func(t *testing.T, v Value) {
				if !v.Bol {
					t.Errorf("expected true, got false")
				}
			},
		},
		{
			name:     "boolean false",
			input:    `false`,
			wantType: TypeBool,
			check: func(t *testing.T, v Value) {
				if v.Bol {
					t.Errorf("expected false, got true")
				}
			},
		},
		{
			name:     "null",
			input:    `null`,
			wantType: TypeNull,
			check: func(t *testing.T, v Value) {
				if v.Type != TypeNull {
					t.Errorf("expected TypeNull, got %v", v.Type)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.input)
			val, err := p.Parse()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if val.Type != tt.wantType {
				t.Fatalf("expected type %v (%s), got %v (%s)",
					tt.wantType, tt.wantType.String(), val.Type, val.Type.String())
			}
			if tt.check != nil {
				tt.check(t, val)
			}
		})
	}
}

func TestParseEscapedStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantType ValueType
		wantStr  string
	}{
		{
			name:     "escaped newline and tab",
			input:    `"Line1\nLine2\tTabbed"`,
			wantType: TypeString,
			wantStr:  "Line1\nLine2\tTabbed",
		},
		{
			name:     "escaped quotes and backslash",
			input:    `"He said \"Hello\" and C:\\path"`,
			wantType: TypeString,
			wantStr:  "He said \"Hello\" and C:\\path",
		},
		{
			name:     "escaped slash and other control escapes",
			input:    `"https:\/\/example.com\b\f\r"`,
			wantType: TypeString,
			wantStr:  "https://example.com\b\f\r",
		},
		{
			name:     "unicode escape",
			input:    `"\u0041\u0042\u0043"`,
			wantType: TypeString,
			wantStr:  "ABC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.input)
			val, err := p.Parse()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if val.Type != tt.wantType {
				t.Fatalf("expected type %v, got %v", tt.wantType, val.Type)
			}
			if val.Str != tt.wantStr {
				t.Errorf("expected %q, got %q", tt.wantStr, val.Str)
			}
		})
	}
}

func TestParseEmptyStructures(t *testing.T) {
	t.Run("empty object", func(t *testing.T) {
		p := NewParser(`{}`)
		val, err := p.Parse()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val.Type != TypeObject {
			t.Fatalf("expected TypeObject, got %v", val.Type)
		}
		if len(val.Obj) != 0 {
			t.Errorf("expected 0 members, got %d", len(val.Obj))
		}
	})

	t.Run("empty array", func(t *testing.T) {
		p := NewParser(`[]`)
		val, err := p.Parse()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val.Type != TypeArray {
			t.Fatalf("expected TypeArray, got %v", val.Type)
		}
		if len(val.Arr) != 0 {
			t.Errorf("expected 0 elements, got %d", len(val.Arr))
		}
	})
}

func TestParseSimpleObject(t *testing.T) {
	input := `{"name": "Alice", "age": 30, "admin": true, "data": null}`
	p := NewParser(input)
	val, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if val.Type != TypeObject {
		t.Fatalf("expected TypeObject, got %v", val.Type)
	}

	if len(val.Obj) != 4 {
		t.Fatalf("expected 4 members, got %d", len(val.Obj))
	}

	// 1. name -> "Alice"
	if val.Obj[0].Key != "name" || val.Obj[0].Val.Type != TypeString || val.Obj[0].Val.Str != "Alice" {
		t.Errorf("member 0 mismatch: %+v", val.Obj[0])
	}

	// 2. age -> 30
	if val.Obj[1].Key != "age" || val.Obj[1].Val.Type != TypeNumber || val.Obj[1].Val.Num != 30 {
		t.Errorf("member 1 mismatch: %+v", val.Obj[1])
	}

	// 3. admin -> true
	if val.Obj[2].Key != "admin" || val.Obj[2].Val.Type != TypeBool || !val.Obj[2].Val.Bol {
		t.Errorf("member 2 mismatch: %+v", val.Obj[2])
	}

	// 4. data -> null
	if val.Obj[3].Key != "data" || val.Obj[3].Val.Type != TypeNull {
		t.Errorf("member 3 mismatch: %+v", val.Obj[3])
	}
}

func TestParseSimpleArray(t *testing.T) {
	input := `[10, "hello", false, null]`
	p := NewParser(input)
	val, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if val.Type != TypeArray {
		t.Fatalf("expected TypeArray, got %v", val.Type)
	}

	if len(val.Arr) != 4 {
		t.Fatalf("expected 4 elements, got %d", len(val.Arr))
	}

	if val.Arr[0].Type != TypeNumber || val.Arr[0].Num != 10 {
		t.Errorf("elem 0 mismatch: %+v", val.Arr[0])
	}

	if val.Arr[1].Type != TypeString || val.Arr[1].Str != "hello" {
		t.Errorf("elem 1 mismatch: %+v", val.Arr[1])
	}

	if val.Arr[2].Type != TypeBool || val.Arr[2].Bol != false {
		t.Errorf("elem 2 mismatch: %+v", val.Arr[2])
	}

	if val.Arr[3].Type != TypeNull {
		t.Errorf("elem 3 mismatch: %+v", val.Arr[3])
	}
}

func TestParseNestedStructures(t *testing.T) {
	t.Run("nested array", func(t *testing.T) {
		input := `[[1, 2], [3, 4], [[5]]]`
		p := NewParser(input)
		val, err := p.Parse()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if val.Type != TypeArray || len(val.Arr) != 3 {
			t.Fatalf("expected array of len 3, got: %+v", val)
		}

		// Check [1, 2]
		arr0 := val.Arr[0]
		if arr0.Type != TypeArray || len(arr0.Arr) != 2 || arr0.Arr[0].Num != 1 || arr0.Arr[1].Num != 2 {
			t.Errorf("arr0 mismatch: %+v", arr0)
		}

		// Check [[5]]
		arr2 := val.Arr[2]
		if arr2.Type != TypeArray || len(arr2.Arr) != 1 || arr2.Arr[0].Type != TypeArray || len(arr2.Arr[0].Arr) != 1 || arr2.Arr[0].Arr[0].Num != 5 {
			t.Errorf("arr2 mismatch: %+v", arr2)
		}
	})

	t.Run("nested object and array", func(t *testing.T) {
		input := `{
			"title": "Stream Project",
			"tags": ["go", "parser"],
			"author": {
				"name": "Tom",
				"id": 7
			}
		}`
		p := NewParser(input)
		val, err := p.Parse()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if val.Type != TypeObject || len(val.Obj) != 3 {
			t.Fatalf("expected object with 3 members, got %+v", val)
		}

		// tags array
		tags := val.Obj[1]
		if tags.Key != "tags" || tags.Val.Type != TypeArray || len(tags.Val.Arr) != 2 {
			t.Errorf("tags mismatch: %+v", tags)
		}

		// author object
		author := val.Obj[2]
		if author.Key != "author" || author.Val.Type != TypeObject || len(author.Val.Obj) != 2 {
			t.Errorf("author mismatch: %+v", author)
		}
	})
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty input", ``},
		{"only whitespace", `   `},
		{"trailing tokens after valid JSON", `{"a": 1} 123`},
		{"trailing comma in object", `{"a": 1,}`},
		{"trailing comma in array", `[1, 2,]`},
		{"missing colon in object", `{"a" 1}`},
		{"missing value in object", `{"a":}`},
		{"missing value in array", `[1,]`},
		{"non-string key in object", `{123: "val"}`},
		{"unclosed object", `{"a": 1`},
		{"unclosed array", `[1, 2`},
		{"invalid keyword", `[foo]`},
		{"lone comma", `,`},
		{"lone colon", `:`},
		{"mismatched brackets", `{"a": [1, 2}}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.input)
			val, err := p.Parse()
			if err == nil {
				t.Errorf("expected error for input %q, but parsed successfully as %+v", tt.input, val)
			}
		})
	}
}

func TestValueHelpers(t *testing.T) {
	input := `{
		"name": "Tom",
		"age": 28,
		"height": 5.9,
		"is_active": true,
		"extra": null,
		"scores": [100, 95, 88]
	}`

	p := NewParser(input)
	root, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 1. Get & GetString
	nameVal, ok := root.Get("name")
	if !ok {
		t.Fatalf("expected to find key 'name'")
	}
	if name, ok := nameVal.GetString(); !ok || name != "Tom" {
		t.Errorf("expected GetString() == %q, got %q (ok=%v)", "Tom", name, ok)
	}

	// 2. GetInt
	ageVal, ok := root.Get("age")
	if !ok {
		t.Fatalf("expected to find key 'age'")
	}
	if age, ok := ageVal.GetInt(); !ok || age != 28 {
		t.Errorf("expected GetInt() == 28, got %d (ok=%v)", age, ok)
	}

	// 3. GetFloat
	heightVal, ok := root.Get("height")
	if !ok {
		t.Fatalf("expected to find key 'height'")
	}
	if height, ok := heightVal.GetFloat(); !ok || height != 5.9 {
		t.Errorf("expected GetFloat() == 5.9, got %f (ok=%v)", height, ok)
	}

	// 4. GetBool
	activeVal, ok := root.Get("is_active")
	if !ok {
		t.Fatalf("expected to find key 'is_active'")
	}
	if active, ok := activeVal.GetBool(); !ok || !active {
		t.Errorf("expected GetBool() == true, got %v (ok=%v)", active, ok)
	}

	// 5. IsNull
	extraVal, ok := root.Get("extra")
	if !ok {
		t.Fatalf("expected to find key 'extra'")
	}
	if !extraVal.IsNull() {
		t.Errorf("expected IsNull() == true")
	}

	// 6. Index on array
	scoresVal, ok := root.Get("scores")
	if !ok {
		t.Fatalf("expected to find key 'scores'")
	}
	firstScore, ok := scoresVal.Index(0)
	if !ok {
		t.Fatalf("expected to find index 0 in scores")
	}
	if score, ok := firstScore.GetInt(); !ok || score != 100 {
		t.Errorf("expected first score to be 100, got %d", score)
	}

	// 7. Missing keys and out of bounds
	if _, ok := root.Get("non_existent"); ok {
		t.Errorf("expected Get('non_existent') to return false")
	}
	if _, ok := scoresVal.Index(99); ok {
		t.Errorf("expected Index(99) to return false")
	}
	if _, ok := scoresVal.Index(-1); ok {
		t.Errorf("expected Index(-1) to return false")
	}
	if _, ok := nameVal.GetInt(); ok {
		t.Errorf("expected GetInt() on string to return false")
	}
}

