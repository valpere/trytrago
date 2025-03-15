package mocks

import (
	"time"

	"github.com/google/uuid"
	"github.com/valpere/trytrago/infrastructure/auth"
)

// MockAuth sets up mock JWT auth functionality for testing
func SetupMockAuth() {
	// Initialize JWT with test key and duration
	auth.InitJWT("test-secret-key-for-unit-tests", 1*time.Hour)
}

// GenerateMockTokens generates mock tokens for testing
func GenerateMockTokens(userID uuid.UUID, username, role string) (string, string, error) {
	// Generate mock tokens
	accessToken, err := auth.GenerateToken(userID.String(), username, role)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := auth.GenerateRefreshToken(userID.String())
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}
