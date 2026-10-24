//go:build offline

package crypto

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPkcs5RoundTrip(t *testing.T) {
	for _, plain := range [][]byte{
		[]byte("a"),
		[]byte("hello"),
		bytes.Repeat([]byte("x"), 16),
		bytes.Repeat([]byte("y"), 32),
	} {
		padded := pkcs5Padding(plain, 16)
		require.Equal(t, 0, len(padded)%16)
		got, err := pkcs5Unpadding(padded, 16)
		require.NoError(t, err)
		assert.Equal(t, plain, got)
	}
}

func TestPkcs5UnpaddingRejectsMalformed(t *testing.T) {
	tests := []struct {
		name string
		src  []byte
	}{
		{name: "empty", src: nil},
		{name: "not block multiple", src: []byte{1, 2, 3}},
		{name: "padding len zero", src: append(bytes.Repeat([]byte{1}, 15), 0)},
		{name: "padding len too large", src: append(bytes.Repeat([]byte{1}, 15), 17)},
		{name: "mismatched padding bytes", src: append(bytes.Repeat([]byte{1}, 13), 3, 3, 2)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := pkcs5Unpadding(tt.src, 16)
			assert.ErrorContains(t, err, "invalid padding")
		})
	}
}
