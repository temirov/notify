package secret

import (
	"errors"
	"fmt"
)

var (
	// ErrInvalidByteLength indicates that a requested secret length is below the supported minimum.
	ErrInvalidByteLength = errors.New("secret: invalid_byte_length")
)

const (
	minSecretByteLength     = 32
	defaultSecretByteLength = 48
)

// ByteLength represents the size of a secret in bytes.
type ByteLength struct {
	value int
}

// NewByteLength constructs a ByteLength validated against the policy minimum.
func NewByteLength(value int) (ByteLength, error) {
	if value < minSecretByteLength {
		return ByteLength{}, fmt.Errorf("%w: %d is less than minimum %d", ErrInvalidByteLength, value, minSecretByteLength)
	}
	return ByteLength{value: value}, nil
}

// DefaultByteLength returns the recommended secret length.
func DefaultByteLength() ByteLength {
	return ByteLength{value: defaultSecretByteLength}
}

// Value exposes the underlying byte count.
func (length ByteLength) Value() int {
	return length.value
}
