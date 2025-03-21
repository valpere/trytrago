package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valpere/trytrago/domain/utils"
)

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Normal text", "Hello World", "Hello World"},
		{"With HTML", "<script>alert('XSS')</script>", "alert(&#39;XSS&#39;)"},
		{"Mixed content", "Text with <b>bold</b> tags", "Text with bold tags"},
		{"Extra whitespace", "  Too   many    spaces  ", "Too many spaces"},
		{"Empty string", "", ""},
		{"Only HTML", "<div></div>", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sanitized := utils.SanitizeString(tt.input)
			assert.Equal(t, tt.expected, sanitized)
		})
	}
}

func TestSanitizeEmail(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Normal email", "user@example.com", "user@example.com"},
		{"Uppercase", "USER@EXAMPLE.COM", "user@example.com"},
		{"Extra whitespace", "  user@example.com  ", "user@example.com"},
		{"Multiple @ signs", "user@host@example.com", "user@example.com"},
		{"Empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sanitized := utils.SanitizeEmail(tt.input)
			assert.Equal(t, tt.expected, sanitized)
		})
	}
}

func TestSanitizeUsername(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Normal username", "john_doe", "john_doe"},
		{"With special chars", "john@doe!", "johndoe"},
		{"With spaces", "john doe", "johndoe"},
		{"With hyphen and dot", "john-doe.123", "john-doe.123"},
		{"Empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sanitized := utils.SanitizeUsername(tt.input)
			assert.Equal(t, tt.expected, sanitized)
		})
	}
}

func TestDetectSQLInjection(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		isAttack bool
	}{
		{"Normal text", "Hello World", false},
		{"SQL comment", "Hello -- World", true},
		{"SQL keyword", "DROP TABLE users", true},
		{"SQL UNION", "UNION SELECT * FROM users", true},
		{"SQL with semicolon", "value'; DROP TABLE users;", true},
		{"Mixed case SQL", "UniOn SeLeCt", true},
		{"Empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isAttack := utils.DetectSQLInjection(tt.input)
			assert.Equal(t, tt.isAttack, isAttack)
		})
	}
}

func TestSanitizeSearchQuery(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Normal query", "search term", "search term"},
		{"With wildcards", "search%term", "search\\%term"},
		{"SQL injection", "term' OR 1=1", "term OR 1=1"},
		{"Multiple wildcards", "term_%[]", "term\\_\\%\\[\\]"},
		{"Empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sanitized := utils.SanitizeSearchQuery(tt.input)
			assert.Equal(t, tt.expected, sanitized)
		})
	}
}

func TestSanitizeJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Valid JSON", `{"key":"value"}`, `{"key":"value"}`},
		{"With HTML tags", `{"key":"<script>alert('XSS')</script>"}`, `{"key":"\u003Cscript\u003Ealert('XSS')\u003C/script\u003E"}`},
		{"With special chars", `{"key":"<>&"}`, `{"key":"\u003C\u003E\u0026"}`},
		{"Empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sanitized := utils.SanitizeJSON(tt.input)
			assert.Equal(t, tt.expected, sanitized)
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Valid filename", "document.txt", "document.txt"},
		{"With path", "/var/www/document.txt", "varwwwdocument.txt"},
		{"With path traversal", "../../../etc/passwd", "etcpasswd"},
		{"With invalid chars", "document?.txt", "document.txt"},
		{"Empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sanitized := utils.SanitizeFilename(tt.input)
			assert.Equal(t, tt.expected, sanitized)
		})
	}
}

func TestSanitizeMap(t *testing.T) {
	input := map[string]interface{}{
		"name":  "John <script>alert('XSS')</script>",
		"email": "john@example.com",
		"preferences": map[string]interface{}{
			"theme": "<div>Dark</div>",
		},
		"tags": []interface{}{"tag1", "<b>tag2</b>"},
	}

	expected := map[string]interface{}{
		"name":  "John alert(&#39;XSS&#39;)",
		"email": "john@example.com",
		"preferences": map[string]interface{}{
			"theme": "Dark",
		},
		"tags": []interface{}{"tag1", "tag2"},
	}

	t.Run("Map sanitization", func(t *testing.T) {
		sanitized := utils.SanitizeMap(input)
		assert.Equal(t, expected, sanitized)
	})
}

func TestSanitizeLanguageCode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Valid code", "en", "en"},
		{"Uppercase", "EN", "en"},
		{"With region", "en-US", "en-us"},
		{"With invalid chars", "en@US", "enus"},
		{"Empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sanitized := utils.SanitizeLanguageCode(tt.input)
			assert.Equal(t, tt.expected, sanitized)
		})
	}
}

func TestSanitizeUUID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Valid UUID", "123e4567-e89b-12d3-a456-426614174000", "123e4567-e89b-12d3-a456-426614174000"},
		{"Uppercase UUID", "123E4567-E89B-12D3-A456-426614174000", "123E4567-E89B-12D3-A456-426614174000"},
		{"With whitespace", "  123e4567-e89b-12d3-a456-426614174000  ", "123e4567-e89b-12d3-a456-426614174000"},
		{"Invalid format", "123e4567e89b12d3a456426614174000", ""},
		{"Invalid characters", "123e4567-e89b-12d3-a456-42661417400g", ""},
		{"Empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sanitized := utils.SanitizeUUID(tt.input)
			assert.Equal(t, tt.expected, sanitized)
		})
	}
}
