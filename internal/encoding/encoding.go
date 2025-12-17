package encoding

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"
	"golang.org/x/text/transform"
)

// Converter is an interface for encoding conversion.
type Converter interface {
	ConvertToUTF8(content []byte) ([]byte, error)
}

// textConverter converts content from a specific encoding to UTF-8.
type textConverter struct {
	encodingName string
	decoder      *encoding.Decoder
}

// NewConverter creates a Converter from the specified encoding name.
// If encodingName is an empty string, it returns nil (no conversion).
// Uses the htmlindex package to support IANA standard encoding names.
// Examples: shift_jis, euc-jp, iso-2022-jp, windows-1252, iso-8859-1, utf-8, etc.
func NewConverter(encodingName string) (Converter, error) {
	if encodingName == "" {
		return nil, nil
	}

	// Normalize: lowercase and replace underscores with hyphens
	normalizedName := strings.ToLower(strings.ReplaceAll(encodingName, "_", "-"))

	// UTF-8 requires no conversion
	if normalizedName == "utf-8" || normalizedName == "utf8" {
		return nil, nil
	}

	// Get Encoding from IANA standard encoding name via htmlindex.Get()
	enc, err := htmlindex.Get(normalizedName)
	if err != nil {
		return nil, fmt.Errorf("unsupported encoding: %s", encodingName)
	}

	return &textConverter{
		encodingName: encodingName,
		decoder:      enc.NewDecoder(),
	}, nil
}

// ConvertToUTF8 converts the input byte slice to UTF-8.
func (c *textConverter) ConvertToUTF8(content []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(content), c.decoder)
	utf8Content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("encoding conversion failed: %w", err)
	}
	return utf8Content, nil
}

// NormalizeExtension removes leading dot and converts to lowercase.
// Examples: ".TXT" -> "txt", "log" -> "log", ".Md" -> "md"
func NormalizeExtension(ext string) string {
	ext = strings.TrimPrefix(ext, ".")
	return strings.ToLower(ext)
}

// ParseEncodingMap parses "ext1:encoding1,ext2:encoding2" format.
// Returns a map of normalized extensions to converters.
// Validates all encoding names upfront (fail-fast).
func ParseEncodingMap(mapStr string) (map[string]Converter, error) {
	if mapStr == "" {
		return nil, nil
	}

	result := make(map[string]Converter)
	pairs := strings.Split(mapStr, ",")

	for _, pair := range pairs {
		parts := strings.Split(strings.TrimSpace(pair), ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid encoding map entry: %s (expected format: ext:encoding)", pair)
		}

		ext := NormalizeExtension(strings.TrimSpace(parts[0]))
		encodingName := strings.TrimSpace(parts[1])

		if ext == "" {
			return nil, fmt.Errorf("empty extension in encoding map entry: %s", pair)
		}
		if encodingName == "" {
			return nil, fmt.Errorf("empty encoding name in encoding map entry: %s", pair)
		}

		// Create converter (validates encoding name)
		converter, err := NewConverter(encodingName)
		if err != nil {
			return nil, fmt.Errorf("invalid encoding for extension '%s': %w", ext, err)
		}

		result[ext] = converter
	}

	return result, nil
}

// RemoveBOM removes UTF-8 BOM (0xEF 0xBB 0xBF) from the beginning of content.
// Returns the content without BOM and a boolean indicating if BOM was present.
func RemoveBOM(content []byte) ([]byte, bool) {
	if len(content) >= 3 && content[0] == 0xEF && content[1] == 0xBB && content[2] == 0xBF {
		return content[3:], true
	}
	return content, false
}

// NormalizeNewlines converts all line endings to LF (\n).
// Handles CRLF (\r\n), CR (\r), and LF (\n).
func NormalizeNewlines(content []byte) []byte {
	// First replace CRLF with LF (must be done before CR replacement)
	content = bytes.ReplaceAll(content, []byte("\r\n"), []byte("\n"))
	// Then replace remaining CR with LF
	content = bytes.ReplaceAll(content, []byte("\r"), []byte("\n"))
	return content
}
