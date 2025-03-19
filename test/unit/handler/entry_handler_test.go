package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/valpere/trytrago/application/dto/request"
	"github.com/valpere/trytrago/application/dto/response"
	"github.com/valpere/trytrago/domain/logging"
	"github.com/valpere/trytrago/interface/api/rest/handler"
	"go.uber.org/zap/zapcore"
)

// MockEntryService mocks the EntryService interface
type MockEntryService struct {
	mock.Mock
}

// Implement EntryService methods
func (m *MockEntryService) CreateEntry(ctx context.Context, req *request.CreateEntryRequest) (*response.EntryResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.EntryResponse), args.Error(1)
}

func (m *MockEntryService) GetEntryByID(ctx context.Context, id uuid.UUID) (*response.EntryResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.EntryResponse), args.Error(1)
}

func (m *MockEntryService) UpdateEntry(ctx context.Context, id uuid.UUID, req *request.UpdateEntryRequest) (*response.EntryResponse, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.EntryResponse), args.Error(1)
}

func (m *MockEntryService) DeleteEntry(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockEntryService) ListEntries(ctx context.Context, req *request.ListEntriesRequest) (*response.EntryListResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.EntryListResponse), args.Error(1)
}

func (m *MockEntryService) AddMeaning(ctx context.Context, entryID uuid.UUID, req *request.CreateMeaningRequest) (*response.MeaningResponse, error) {
	args := m.Called(ctx, entryID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.MeaningResponse), args.Error(1)
}

func (m *MockEntryService) UpdateMeaning(ctx context.Context, id uuid.UUID, req *request.UpdateMeaningRequest) (*response.MeaningResponse, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.MeaningResponse), args.Error(1)
}

func (m *MockEntryService) DeleteMeaning(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockEntryService) ListMeanings(ctx context.Context, entryID uuid.UUID) (*response.MeaningListResponse, error) {
	args := m.Called(ctx, entryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.MeaningListResponse), args.Error(1)
}

func (m *MockEntryService) AddMeaningComment(ctx context.Context, meaningID uuid.UUID, req *request.CreateCommentRequest) (*response.CommentResponse, error) {
	args := m.Called(ctx, meaningID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.CommentResponse), args.Error(1)
}

func (m *MockEntryService) ToggleMeaningLike(ctx context.Context, meaningID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, meaningID, userID)
	return args.Error(0)
}

// MockLogger is a mock implementation of logging.Logger
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(msg string, fields ...zapcore.Field)   {}
func (m *MockLogger) Info(msg string, fields ...zapcore.Field)    {}
func (m *MockLogger) Warn(msg string, fields ...zapcore.Field)    {}
func (m *MockLogger) Error(msg string, fields ...zapcore.Field)   {}
func (m *MockLogger) With(fields ...zapcore.Field) logging.Logger { return m }
func (m *MockLogger) Sync() error                                 { return nil }

// Setup function for tests
func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

// Basic test for ListEntries
func TestListEntries(t *testing.T) {
	// Setup
	mockService := new(MockEntryService)
	mockLogger := new(MockLogger)
	handler := handler.NewEntryHandler(mockService, mockLogger)

	// Setup router
	router := setupRouter()
	router.GET("/entries", handler.ListEntries)

	// Create request with query parameters
	req, _ := http.NewRequest("GET", "/entries?word_filter=test&type=WORD", nil)
	w := httptest.NewRecorder()

	// Setup mock expectations with ANY request parameter
	// This is important because we don't know exactly how the handler will construct the request object
	mockService.On("ListEntries", mock.Anything, mock.Anything).Return(
		&response.EntryListResponse{
			Entries: []*response.EntryResponse{},
			Total:   0,
			Limit:   20,
			Offset:  0,
		}, nil)

	// Perform request
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

// Test GetEntry with valid and invalid input
func TestGetEntry(t *testing.T) {
	// Setup
	mockService := new(MockEntryService)
	mockLogger := new(MockLogger)
	h := handler.NewEntryHandler(mockService, mockLogger)
	router := setupRouter()
	router.GET("/entries/:id", h.GetEntry)

	// Valid UUID
	validID := uuid.New()

	// Setup mock expectations for valid case
	mockService.On("GetEntryByID", mock.Anything, validID).Return(
		&response.EntryResponse{
			ID:   validID,
			Word: "example",
			Type: "WORD",
		}, nil)

	// Test valid UUID
	req, _ := http.NewRequest("GET", "/entries/"+validID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert valid case
	assert.Equal(t, http.StatusOK, w.Code)

	// Test invalid UUID
	req, _ = http.NewRequest("GET", "/entries/not-a-uuid", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert invalid case
	assert.Equal(t, http.StatusBadRequest, w.Code)

	mockService.AssertExpectations(t)
}

// Test CreateEntry
func TestCreateEntry(t *testing.T) {
	// Setup
	mockService := new(MockEntryService)
	mockLogger := new(MockLogger)
	h := handler.NewEntryHandler(mockService, mockLogger)
	router := setupRouter()
	router.POST("/entries", h.CreateEntry)

	// Create mock request body
	entryID := uuid.New()
	requestBody := map[string]interface{}{
		"word":          "example",
		"type":          "WORD",
		"pronunciation": "eg-zam-pul",
	}

	// Setup mock expectations - match ANY context and similar request
	mockService.On("CreateEntry", mock.Anything, mock.Anything).Return(
		&response.EntryResponse{
			ID:            entryID,
			Word:          "example",
			Type:          "WORD",
			Pronunciation: "eg-zam-pul",
		}, nil)

	// Create request with JSON body
	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/entries", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)
}

// Test UpdateEntry
func TestUpdateEntry(t *testing.T) {
	// Setup
	mockService := new(MockEntryService)
	mockLogger := new(MockLogger)
	h := handler.NewEntryHandler(mockService, mockLogger)
	router := setupRouter()
	router.PUT("/entries/:id", h.UpdateEntry)

	// Valid UUID
	entryID := uuid.New()

	// Create mock request body
	requestBody := map[string]interface{}{
		"word": "updated",
		"type": "WORD",
	}

	// Setup mock expectations
	mockService.On("UpdateEntry", mock.Anything, entryID, mock.Anything).Return(
		&response.EntryResponse{
			ID:   entryID,
			Word: "updated",
			Type: "WORD",
		}, nil)

	// Create request with JSON body
	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("PUT", "/entries/"+entryID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

// Test DeleteEntry
func TestDeleteEntry(t *testing.T) {
	// Setup
	mockService := new(MockEntryService)
	mockLogger := new(MockLogger)
	h := handler.NewEntryHandler(mockService, mockLogger)
	router := setupRouter()
	router.DELETE("/entries/:id", h.DeleteEntry)

	// Valid UUID
	entryID := uuid.New()

	// Setup mock expectations
	mockService.On("DeleteEntry", mock.Anything, entryID).Return(nil)

	// Create request
	req, _ := http.NewRequest("DELETE", "/entries/"+entryID.String(), nil)
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNoContent, w.Code)
	mockService.AssertExpectations(t)
}
