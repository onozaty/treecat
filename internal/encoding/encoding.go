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
