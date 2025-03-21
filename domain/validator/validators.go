package validator

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

// RegisterCustomValidators registers custom validators with the provided validator
func RegisterCustomValidators(v *validator.Validate) {
	// Register custom validation functions
	v.RegisterValidation("alphanumdash", ValidateAlphaNumDash)
	v.RegisterValidation("secure_password", ValidateSecurePassword)
	v.RegisterValidation("no_html", ValidateNoHTML)
	v.RegisterValidation("no_sql", ValidateNoSQL)
	v.RegisterValidation("safe_text", ValidateSafeText)
	v.RegisterValidation("language_code", ValidateLanguageCode)
	v.RegisterValidation("entry_type", ValidateEntryType)
}

// ValidateAlphaNumDash validates that a string contains only alphanumeric characters and dashes
func ValidateAlphaNumDash(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true
	}

	// Update the regex to include underscores
	match, _ := regexp.MatchString("^[-a-zA-Z0-9_]+$", value)
	return match
}

// ValidateSecurePassword validates that a password meets security requirements
func ValidateSecurePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	if password == "" {
		return true
	}

	// Minimum length of 8 characters
	if len(password) < 8 {
		return false
	}

	// Check for at least one uppercase letter
	hasUpper := false
	// Check for at least one lowercase letter
	hasLower := false
	// Check for at least one digit
	hasDigit := false
	// Check for at least one special character
	hasSpecial := false

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	// Must satisfy at least 3 of the 4 requirements
	requirements := 0
	if hasUpper {
		requirements++
	}
	if hasLower {
		requirements++
	}
	if hasDigit {
		requirements++
	}
	if hasSpecial {
		requirements++
	}

	return requirements >= 3
}

// ValidateNoHTML validates that a string doesn't contain HTML tags
func ValidateNoHTML(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true
	}

	// Check for HTML tags
	match, _ := regexp.MatchString("<[^>]*>", value)
	return !match
}

// ValidateNoSQL validates that a string doesn't contain SQL injection attempts
func ValidateNoSQL(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true
	}

	// Check for common SQL injection patterns (case insensitive)
	sqlPatterns := []string{
		"--",
		";",
		"/*",
		"*/",
		"UNION",
		"SELECT",
		"DROP",
		"DELETE",
		"UPDATE",
		"INSERT",
		"EXEC",
	}

	valueLower := strings.ToLower(value)
	for _, pattern := range sqlPatterns {
		if strings.Contains(valueLower, strings.ToLower(pattern)) {
			return false
		}
	}

	return true
}

// ValidateSafeText validates that a string contains only safe text content
func ValidateSafeText(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true
	}

	// Check for HTML tags
	htmlMatch, _ := regexp.MatchString("<[^>]*>", value)
	if htmlMatch {
		return false
	}

	// Check for SQL injection attempts
	sqlMatch := ValidateNoSQL(fl)
	if !sqlMatch {
		return false
	}

	// Check for script tags and potentially dangerous JavaScript
	scriptPatterns := []string{
		"<script",
		"javascript:",
		"onclick=",
		"onerror=",
		"onload=",
		"eval(",
		"alert(",
	}

	valueLower := strings.ToLower(value)
	for _, pattern := range scriptPatterns {
		if strings.Contains(valueLower, pattern) {
			return false
		}
	}

	return true
}

// ValidateLanguageCode validates that a string is a valid ISO 639-1 language code
func ValidateLanguageCode(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true
	}

	// ISO 639-1 codes are 2 letters (occasionally with a dash and additional chars for variants)
	match, _ := regexp.MatchString("^[a-z]{2}(-[a-z0-9]+)?$", strings.ToLower(value))
	return match
}

// ValidateEntryType validates that a string is a valid entry type
func ValidateEntryType(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true
	}

	validTypes := map[string]bool{
		"WORD":          true,
		"COMPOUND_WORD": true,
		"PHRASE":        true,
	}

	return validTypes[value]
}
