package utils

import (
	"html"
	"regexp"
	"strings"
)

var (
	// Regular expressions for sanitization
	htmlTagsRegex     = regexp.MustCompile("<[^>]*>")
	multiSpaceRegex   = regexp.MustCompile(`\s+`)
	sqlInjectionRegex = regexp.MustCompile(`(?i)(?:'|;|--|/\*|\*/|\b(?:ALTER|CREATE|DELETE|DROP|EXEC(?:UTE)?|INSERT(?: +INTO)?|MERGE|SELECT|UPDATE|UNION(?: +ALL)?)\b)`)
)

// SanitizeString sanitizes a string by removing HTML tags and potentially dangerous content
func SanitizeString(input string) string {
	if input == "" {
		return input
	}

	// First remove any HTML tags
	sanitized := htmlTagsRegex.ReplaceAllString(input, "")

	// Then escape HTML entities in the remaining text
	sanitized = html.EscapeString(sanitized)

	// Normalize whitespace
	sanitized = multiSpaceRegex.ReplaceAllString(sanitized, " ")

	// Trim whitespace
	sanitized = strings.TrimSpace(sanitized)

	return sanitized
}

// SanitizeHTML sanitizes HTML content, allowing certain safe tags and attributes
func SanitizeHTML(input string) string {
	if input == "" {
		return input
	}

	// This is a simplified version - in a real implementation, you would use
	// a proper HTML sanitizer library like bluemonday

	// For now, we'll just escape all HTML
	sanitized := html.EscapeString(input)
	return sanitized
}

// SanitizeEmail sanitizes an email address
func SanitizeEmail(email string) string {
	if email == "" {
		return email
	}

	// Trim whitespace
	email = strings.TrimSpace(email)

	// Convert to lowercase
	email = strings.ToLower(email)

	// Replace multiple @ signs with a single one
	parts := strings.Split(email, "@")
	if len(parts) > 2 {
		email = parts[0] + "@" + parts[len(parts)-1]
	}

	return email
}

// SanitizeUsername sanitizes a username
func SanitizeUsername(username string) string {
	if username == "" {
		return username
	}

	// Trim whitespace
	username = strings.TrimSpace(username)

	// Remove any potentially dangerous characters
	// Allow only alphanumeric, underscore, dash, and dot
	re := regexp.MustCompile(`[^-a-zA-Z0-9_.]`)
	username = re.ReplaceAllString(username, "")

	return username
}

// DetectSQLInjection checks if a string contains SQL injection patterns
func DetectSQLInjection(input string) bool {
	return sqlInjectionRegex.MatchString(input)
}

// SanitizeSearchQuery sanitizes a search query
func SanitizeSearchQuery(query string) string {
	if query == "" {
		return query
	}

	// Trim whitespace
	query = strings.TrimSpace(query)

	// Escape wildcards and other special characters used in SQL LIKE
	escapedChars := []string{"%", "_", "[", "]", "^"}
	for _, char := range escapedChars {
		query = strings.ReplaceAll(query, char, "\\"+char)
	}

	// Remove any SQL injection patterns
	if DetectSQLInjection(query) {
		// If SQL injection is detected, sanitize more aggressively
		query = sqlInjectionRegex.ReplaceAllString(query, "")
	}

	return query
}

// SanitizeJSON sanitizes a JSON string
func SanitizeJSON(jsonStr string) string {
	if jsonStr == "" {
		return jsonStr
	}

	// Replace potentially dangerous characters in JSON
	sanitized := strings.ReplaceAll(jsonStr, "<", "\\u003C")
	sanitized = strings.ReplaceAll(sanitized, ">", "\\u003E")
	sanitized = strings.ReplaceAll(sanitized, "&", "\\u0026")

	return sanitized
}

// SanitizeFilename sanitizes a filename
func SanitizeFilename(filename string) string {
	if filename == "" {
		return filename
	}

	// Remove any path traversal sequences
	filename = strings.ReplaceAll(filename, "..", "")
	filename = strings.ReplaceAll(filename, "/", "")
	filename = strings.ReplaceAll(filename, "\\", "")

	// Remove any potentially dangerous characters
	re := regexp.MustCompile(`[^a-zA-Z0-9_\\-\\.]`)
	filename = re.ReplaceAllString(filename, "")

	return filename
}

// SanitizeMap sanitizes all string values in a map
func SanitizeMap(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range data {
		// Sanitize the key
		sanitizedKey := SanitizeString(key)

		// Process the value based on its type
		switch v := value.(type) {
		case string:
			result[sanitizedKey] = SanitizeString(v)
		case map[string]interface{}:
			result[sanitizedKey] = SanitizeMap(v)
		case []interface{}:
			result[sanitizedKey] = SanitizeArray(v)
		default:
			// For non-string types, keep as is
			result[sanitizedKey] = v
		}
	}

	return result
}

// SanitizeArray sanitizes all string values in an array
func SanitizeArray(data []interface{}) []interface{} {
	result := make([]interface{}, len(data))

	for i, value := range data {
		switch v := value.(type) {
		case string:
			result[i] = SanitizeString(v)
		case map[string]interface{}:
			result[i] = SanitizeMap(v)
		case []interface{}:
			result[i] = SanitizeArray(v)
		default:
			// For non-string types, keep as is
			result[i] = v
		}
	}

	return result
}

// SanitizeLanguageCode sanitizes a language code
func SanitizeLanguageCode(code string) string {
	if code == "" {
		return code
	}

	// Trim whitespace
	code = strings.TrimSpace(code)

	// Convert to lowercase
	code = strings.ToLower(code)

	// Remove any non-alphanumeric characters
	re := regexp.MustCompile(`[^-a-z0-9]`)
	code = re.ReplaceAllString(code, "")

	return code
}

// EscapeJavaScript sanitizes JavaScript content
func EscapeJavaScript(js string) string {
	if js == "" {
		return js
	}

	// Replace potentially dangerous characters
	sanitized := strings.ReplaceAll(js, "<", "\\u003C")
	sanitized = strings.ReplaceAll(sanitized, ">", "\\u003E")
	sanitized = strings.ReplaceAll(sanitized, "'", "\\'")
	sanitized = strings.ReplaceAll(sanitized, "\"", "\\\"")
	sanitized = strings.ReplaceAll(sanitized, "\n", "\\n")
	sanitized = strings.ReplaceAll(sanitized, "\r", "\\r")
	sanitized = strings.ReplaceAll(sanitized, "&", "\\u0026")

	return sanitized
}

// SanitizeUUID validates and sanitizes a UUID string
func SanitizeUUID(uuid string) string {
	if uuid == "" {
		return uuid
	}

	// Trim whitespace
	uuid = strings.TrimSpace(uuid)

	// Check if it matches UUID format
	re := regexp.MustCompile("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
	if !re.MatchString(strings.ToLower(uuid)) {
		// If not a valid UUID, return empty string
		return ""
	}

	return uuid
}
