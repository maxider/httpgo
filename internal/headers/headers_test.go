package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseHeader(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("HoSt: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 25, n)
	assert.True(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       HOST : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Valid single header with extra whitespace
	headers = NewHeaders()
	data = []byte("HoST: localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, 32, n)
	assert.True(t, done)

	// Test: Valid 2 headers with existing headers
	headers = NewHeaders()
	data = []byte("HoST: localhost:42069\r\n foObar: barbaz\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, "barbaz", headers["foobar"])
	assert.Equal(t, 42, n)
	assert.True(t, done)

	// Test: Valid 2 headers with existing headers, not done
	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\n foobar: barbaz\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, "barbaz", headers["foobar"])
	assert.Equal(t, 40, n)
	assert.False(t, done)

	// Test: Invalid character in header
	headers = NewHeaders()
	data = []byte("H@st: localhost:42069\r\n foobar: barbaz\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Multiple field-names with same name but different values
	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\n HoST: barbaz\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069, barbaz", headers["host"])
	assert.Equal(t, 38, n)
	assert.False(t, done)

}
