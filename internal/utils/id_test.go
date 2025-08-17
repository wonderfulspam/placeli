package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateID(t *testing.T) {
	id1 := GenerateID("test-place-id")
	id2 := GenerateID("test-place-id")
	id3 := GenerateID("different-place-id")

	assert.Len(t, id1, 12, "Generated ID should be 12 characters")
	assert.Equal(t, id1, id2, "Same input should generate same ID")
	assert.NotEqual(t, id1, id3, "Different inputs should generate different IDs")

	// Test that the ID is a valid hex string
	assert.Regexp(t, "^[a-f0-9]+$", id1, "ID should be hexadecimal")
}

func TestGenerateID_EdgeCases(t *testing.T) {
	// Test with empty string
	id := GenerateID("")
	assert.Len(t, id, 12)

	// Test with very long string
	longString := string(make([]byte, 1000))
	id = GenerateID(longString)
	assert.Len(t, id, 12)
}
