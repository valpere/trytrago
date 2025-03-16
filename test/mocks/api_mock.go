package mocks

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
)

// MockEntryHandler provides a mock implementation of EntryHandlerInterface
type MockEntryHandler struct {
	mock.Mock
}

func (m *MockEntryHandler) ListEntries(c *gin.Context) {
	m.Called(c)
}

func (m *MockEntryHandler) GetEntry(c *gin.Context) {
	m.Called(c)
}

func (m *MockEntryHandler) CreateEntry(c *gin.Context) {
	m.Called(c)
}

func (m *MockEntryHandler) UpdateEntry(c *gin.Context) {
	m.Called(c)
}

func (m *MockEntryHandler) DeleteEntry(c *gin.Context) {
	m.Called(c)
}

func (m *MockEntryHandler) GetMeaning(c *gin.Context) {
	m.Called(c)
}

func (m *MockEntryHandler) ListMeanings(c *gin.Context) {
	m.Called(c)
}

func (m *MockEntryHandler) AddMeaning(c *gin.Context) {
	m.Called(c)
}

func (m *MockEntryHandler) UpdateMeaning(c *gin.Context) {
	m.Called(c)
}

func (m *MockEntryHandler) DeleteMeaning(c *gin.Context) {
	m.Called(c)
}

func (m *MockEntryHandler) AddMeaningComment(c *gin.Context) {
	m.Called(c)
}

func (m *MockEntryHandler) ToggleMeaningLike(c *gin.Context) {
	m.Called(c)
}

// MockTranslationHandler provides a mock implementation of TranslationHandlerInterface
type MockTranslationHandler struct {
	mock.Mock
}

func (m *MockTranslationHandler) ListTranslations(c *gin.Context) {
	m.Called(c)
}

func (m *MockTranslationHandler) CreateTranslation(c *gin.Context) {
	m.Called(c)
}

func (m *MockTranslationHandler) UpdateTranslation(c *gin.Context) {
	m.Called(c)
}

func (m *MockTranslationHandler) DeleteTranslation(c *gin.Context) {
	m.Called(c)
}

func (m *MockTranslationHandler) AddTranslationComment(c *gin.Context) {
	m.Called(c)
}

func (m *MockTranslationHandler) ToggleTranslationLike(c *gin.Context) {
	m.Called(c)
}

// MockUserHandler provides a mock implementation of UserHandlerInterface
type MockUserHandler struct {
	mock.Mock
}

func (m *MockUserHandler) CreateUser(c *gin.Context) {
	m.Called(c)
}

func (m *MockUserHandler) GetCurrentUser(c *gin.Context) {
	m.Called(c)
}

func (m *MockUserHandler) UpdateCurrentUser(c *gin.Context) {
	m.Called(c)
}

func (m *MockUserHandler) DeleteCurrentUser(c *gin.Context) {
	m.Called(c)
}

func (m *MockUserHandler) Login(c *gin.Context) {
	m.Called(c)
}

func (m *MockUserHandler) RefreshToken(c *gin.Context) {
	m.Called(c)
}

func (m *MockUserHandler) ListUserEntries(c *gin.Context) {
	m.Called(c)
}

func (m *MockUserHandler) ListUserTranslations(c *gin.Context) {
	m.Called(c)
}

func (m *MockUserHandler) ListUserComments(c *gin.Context) {
	m.Called(c)
}

func (m *MockUserHandler) ListUserLikes(c *gin.Context) {
	m.Called(c)
}
