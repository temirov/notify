package secret

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"io"
	"testing"
)

type stubRandomReader struct {
	buffer []byte
	index  int
	err    error
}

func (reader *stubRandomReader) Read(p []byte) (int, error) {
	if reader.err != nil {
		return 0, reader.err
	}
	if reader.index >= len(reader.buffer) {
		return 0, io.EOF
	}
	n := copy(p, reader.buffer[reader.index:])
	reader.index += n
	return n, nil
}

func TestNewByteLengthValidation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		value       int
		expectError bool
	}{
		{
			name:        "rejects too small length",
			value:       16,
			expectError: true,
		},
		{
			name:  "accepts minimum length",
			value: 32,
		},
		{
			name:  "accepts larger length",
			value: 96,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			length, err := NewByteLength(testCase.value)
			if testCase.expectError {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}
				if !errors.Is(err, ErrInvalidByteLength) {
					t.Fatalf("expected ErrInvalidByteLength, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
			if length.Value() != testCase.value {
				t.Fatalf("expected value %d, got %d", testCase.value, length.Value())
			}
		})
	}
}

func TestGenerateSecretSuccess(t *testing.T) {
	t.Parallel()

	length, err := NewByteLength(48)
	if err != nil {
		t.Fatalf("expected nil error constructing length, got %v", err)
	}

	data := make([]byte, length.Value())
	for index := range data {
		data[index] = byte(index)
	}

	reader := &stubRandomReader{
		buffer: data,
	}
	ctx := context.Background()

	secretValue, generateErr := GenerateSecret(ctx, reader, length)
	if generateErr != nil {
		t.Fatalf("expected nil error, got %v", generateErr)
	}

	expected := base64.RawURLEncoding.EncodeToString(data)
	if secretValue != expected {
		t.Fatalf("expected %q, got %q", expected, secretValue)
	}
	if len(secretValue) == 0 {
		t.Fatalf("expected non-empty secret")
	}
}

func TestGenerateSecretRandomFailure(t *testing.T) {
	t.Parallel()

	length, err := NewByteLength(48)
	if err != nil {
		t.Fatalf("expected nil error constructing length, got %v", err)
	}
	expectedErr := errors.New("read failure")
	reader := &stubRandomReader{
		err: expectedErr,
	}

	_, generateErr := GenerateSecret(context.Background(), reader, length)
	if generateErr == nil {
		t.Fatalf("expected error but got nil")
	}
	if !errors.Is(generateErr, ErrRandomSourceFailure) {
		t.Fatalf("expected ErrRandomSourceFailure, got %v", generateErr)
	}
	if !errors.Is(generateErr, expectedErr) {
		t.Fatalf("expected wrapped read error, got %v", generateErr)
	}
}

func TestGenerateSecretHonorsContextCancellation(t *testing.T) {
	t.Parallel()

	length, err := NewByteLength(48)
	if err != nil {
		t.Fatalf("expected nil error constructing length, got %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, generateErr := GenerateSecret(ctx, bytes.NewReader(make([]byte, length.Value())), length)
	if generateErr == nil {
		t.Fatalf("expected error but got nil")
	}
	if !errors.Is(generateErr, context.Canceled) {
		t.Fatalf("expected context cancelled error, got %v", generateErr)
	}
}
