package headers

import (
	"bytes"
	"fmt"
	"slices"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	return Headers{}
}

// Parse parses the bytestream and adds the parsed headers to h.
func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	// Check if we have at least one complete line (ending with \r\n)
	crlfIndex := bytes.Index(data, []byte("\r\n"))
	if crlfIndex == -1 {
		// Not enough data for a full line yet
		return 0, false, nil
	}

	// Extract the current line (excluding \r\n)
	line := data[:crlfIndex]
	remaining := data[crlfIndex+2:] // Skip \r\n

	// Check for end of headers (empty line \r\n)
	if len(line) == 0 {
		return crlfIndex + 2, true, nil // Consume \r\n and signal completion
	}

	// Split into key:value
	colonIndex := bytes.IndexByte(line, ':')
	if colonIndex <= 0 {
		return 0, false, fmt.Errorf("malformed header: missing colon")
	}
	key := strings.ToLower(string(bytes.TrimSpace(line[:colonIndex])))
	value := string(bytes.TrimSpace(line[colonIndex+1:]))

	// Validate key
	allowed := []string{"!", "#", "$", "%", "&", "'", "*", "+", "-", ".", "^", "_", "`", "|", "~"}
	for _, char := range key {
		if (char < 'a' || char > 'z') && (char < 'A' || char > 'Z') {
			if !slices.Contains(allowed, string(char)) {
				return 0, false, fmt.Errorf("invalid character in header name: %q", char)
			}
		}
	}

	// Append to existing header or set new one
	if existing, ok := h[key]; ok {
		h[key] = existing + ", " + value // RFC 7230: Combine with comma
	} else {
		h[key] = value
	}

	// Check if the next part is the end of headers (\r\n)
	if len(remaining) >= 2 && bytes.Equal(remaining[:2], []byte("\r\n")) {
		// Consume the current line + \r\n + the next \r\n (total: crlfIndex + 4)
		return crlfIndex + 4, true, nil
	}

	// Return consumed bytes (line + \r\n)
	return crlfIndex + 2, false, nil
}

// Get gets the corresponding value of the key from h.
func (h Headers) Get(key string) string {
	return h[strings.ToLower(key)]
}

// Replace replaces the old value of the key with the new one
// and returns the new value. This is an no-op and returns "" if the key doesn't already exist.
func (h Headers) Replace(key string, value string) string {
	key = strings.ToLower(key)
	found := false
	for k, _ := range h {
		if strings.ToLower(k) == key {
			found = true
			h[k] = value
			return value
		}
	}

	if !found {
		return ""
	}

	return ""
}
