package secret

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

var (
	// ErrRandomSourceFailure indicates that entropy retrieval failed.
	ErrRandomSourceFailure = errors.New("secret: random_source_failure")
	// ErrMissingRandomSource indicates that a nil entropy source was provided.
	ErrMissingRandomSource = errors.New("secret: missing_random_source")
)

// Generator wraps an entropy source for generating secrets.
type Generator struct {
	randomSource io.Reader
}

// NewGenerator constructs a Generator that draws entropy from the provided reader.
func NewGenerator(randomSource io.Reader) (*Generator, error) {
	if randomSource == nil {
		return nil, ErrMissingRandomSource
	}
	return &Generator{
		randomSource: randomSource,
	}, nil
}

// NewCryptoGenerator creates a Generator backed by crypto/rand.Reader.
func NewCryptoGenerator() (*Generator, error) {
	return NewGenerator(rand.Reader)
}

// GenerateSecret produces a URL-safe secret string using the configured entropy source.
func (generator *Generator) GenerateSecret(ctx context.Context, length ByteLength) (string, error) {
	if generator == nil {
		return "", ErrMissingRandomSource
	}
	return GenerateSecret(ctx, generator.randomSource, length)
}

// GenerateSecret creates a URL-safe secret string using the provided entropy source.
func GenerateSecret(ctx context.Context, randomSource io.Reader, length ByteLength) (string, error) {
	if ctx == nil {
		return "", fmt.Errorf("%w: nil context", ErrRandomSourceFailure)
	}
	if randomSource == nil {
		return "", ErrMissingRandomSource
	}
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("secret: context canceled: %w", err)
	}

	buffer := make([]byte, length.Value())
	if _, err := io.ReadFull(randomSource, buffer); err != nil {
		return "", fmt.Errorf("%w: %w", ErrRandomSourceFailure, err)
	}

	return base64.RawURLEncoding.EncodeToString(buffer), nil
}
