package model

import (
	"time"

	"github.com/google/uuid"
)

// User represents a dictionary user
type User struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key"`
	Username  string    `gorm:"type:varchar(50);uniqueIndex;not null"`
	Email     string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	Password  string    `gorm:"type:varchar(255);not null"` // Stored as bcrypt hash
	Avatar    string    `gorm:"type:varchar(255)"`
	Role      UserRole  `gorm:"type:varchar(20);not null;default:'USER'"`
	IsActive  bool      `gorm:"not null;default:true"`
	CreatedAt time.Time
	UpdatedAt time.Time
	LastLogin *time.Time

	// Relations
	Comments []Comment `gorm:"foreignKey:UserID"`
	Likes    []Like    `gorm:"foreignKey:UserID"`
}

// UserRole represents user permission levels
type UserRole string

// Available user roles
const (
	RoleUser  UserRole = "USER"
	RoleAdmin UserRole = "ADMIN"
)

// AuthToken represents an authentication token for a user
type AuthToken struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key"`
	UserID       uuid.UUID `gorm:"type:uuid;index;not null"`
	AccessToken  string    `gorm:"type:varchar(255);not null"`
	RefreshToken string    `gorm:"type:varchar(255);not null"`
	ExpiresAt    time.Time `gorm:"not null"`
	CreatedAt    time.Time
	RevokedAt    *time.Time
	UserAgent    string    `gorm:"type:varchar(255)"`
	ClientIP     string    `gorm:"type:varchar(45)"`
}

// UserPreference represents user settings and preferences
type UserPreference struct {
	ID              uuid.UUID `gorm:"type:uuid;primary_key"`
	UserID          uuid.UUID `gorm:"type:uuid;uniqueIndex;not null"`
	DefaultLanguage string    `gorm:"type:varchar(5);not null;default:'en'"` // ISO 639-1 language code
	ThemePreference string    `gorm:"type:varchar(20);not null;default:'system'"`
	EmailNotify     bool      `gorm:"not null;default:true"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// UserStats tracks usage statistics for a user
type UserStats struct {
	ID                 uuid.UUID `gorm:"type:uuid;primary_key"`
	UserID             uuid.UUID `gorm:"type:uuid;uniqueIndex;not null"`
	EntriesCreated     int       `gorm:"not null;default:0"`
	EntriesUpdated     int       `gorm:"not null;default:0"`
	MeaningsAdded      int       `gorm:"not null;default:0"`
	TranslationsAdded  int       `gorm:"not null;default:0"`
	CommentsPosted     int       `gorm:"not null;default:0"`
	LikesGiven         int       `gorm:"not null;default:0"`
	ReputationPoints   int       `gorm:"not null;default:0"`
	LastActivityAt     time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
