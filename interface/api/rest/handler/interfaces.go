package handler

import (
    "github.com/gin-gonic/gin"
)

// EntryHandlerInterface defines the interface for entry-related endpoints
type EntryHandlerInterface interface {
    ListEntries(c *gin.Context)
    GetEntry(c *gin.Context)
    CreateEntry(c *gin.Context)
    UpdateEntry(c *gin.Context)
    DeleteEntry(c *gin.Context)
    GetMeaning(c *gin.Context)
    ListMeanings(c *gin.Context)
    AddMeaning(c *gin.Context)
    UpdateMeaning(c *gin.Context)
    DeleteMeaning(c *gin.Context)
    AddMeaningComment(c *gin.Context)
    ToggleMeaningLike(c *gin.Context)
}

// TranslationHandlerInterface defines the interface for translation-related endpoints
type TranslationHandlerInterface interface {
    ListTranslations(c *gin.Context)
    CreateTranslation(c *gin.Context)
    UpdateTranslation(c *gin.Context)
    DeleteTranslation(c *gin.Context)
    AddTranslationComment(c *gin.Context)
    ToggleTranslationLike(c *gin.Context)
}

// UserHandlerInterface defines the interface for user-related endpoints
type UserHandlerInterface interface {
    CreateUser(c *gin.Context)
    GetCurrentUser(c *gin.Context)
    UpdateCurrentUser(c *gin.Context)
    DeleteCurrentUser(c *gin.Context)
    Login(c *gin.Context)
    RefreshToken(c *gin.Context)
    ListUserEntries(c *gin.Context)
    ListUserTranslations(c *gin.Context)
    ListUserComments(c *gin.Context)
    ListUserLikes(c *gin.Context)
}
