package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	// ErrInvalidToken is returned when a token is invalid
	ErrInvalidToken = errors.New("invalid token")

	// ErrExpiredToken is returned when a token is expired
	ErrExpiredToken = errors.New("token has expired")

	// ErrInvalidSigningMethod is returned when a token uses an invalid signing method
	ErrInvalidSigningMethod = errors.New("invalid signing method")

	// ErrTokenNotInitialized is returned when JWT hasn't been initialized
	ErrTokenNotInitialized = errors.New("JWT not initialized")

	// jwtSecret is the secret key used to sign JWT tokens
	jwtSecret []byte

	// tokenExpiry is the token expiration time
	tokenExpiry time.Duration
)

// TokenClaims represents the extracted data from a validated token
type TokenClaims struct {
	UserID    string
	Username  string
	Role      string
	IssuedAt  time.Time
	ExpiresAt time.Time
}

// CustomClaims defines the claims for our JWT tokens
type CustomClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// InitJWT initializes the JWT package with configuration
func InitJWT(secret string, expiry time.Duration) {
	jwtSecret = []byte(secret)
	tokenExpiry = expiry
}

// IsInitialized returns whether JWT has been initialized
func IsInitialized() bool {
	return len(jwtSecret) > 0
}

// GenerateToken creates a new JWT token for a user
func GenerateToken(userID, username, role string) (string, error) {
	// Check if JWT has been initialized
	if !IsInitialized() {
		return "", ErrTokenNotInitialized
	}

	// Create token claims
	now := time.Now()
	claims := CustomClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(tokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "trytrago",
			Subject:   userID,
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// GenerateTokenFromUUID is a convenience function that accepts a UUID
func GenerateTokenFromUUID(userID uuid.UUID, username, role string) (string, error) {
	return GenerateToken(userID.String(), username, role)
}

// GenerateRefreshToken creates a new refresh token
func GenerateRefreshToken(userID string) (string, error) {
	// Check if JWT has been initialized
	if !IsInitialized() {
		return "", ErrTokenNotInitialized
	}

	// Create token claims with longer expiry
	now := time.Now()
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(tokenExpiry * 24 * 7)), // 7 days
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		Issuer:    "trytrago-refresh",
		Subject:   userID,
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return tokenString, nil
}

// GenerateRefreshTokenFromUUID is a convenience function that accepts a UUID
func GenerateRefreshTokenFromUUID(userID uuid.UUID) (string, error) {
	return GenerateRefreshToken(userID.String())
}

// ValidateToken validates a JWT token and returns the extracted claims
func ValidateToken(tokenString string) (*TokenClaims, error) {
	// Check if JWT has been initialized
	if !IsInitialized() {
		return nil, ErrTokenNotInitialized
	}

	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidSigningMethod
		}
		return jwtSecret, nil
	})

	// Handle parsing errors
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	// Check if the token is valid
	if !token.Valid {
		return nil, ErrInvalidToken
	}

	// Extract and validate the claims
	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	// Create TokenClaims structure
	return &TokenClaims{
		UserID:    claims.UserID,
		Username:  claims.Username,
		Role:      claims.Role,
		IssuedAt:  claims.IssuedAt.Time,
		ExpiresAt: claims.ExpiresAt.Time,
	}, nil
}

// ExtractUserIDFromToken extracts the user ID from a JWT token
func ExtractUserIDFromToken(tokenString string) (string, error) {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	return claims.UserID, nil
}

// ExtractUserUUIDFromToken extracts the user ID as UUID from a JWT token
func ExtractUserUUIDFromToken(tokenString string) (uuid.UUID, error) {
	userID, err := ExtractUserIDFromToken(tokenString)
	if err != nil {
		return uuid.Nil, err
	}

	// Convert the ID to UUID
	id, err := uuid.Parse(userID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: invalid UUID format", ErrInvalidToken)
	}

	return id, nil
}

// ValidateRefreshToken validates a refresh token and returns the user ID
func ValidateRefreshToken(tokenString string) (string, error) {
	// Check if JWT has been initialized
	if !IsInitialized() {
		return "", ErrTokenNotInitialized
	}

	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidSigningMethod
		}
		return jwtSecret, nil
	})

	// Handle errors
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", ErrExpiredToken
		}
		return "", fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	// Check if the token is valid
	if !token.Valid {
		return "", ErrInvalidToken
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", ErrInvalidToken
	}

	// Verify issuer
	if issuer, ok := claims["iss"].(string); !ok || issuer != "trytrago-refresh" {
		return "", ErrInvalidToken
	}

	// Extract user ID
	userID, ok := claims["sub"].(string)
	if !ok {
		return "", ErrInvalidToken
	}

	return userID, nil
}

// IsTokenExpired checks if a token is expired without validating the signature
func IsTokenExpired(tokenString string) bool {
	// Parse the token without validating signature
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &CustomClaims{})
	if err != nil {
		return true
	}

	// Check expiration time
	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return true
	}

	// Check if token is expired
	return claims.ExpiresAt.Time.Before(time.Now())
}
