package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	// ErrInvalidToken is returned when a token is invalid
	ErrInvalidToken = errors.New("invalid token")
	
	// ErrExpiredToken is returned when a token is expired
	ErrExpiredToken = errors.New("token has expired")
	
	// jwtSecret is the secret key used to sign JWT tokens
	jwtSecret []byte
	
	// tokenExpiry is the token expiration time
	tokenExpiry time.Duration
)

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

// GenerateToken creates a new JWT token for a user
func GenerateToken(userID, username, role string) (string, error) {
	// Check if JWT has been initialized
	if len(jwtSecret) == 0 {
		return "", errors.New("JWT not initialized")
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

// GenerateRefreshToken creates a new refresh token
func GenerateRefreshToken(userID string) (string, error) {
	// Create token claims with longer expiry
	now := time.Now()
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(tokenExpiry * 24 * 7)), // 1 week
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

// ValidateToken validates a JWT token and returns the parsed token
func ValidateToken(tokenString string) (*jwt.Token, error) {
	// Check if JWT has been initialized
	if len(jwtSecret) == 0 {
		return nil, errors.New("JWT not initialized")
	}

	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	// Handle errors
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	// Check if the token is valid
	if !token.Valid {
		return nil, ErrInvalidToken
	}

	return token, nil
}

// ExtractUserIDFromToken extracts the user ID from a JWT token
func ExtractUserIDFromToken(tokenString string) (string, error) {
	token, err := ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return "", ErrInvalidToken
	}

	return claims.UserID, nil
}

// ValidateRefreshToken validates a refresh token and returns the user ID
func ValidateRefreshToken(tokenString string) (string, error) {
	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	// Handle errors
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", ErrExpiredToken
		}
		return "", ErrInvalidToken
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
