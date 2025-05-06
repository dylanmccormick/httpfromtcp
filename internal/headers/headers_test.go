package headers_test

import (
	"testing"

	"github.com/dylanmccormick/httpfromtcp/internal/headers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

func NewHeaders() headers.Headers {
	H := make(headers.Headers)
	return H
}

func TestHeaderLineParse(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Test: standardize caps
	headers = NewHeaders()
	data = []byte("hOsT: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, 23, n)
	assert.False(t, done)
	assert.Equal(t, "localhost:42069", headers["host"])

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\n Expert: thePrimeagen\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, 23, n)
	assert.False(t, done)
	data = data[n:]
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, 23, n)
	assert.False(t, done)
	assert.Equal(t, "thePrimeagen", headers["expert"])
	assert.Equal(t, "localhost:42069", headers["host"])

	// Test: bad char
	headers = NewHeaders()
	data = []byte("h;sT: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)


	headers = NewHeaders()
	data = []byte("Vibe: shift\r\n Vibe: unshift\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, 13, n)
	assert.False(t, done)
	data = data[n:]
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, 16, n)
	assert.False(t, done)
	assert.Equal(t, "shift, unshift", headers["vibe"])

	headers = NewHeaders()
	headers["vibe"] = "shift"  // Set before parsing
	data = []byte("Vibe: unshift\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, 15, n)
	assert.False(t, done)
	assert.Equal(t, "shift, unshift", headers["vibe"])
}
