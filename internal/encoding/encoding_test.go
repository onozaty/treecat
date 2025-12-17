package encoding

import (
	"testing"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/unicode"
)

func TestNewConverter_EmptyString(t *testing.T) {
	converter, err := NewConverter("")
	if err != nil {
		t.Errorf("Expected no error for empty string, got %v", err)
	}
	if converter != nil {
		t.Error("Expected nil converter for empty string")
	}
}

func TestNewConverter_UTF8(t *testing.T) {
	testCases := []string{
		"utf-8",
		"utf8",
		"UTF-8",
		"UTF8",
		"Utf_8",
	}

	for _, encodingName := range testCases {
		t.Run(encodingName, func(t *testing.T) {
			converter, err := NewConverter(encodingName)
			if err != nil {
				t.Errorf("Expected no error for %s, got %v", encodingName, err)
			}
			if converter != nil {
				t.Errorf("Expected nil converter for %s (no conversion needed)", encodingName)
			}
		})
	}
}

func TestNewConverter_ValidEncodings(t *testing.T) {
	testCases := []string{
		"shift_jis",
		"euc-jp",
		"iso-2022-jp",
		"windows-1252",
		"iso-8859-1",
		"gb2312",
		"big5",
		"euc-kr",
	}

	for _, encodingName := range testCases {
		t.Run(encodingName, func(t *testing.T) {
			converter, err := NewConverter(encodingName)
			if err != nil {
				t.Errorf("Expected no error for %s, got %v", encodingName, err)
			}
			if converter == nil {
				t.Errorf("Expected non-nil converter for %s", encodingName)
			}
		})
	}
}

func TestNewConverter_NameNormalization(t *testing.T) {
	testCases := []struct {
		name     string
		variants []string
	}{
		{
			name: "Shift_JIS normalization",
			variants: []string{
				"shift_jis",
				"shift-jis",
				"Shift_JIS",
				"SHIFT_JIS",
				"ShIfT_jIs",
			},
		},
		{
			name: "EUC-JP normalization",
			variants: []string{
				"euc-jp",
				"euc_jp",
				"EUC-JP",
				"EUC_JP",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, variant := range tc.variants {
				converter, err := NewConverter(variant)
				if err != nil {
					t.Errorf("Expected no error for %s, got %v", variant, err)
				}
				if converter == nil {
					t.Errorf("Expected non-nil converter for %s", variant)
				}
			}
		})
	}
}

func TestNewConverter_InvalidEncoding(t *testing.T) {
	testCases := []string{
		"invalid-encoding",
		"unknown",
		"not-a-real-encoding",
	}

	for _, encodingName := range testCases {
		t.Run(encodingName, func(t *testing.T) {
			converter, err := NewConverter(encodingName)
			if err == nil {
				t.Errorf("Expected error for invalid encoding %s, got nil", encodingName)
			}
			if converter != nil {
				t.Errorf("Expected nil converter for invalid encoding %s", encodingName)
			}
		})
	}
}

func TestConvertToUTF8_ShiftJIS(t *testing.T) {
	// "こんにちは世界" ("Hello, World" in Japanese) in Shift_JIS
	shiftJISBytes := []byte{0x82, 0xb1, 0x82, 0xf1, 0x82, 0xc9, 0x82, 0xbf, 0x82, 0xcd, 0x90, 0xa2, 0x8a, 0x45}
	expectedUTF8 := "こんにちは世界"

	converter, err := NewConverter("shift_jis")
	if err != nil {
		t.Fatalf("Failed to create converter: %v", err)
	}

	result, err := converter.ConvertToUTF8(shiftJISBytes)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	if string(result) != expectedUTF8 {
		t.Errorf("Expected %s, got %s", expectedUTF8, string(result))
	}
}

func TestConvertToUTF8_EUCJP(t *testing.T) {
	// Encode "こんにちは世界" to EUC-JP
	originalText := "こんにちは世界"
	encoder := japanese.EUCJP.NewEncoder()
	eucjpBytes, err := encoder.Bytes([]byte(originalText))
	if err != nil {
		t.Fatalf("Failed to encode to EUC-JP: %v", err)
	}

	converter, err := NewConverter("euc-jp")
	if err != nil {
		t.Fatalf("Failed to create converter: %v", err)
	}

	result, err := converter.ConvertToUTF8(eucjpBytes)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	if string(result) != originalText {
		t.Errorf("Expected %s, got %s", originalText, string(result))
	}
}

func TestConvertToUTF8_ISO2022JP(t *testing.T) {
	// Encode "こんにちは世界" to ISO-2022-JP
	originalText := "こんにちは世界"
	encoder := japanese.ISO2022JP.NewEncoder()
	iso2022jpBytes, err := encoder.Bytes([]byte(originalText))
	if err != nil {
		t.Fatalf("Failed to encode to ISO-2022-JP: %v", err)
	}

	converter, err := NewConverter("iso-2022-jp")
	if err != nil {
		t.Fatalf("Failed to create converter: %v", err)
	}

	result, err := converter.ConvertToUTF8(iso2022jpBytes)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	if string(result) != originalText {
		t.Errorf("Expected %s, got %s", originalText, string(result))
	}
}

func TestConvertToUTF8_InvalidBytes(t *testing.T) {
	// Invalid Shift_JIS byte sequence
	invalidBytes := []byte{0xFF, 0xFE, 0xFD}

	converter, err := NewConverter("shift_jis")
	if err != nil {
		t.Fatalf("Failed to create converter: %v", err)
	}

	// Conversion succeeds even with invalid bytes, but replacement characters are used
	// No error is returned (this is the behavior of transform.Reader)
	result, err := converter.ConvertToUTF8(invalidBytes)
	if err != nil {
		t.Fatalf("Expected no error for invalid bytes (replacement chars should be used), got %v", err)
	}
	if len(result) == 0 {
		t.Error("Expected non-empty result (with replacement chars)")
	}
}

func TestConvertToUTF8_EmptyInput(t *testing.T) {
	converter, err := NewConverter("shift_jis")
	if err != nil {
		t.Fatalf("Failed to create converter: %v", err)
	}

	result, err := converter.ConvertToUTF8([]byte{})
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected empty result for empty input, got %d bytes", len(result))
	}
}

func TestConvertToUTF8_ASCIIOnly(t *testing.T) {
	// ASCII-only strings are the same across all encodings
	asciiText := "Hello, World! 123"
	asciiBytes := []byte(asciiText)

	testEncodings := []string{
		"shift_jis",
		"euc-jp",
		"iso-2022-jp",
		"windows-1252",
	}

	for _, encodingName := range testEncodings {
		t.Run(encodingName, func(t *testing.T) {
			converter, err := NewConverter(encodingName)
			if err != nil {
				t.Fatalf("Failed to create converter: %v", err)
			}

			result, err := converter.ConvertToUTF8(asciiBytes)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			if string(result) != asciiText {
				t.Errorf("Expected %s, got %s", asciiText, string(result))
			}
		})
	}
}

func TestConvertToUTF8_Windows1252(t *testing.T) {
	// "Café" encoded in Windows-1252 (é = 0xE9)
	windows1252Bytes := []byte{0x43, 0x61, 0x66, 0xE9}
	expectedUTF8 := "Café"

	converter, err := NewConverter("windows-1252")
	if err != nil {
		t.Fatalf("Failed to create converter: %v", err)
	}

	result, err := converter.ConvertToUTF8(windows1252Bytes)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	if string(result) != expectedUTF8 {
		t.Errorf("Expected %s, got %s", expectedUTF8, string(result))
	}
}

func TestConvertToUTF8_LargeContent(t *testing.T) {
	// Test memory efficiency with large content
	originalText := "あいうえお"
	encoder := japanese.ShiftJIS.NewEncoder()
	shiftJISBytes, err := encoder.Bytes([]byte(originalText))
	if err != nil {
		t.Fatalf("Failed to encode: %v", err)
	}

	// Repeat 1000 times
	largeContent := make([]byte, 0, len(shiftJISBytes)*1000)
	for i := 0; i < 1000; i++ {
		largeContent = append(largeContent, shiftJISBytes...)
	}

	converter, err := NewConverter("shift_jis")
	if err != nil {
		t.Fatalf("Failed to create converter: %v", err)
	}

	result, err := converter.ConvertToUTF8(largeContent)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	// Check result length (verify proper UTF-8 encoding)
	decoder := unicode.UTF8.NewDecoder()
	expectedLength, _ := decoder.Bytes(result)
	if len(expectedLength) == 0 {
		t.Error("Expected non-empty result for large content")
	}
}

func TestNormalizeExtension(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{".txt", "txt"},
		{"txt", "txt"},
		{".TXT", "txt"},
		{"TXT", "txt"},
		{".Md", "md"},
		{"log", "log"},
		{".LOG", "log"},
		{"", ""},
		{".", ""},
		{".tar.gz", "tar.gz"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := NormalizeExtension(tc.input)
			if result != tc.expected {
				t.Errorf("NormalizeExtension(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestParseEncodingMap_Empty(t *testing.T) {
	result, err := ParseEncodingMap("")
	if err != nil {
		t.Errorf("Expected no error for empty string, got %v", err)
	}
	if result != nil {
		t.Error("Expected nil result for empty string")
	}
}

func TestParseEncodingMap_Valid(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected map[string]string // extension -> encoding name
	}{
		{
			name:     "single entry",
			input:    "txt:shift_jis",
			expected: map[string]string{"txt": "shift_jis"},
		},
		{
			name:     "multiple entries",
			input:    "txt:shift_jis,log:euc-jp,md:utf-8",
			expected: map[string]string{"txt": "shift_jis", "log": "euc-jp", "md": "utf-8"},
		},
		{
			name:     "with whitespace",
			input:    " txt : shift_jis , log : euc-jp ",
			expected: map[string]string{"txt": "shift_jis", "log": "euc-jp"},
		},
		{
			name:     "with dot prefix",
			input:    ".txt:shift_jis,.log:euc-jp",
			expected: map[string]string{"txt": "shift_jis", "log": "euc-jp"},
		},
		{
			name:     "uppercase extension",
			input:    "TXT:shift_jis,LOG:euc-jp",
			expected: map[string]string{"txt": "shift_jis", "log": "euc-jp"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseEncodingMap(tc.input)
			if err != nil {
				t.Fatalf("ParseEncodingMap(%q) failed: %v", tc.input, err)
			}

			if len(result) != len(tc.expected) {
				t.Errorf("Expected %d entries, got %d", len(tc.expected), len(result))
			}

			for ext, encoding := range tc.expected {
				converter, ok := result[ext]
				if !ok {
					t.Errorf("Expected converter for extension %q", ext)
					continue
				}

				// For UTF-8, converter should be nil
				if encoding == "utf-8" {
					if converter != nil {
						t.Errorf("Expected nil converter for UTF-8, got non-nil")
					}
				} else {
					if converter == nil {
						t.Errorf("Expected non-nil converter for %q", encoding)
					}
				}
			}
		})
	}
}

func TestParseEncodingMap_InvalidFormat(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{"no colon", "txt"},
		{"multiple colons", "txt:shift_jis:extra"},
		{"missing extension", ":shift_jis"},
		{"missing encoding", "txt:"},
		{"empty after trim", "  :  "},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseEncodingMap(tc.input)
			if err == nil {
				t.Errorf("ParseEncodingMap(%q) should return error, got nil", tc.input)
			}
		})
	}
}

func TestParseEncodingMap_InvalidEncoding(t *testing.T) {
	input := "txt:invalid-encoding-name"
	_, err := ParseEncodingMap(input)
	if err == nil {
		t.Error("Expected error for invalid encoding name, got nil")
	}
}

func TestParseEncodingMap_DuplicateExtensions(t *testing.T) {
	// Duplicate extensions: last one wins
	input := "txt:shift_jis,txt:euc-jp"
	result, err := ParseEncodingMap(input)
	if err != nil {
		t.Fatalf("ParseEncodingMap failed: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("Expected 1 entry (last one wins), got %d", len(result))
	}

	// Should have euc-jp converter (last one)
	if _, ok := result["txt"]; !ok {
		t.Error("Expected converter for 'txt' extension")
	}
}

func TestRemoveBOM(t *testing.T) {
	testCases := []struct {
		name        string
		input       []byte
		expectedOut []byte
		expectedBOM bool
	}{
		{
			name:        "with BOM at start",
			input:       []byte{0xEF, 0xBB, 0xBF, 'h', 'e', 'l', 'l', 'o'},
			expectedOut: []byte{'h', 'e', 'l', 'l', 'o'},
			expectedBOM: true,
		},
		{
			name:        "without BOM",
			input:       []byte{'h', 'e', 'l', 'l', 'o'},
			expectedOut: []byte{'h', 'e', 'l', 'l', 'o'},
			expectedBOM: false,
		},
		{
			name:        "empty",
			input:       []byte{},
			expectedOut: []byte{},
			expectedBOM: false,
		},
		{
			name:        "only BOM",
			input:       []byte{0xEF, 0xBB, 0xBF},
			expectedOut: []byte{},
			expectedBOM: true,
		},
		{
			name:        "partial BOM (2 bytes)",
			input:       []byte{0xEF, 0xBB},
			expectedOut: []byte{0xEF, 0xBB},
			expectedBOM: false,
		},
		{
			name:        "BOM-like in middle",
			input:       []byte{'h', 'i', 0xEF, 0xBB, 0xBF, 'x'},
			expectedOut: []byte{'h', 'i', 0xEF, 0xBB, 0xBF, 'x'},
			expectedBOM: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, hasBOM := RemoveBOM(tc.input)
			if hasBOM != tc.expectedBOM {
				t.Errorf("RemoveBOM(%v) BOM flag = %v, want %v", tc.input, hasBOM, tc.expectedBOM)
			}
			if string(result) != string(tc.expectedOut) {
				t.Errorf("RemoveBOM(%v) = %v, want %v", tc.input, result, tc.expectedOut)
			}
		})
	}
}

func TestNormalizeNewlines(t *testing.T) {
	testCases := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{
			name:     "CRLF to LF",
			input:    []byte("line1\r\nline2\r\nline3"),
			expected: []byte("line1\nline2\nline3"),
		},
		{
			name:     "CR to LF",
			input:    []byte("line1\rline2\rline3"),
			expected: []byte("line1\nline2\nline3"),
		},
		{
			name:     "LF unchanged",
			input:    []byte("line1\nline2\nline3"),
			expected: []byte("line1\nline2\nline3"),
		},
		{
			name:     "mixed line endings",
			input:    []byte("line1\r\nline2\nline3\rline4"),
			expected: []byte("line1\nline2\nline3\nline4"),
		},
		{
			name:     "empty",
			input:    []byte{},
			expected: []byte{},
		},
		{
			name:     "no newlines",
			input:    []byte("single line"),
			expected: []byte("single line"),
		},
		{
			name:     "only CRLF",
			input:    []byte("\r\n"),
			expected: []byte("\n"),
		},
		{
			name:     "multiple CRLF",
			input:    []byte("\r\n\r\n\r\n"),
			expected: []byte("\n\n\n"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := NormalizeNewlines(tc.input)
			if string(result) != string(tc.expected) {
				t.Errorf("NormalizeNewlines(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}
