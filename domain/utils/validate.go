package utils

import (
	"regexp"
	"strings"
)

// Regular expressions for validation
var (
	uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-\.]{3,50}$`)
	languageCodeRegex = regexp.MustCompile(`^[a-z]{2}(-[a-z0-9]+)?$`)
)

// IsValidUUID validates a UUID string
func IsValidUUID(uuid string) bool {
	if uuid == "" {
		return false
	}

	// Convert to lowercase and check against regex
	return uuidRegex.MatchString(strings.ToLower(uuid))
}

// IsValidEmail validates an email address
func IsValidEmail(email string) bool {
	if email == "" {
		return false
	}

	// Convert to lowercase and check against regex
	return emailRegex.MatchString(strings.ToLower(email))
}

// IsValidUsername validates a username
func IsValidUsername(username string) bool {
	if username == "" {
		return false
	}

	// Check against regex
	return usernameRegex.MatchString(username)
}

// IsValidLanguageCode validates a language code
func IsValidLanguageCode(code string) bool {
	if code == "" {
		return false
	}

	// Convert to lowercase and check against regex
	return languageCodeRegex.MatchString(strings.ToLower(code))
}

// IsValidEntryType validates an entry type
func IsValidEntryType(entryType string) bool {
	if entryType == "" {
		return false
	}

	validTypes := map[string]bool{
		"WORD":          true,
		"COMPOUND_WORD": true,
		"PHRASE":        true,
	}

	return validTypes[entryType]
}

// IsValidPaginationParams validates pagination parameters
func IsValidPaginationParams(limit, offset int) bool {
	// Limit must be between 1 and 100
	if limit < 1 || limit > 100 {
		return false
	}

	// Offset must be non-negative
	if offset < 0 {
		return false
	}

	return true
}

// IsValidSortField validates a sort field for a specific entity
func IsValidSortField(field, entity string) bool {
	if field == "" {
		return true
	}

	// Define allowed sort fields per entity
	validFields := map[string]map[string]bool{
		"entry": {
			"word":       true,
			"created_at": true,
			"updated_at": true,
		},
		"meaning": {
			"created_at": true,
			"updated_at": true,
		},
		"translation": {
			"created_at":  true,
			"updated_at":  true,
			"language_id": true,
		},
		"comment": {
			"created_at":  true,
			"target_type": true,
		},
		"like": {
			"created_at":  true,
			"target_type": true,
		},
	}

	// Check if the entity exists
	if fields, ok := validFields[entity]; ok {
		// Check if the field is valid for this entity
		return fields[field]
	}

	return false
}

// IsValidSearchQuery validates a search query
func IsValidSearchQuery(query string) bool {
	// Query should not be too long
	if len(query) > 100 {
		return false
	}

	// Query should not contain SQL injection patterns
	if DetectSQLInjection(query) {
		return false
	}

	return true
}

// IsValidPassword validates a password
func IsValidPassword(password string) bool {
	// Password must be at least 8 characters
	if len(password) < 8 {
		return false
	}

	// Password must contain at least one uppercase letter
	hasUpper := false
	// Password must contain at least one lowercase letter
	hasLower := false
	// Password must contain at least one digit
	hasDigit := false
	// Password must contain at least one special character
	hasSpecial := false

	for _, char := range password {
		switch {
		case 'A' <= char && char <= 'Z':
			hasUpper = true
		case 'a' <= char && char <= 'z':
			hasLower = true
		case '0' <= char && char <= '9':
			hasDigit = true
		default:
			hasSpecial = true
		}
	}

	// Password must satisfy at least 3 of the 4 requirements
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
