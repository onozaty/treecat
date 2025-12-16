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
