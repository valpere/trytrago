package mocks

import (
    "context"

    "github.com/google/uuid"
    "github.com/stretchr/testify/mock"
    "gorm.io/gorm"

    "github.com/valpere/trytrago/domain/database"
    "github.com/valpere/trytrago/domain/database/repository"
    "github.com/valpere/trytrago/domain/model"
)

// MockRepository is a mock implementation of repository.Repository for testing
type MockRepository struct {
    mock.Mock
}

// Ensure MockRepository implements Repository interface
var _ repository.Repository = &MockRepository{}

// CreateEntry mocks the Repository.CreateEntry method
func (m *MockRepository) CreateEntry(ctx context.Context, entry *database.Entry) error {
    args := m.Called(ctx, entry)
    return args.Error(0)
}

// GetEntryByID mocks the Repository.GetEntryByID method
func (m *MockRepository) GetEntryByID(ctx context.Context, id uuid.UUID) (*database.Entry, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*database.Entry), args.Error(1)
}

// UpdateEntry mocks the Repository.UpdateEntry method
func (m *MockRepository) UpdateEntry(ctx context.Context, entry *database.Entry) error {
    args := m.Called(ctx, entry)
    return args.Error(0)
}

// DeleteEntry mocks the Repository.DeleteEntry method
func (m *MockRepository) DeleteEntry(ctx context.Context, id uuid.UUID) error {
    args := m.Called(ctx, id)
    return args.Error(0)
}

// ListEntries mocks the Repository.ListEntries method
func (m *MockRepository) ListEntries(ctx context.Context, params repository.ListParams) ([]database.Entry, error) {
    args := m.Called(ctx, params)
    if args.Get(0) == nil {
        return []database.Entry{}, args.Error(1)
    }
    return args.Get(0).([]database.Entry), args.Error(1)
}

// FindTranslations mocks the Repository.FindTranslations method
func (m *MockRepository) FindTranslations(ctx context.Context, word string, langID string) ([]database.Translation, error) {
    args := m.Called(ctx, word, langID)
    if args.Get(0) == nil {
        return []database.Translation{}, args.Error(1)
    }
    return args.Get(0).([]database.Translation), args.Error(1)
}

// RecordChange mocks the Repository.RecordChange method
func (m *MockRepository) RecordChange(ctx context.Context, change *database.ChangeHistory) error {
    args := m.Called(ctx, change)
    return args.Error(0)
}

// GetEntryHistory mocks the Repository.GetEntryHistory method
func (m *MockRepository) GetEntryHistory(ctx context.Context, entryID uuid.UUID) ([]database.ChangeHistory, error) {
    args := m.Called(ctx, entryID)
    if args.Get(0) == nil {
        return []database.ChangeHistory{}, args.Error(1)
    }
    return args.Get(0).([]database.ChangeHistory), args.Error(1)
}

// User operations
func (m *MockRepository) CreateUser(ctx context.Context, user *model.User) error {
    args := m.Called(ctx, user)
    return args.Error(0)
}

func (m *MockRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockRepository) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
    args := m.Called(ctx, username)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
    args := m.Called(ctx, email)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockRepository) UpdateUser(ctx context.Context, user *model.User) error {
    args := m.Called(ctx, user)
    return args.Error(0)
}

func (m *MockRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
    args := m.Called(ctx, id)
    return args.Error(0)
}

func (m *MockRepository) ListUserEntries(ctx context.Context, userID uuid.UUID, params repository.ListParams) ([]database.Entry, error) {
    args := m.Called(ctx, userID, params)
    if args.Get(0) == nil {
        return []database.Entry{}, args.Error(1)
    }
    return args.Get(0).([]database.Entry), args.Error(1)
}

// Social operations
func (m *MockRepository) CreateComment(ctx context.Context, comment *model.Comment) error {
    args := m.Called(ctx, comment)
    return args.Error(0)
}

func (m *MockRepository) GetCommentByID(ctx context.Context, id uuid.UUID) (*model.Comment, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*model.Comment), args.Error(1)
}

func (m *MockRepository) ListComments(ctx context.Context, targetType string, targetID uuid.UUID) ([]model.Comment, error) {
    args := m.Called(ctx, targetType, targetID)
    if args.Get(0) == nil {
        return []model.Comment{}, args.Error(1)
    }
    return args.Get(0).([]model.Comment), args.Error(1)
}

func (m *MockRepository) DeleteComment(ctx context.Context, id uuid.UUID) error {
    args := m.Called(ctx, id)
    return args.Error(0)
}

func (m *MockRepository) CreateLike(ctx context.Context, like *model.Like) error {
    args := m.Called(ctx, like)
    return args.Error(0)
}

func (m *MockRepository) DeleteLike(ctx context.Context, userID uuid.UUID, targetType string, targetID uuid.UUID) error {
    args := m.Called(ctx, userID, targetType, targetID)
    return args.Error(0)
}

func (m *MockRepository) GetLike(ctx context.Context, userID uuid.UUID, targetType string, targetID uuid.UUID) (*model.Like, error) {
    args := m.Called(ctx, userID, targetType, targetID)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*model.Like), args.Error(1)
}

func (m *MockRepository) CountLikes(ctx context.Context, targetType string, targetID uuid.UUID) (int64, error) {
    args := m.Called(ctx, targetType, targetID)
    return args.Get(0).(int64), args.Error(1)
}

// Maintenance operations
func (m *MockRepository) Ping(ctx context.Context) error {
    args := m.Called(ctx)
    return args.Error(0)
}

func (m *MockRepository) Close() error {
    args := m.Called()
    return args.Error(0)
}

func (m *MockRepository) GetDB() (*gorm.DB, error) {
    args := m.Called()
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*gorm.DB), args.Error(1)
}

func (m *MockRepository) WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
    args := m.Called(ctx, fn)
    return args.Error(0)
}
