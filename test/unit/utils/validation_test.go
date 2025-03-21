package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valpere/trytrago/domain/utils"
)

func TestIsValidUUID(t *testing.T) {
	tests := []struct {
		name  string
		uuid  string
		valid bool
	}{
		{"Valid UUID", "123e4567-e89b-12d3-a456-426614174000", true},
		{"Uppercase UUID", "123E4567-E89B-12D3-A456-426614174000", true},
		{"Invalid format", "123e4567e89b12d3a456426614174000", false},
		{"Too short", "123e4567-e89b-12d3-a456", false},
		{"Too long", "123e4567-e89b-12d3-a456-4266141740001", false},
		{"Invalid characters", "123e4567-e89b-12d3-a456-42661417400g", false},
		{"Empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.IsValidUUID(tt.uuid)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		valid bool
	}{
		{"Valid email", "user@example.com", true},
		{"Valid with subdomain", "user@sub.example.com", true},
		{"Valid with plus", "user+tag@example.com", true},
		{"Valid with dots", "user.name@example.com", true},
		{"Missing @", "userexample.com", false},
		{"Missing domain", "user@", false},
		{"Missing username", "@example.com", false},
		{"Invalid TLD", "user@example", false},
		{"Empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.IsValidEmail(tt.email)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestIsValidEntryType(t *testing.T) {
	tests := []struct {
		name      string
		entryType string
		valid     bool
	}{
		{"Valid WORD", "WORD", true},
		{"Valid COMPOUND_WORD", "COMPOUND_WORD", true},
		{"Valid PHRASE", "PHRASE", true},
		{"Lowercase", "word", false},
		{"Invalid type", "SENTENCE", false},
		{"Empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.IsValidEntryType(tt.entryType)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestIsValidLanguageCode(t *testing.T) {
	tests := []struct {
		name  string
		code  string
		valid bool
	}{
		{"Valid code", "en", true},
		{"Valid with region", "en-us", true},
		{"Too long", "eng", false},
		{"Too short", "e", false},
		{"Invalid characters", "e$", false},
		{"Empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.IsValidLanguageCode(tt.code)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestIsValidPaginationParams(t *testing.T) {
	tests := []struct {
		name   string
		limit  int
		offset int
		valid  bool
	}{
		{"Valid params", 10, 0, true},
		{"Valid with offset", 20, 40, true},
		{"Max limit", 100, 0, true},
		{"Min limit", 1, 0, true},
		{"Limit too small", 0, 0, false},
		{"Limit too large", 101, 0, false},
		{"Negative offset", 10, -1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.IsValidPaginationParams(tt.limit, tt.offset)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestIsValidSortField(t *testing.T) {
	tests := []struct {
		name   string
		field  string
		entity string
		valid  bool
	}{
		{"Valid entry field", "word", "entry", true},
		{"Valid entry created_at", "created_at", "entry", true},
		{"Valid translation field", "language_id", "translation", true},
		{"Invalid entry field", "unknown", "entry", false},
		{"Invalid entity", "word", "unknown_entity", false},
		{"Empty field", "", "entry", true}, // Empty field should be valid (uses default)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.IsValidSortField(tt.field, tt.entity)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestIsValidPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		valid    bool
	}{
		{"Strong password", "StrongP@ss1", true},
		{"Mixed case and numbers", "Password123", true},
		{"With special chars", "Pass@word1", true},
		{"Too short", "Pass1", false},
		{"No uppercase", "password123", false},
		{"No numbers or special", "Password", false},
		{"Empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.IsValidPassword(tt.password)
			assert.Equal(t, tt.valid, result)
		})
	}
}
