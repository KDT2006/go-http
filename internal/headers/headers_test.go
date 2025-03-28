package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeaders(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["Host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Test: Valid single header with extra whitespace
	headers = NewHeaders()
	data = []byte("Content-Type:application/json \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.Equal(t, "application/json", headers["Content-Type"])
	require.Equal(t, len(data)-2, n)
	assert.False(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}

func TestMultipleHeaders(t *testing.T) {
	// Test: Valid 2 headers with existing headers
	headers := NewHeaders()
	headers["user-agent"] = "edge"
	data := []byte("Content-Type:application/json\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.Equal(t, "application/json", headers.Get("Content-Type"))
	require.Equal(t, len(data), n)
	assert.False(t, done)

	data = []byte("Host:localhost:8080\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.Equal(t, "localhost:8080", headers.Get("Host"))
	require.Equal(t, len(data), n)
	assert.False(t, done)

	data = []byte("\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.Equal(t, 2, n)
	assert.True(t, done)

	// Check all existing key-value pairs
	require.Equal(t, "edge", headers.Get("User-Agent"))
	require.Equal(t, "application/json", headers.Get("Content-Type"))
	require.Equal(t, "localhost:8080", headers.Get("Host"))
}

func TestInconsistentCasing(t *testing.T) {
	headers := NewHeaders()
	data := []byte("CONTent-TYpe: application/json\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.Nil(t, err)
	require.Equal(t, len(data)-2, n)
	require.Equal(t, false, done)
	require.Equal(t, "application/json", headers["content-type"])
}

func TestInvalidChar(t *testing.T) {
	headers := NewHeaders()
	data := []byte("H©st: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.Error(t, err)
	require.Equal(t, 0, n)
	require.False(t, done)
	require.Equal(t, 0, len(headers))
}

func TestMultipleFieldValues(t *testing.T) {
	headers := NewHeaders()
	headers["lang"] = "Go"
	data := []byte("lang:Ocaml\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.Nil(t, err)
	require.Equal(t, len(data), n)
	require.Equal(t, "Go, Ocaml", headers["lang"])
	require.Equal(t, 1, len(headers))
	require.True(t, done)
}
