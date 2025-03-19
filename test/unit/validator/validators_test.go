package validator_test

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	domainValidator "github.com/valpere/trytrago/domain/validator"
)

// TestStruct is used to test custom validators
type TestStruct struct {
	Username     string `validate:"alphanumdash"`
	Password     string `validate:"secure_password"`
	Description  string `validate:"no_html"`
	SearchQuery  string `validate:"no_sql"`
	Content      string `validate:"safe_text"`
	LanguageCode string `validate:"language_code"`
	EntryType    string `validate:"entry_type"`
}

func TestValidateAlphaNumDash(t *testing.T) {
	validate := validator.New()
	domainValidator.RegisterCustomValidators(validate)

	tests := []struct {
		input string
		valid bool
	}{
		{"username123", true},
		{"user-name", true},
		{"user_name", true},
		{"user@name", false},
		{"user name", false},
		{"user<script>", false},
		{"", true}, // Empty strings are valid
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			testStruct := TestStruct{Username: tt.input}
			err := validate.Var(testStruct.Username, "alphanumdash")
			if tt.valid {
				assert.Nil(t, err)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}

func TestValidateSecurePassword(t *testing.T) {
	validate := validator.New()
	domainValidator.RegisterCustomValidators(validate)

	tests := []struct {
		input string
		valid bool
	}{
		{"Password123", true},  // Upper, lower, number
		{"Pass@word1", true},   // Upper, lower, number, special
		{"password", false},    // Too simple
		{"PASSWORD123", false}, // No lowercase
		{"pa$$w0rd", true},     // Lower, number, special
		{"Ab1@", false},        // Too short
		{"", true},             // Empty strings are valid
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			testStruct := TestStruct{Password: tt.input}
			err := validate.Var(testStruct.Password, "secure_password")
			if tt.valid {
				assert.Nil(t, err)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}

func TestValidateNoHTML(t *testing.T) {
	validate := validator.New()
	domainValidator.RegisterCustomValidators(validate)

	tests := []struct {
		input string
		valid bool
	}{
		{"Normal text", true},
		{"<script>alert('XSS')</script>", false},
		{"<div>Content</div>", false},
		{"Text with <b>bold</b>", false},
		{"", true}, // Empty strings are valid
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			testStruct := TestStruct{Description: tt.input}
			err := validate.Var(testStruct.Description, "no_html")
			if tt.valid {
				assert.Nil(t, err)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}

func TestValidateNoSQL(t *testing.T) {
	validate := validator.New()
	domainValidator.RegisterCustomValidators(validate)

	tests := []struct {
		input string
		valid bool
	}{
		{"Normal text", true},
		{"SELECT * FROM users", false},
		{"Drop Table", false},
		{"'; DROP TABLE users; --", false},
		{"", true}, // Empty strings are valid
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			testStruct := TestStruct{SearchQuery: tt.input}
			err := validate.Var(testStruct.SearchQuery, "no_sql")
			if tt.valid {
				assert.Nil(t, err)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}

func TestValidateSafeText(t *testing.T) {
	validate := validator.New()
	domainValidator.RegisterCustomValidators(validate)

	tests := []struct {
		input string
		valid bool
	}{
		{"Normal text", true},
		{"<script>alert('XSS')</script>", false},
		{"SELECT * FROM users", false},
		{"javascript:alert(1)", false},
		{"", true}, // Empty strings are valid
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			testStruct := TestStruct{Content: tt.input}
			err := validate.Var(testStruct.Content, "safe_text")
			if tt.valid {
				assert.Nil(t, err)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}

func TestValidateLanguageCode(t *testing.T) {
	validate := validator.New()
	domainValidator.RegisterCustomValidators(validate)

	tests := []struct {
		input string
		valid bool
	}{
		{"en", true},
		{"en-us", true},
		{"EN", true},   // Should be case-insensitive
		{"eng", false}, // Too long
		{"e", false},   // Too short
		{"e1", false},  // Invalid format
		{"", true},     // Empty strings are valid
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			testStruct := TestStruct{LanguageCode: tt.input}
			err := validate.Var(testStruct.LanguageCode, "language_code")
			if tt.valid {
				assert.Nil(t, err)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}

func TestValidateEntryType(t *testing.T) {
	validate := validator.New()
	domainValidator.RegisterCustomValidators(validate)

	tests := []struct {
		input string
		valid bool
	}{
		{"WORD", true},
		{"COMPOUND_WORD", true},
		{"PHRASE", true},
		{"word", false},     // Case sensitive
		{"SENTENCE", false}, // Not valid entry type
		{"", true},          // Empty strings are valid
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			testStruct := TestStruct{EntryType: tt.input}
			err := validate.Var(testStruct.EntryType, "entry_type")
			if tt.valid {
				assert.Nil(t, err)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}
