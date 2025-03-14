package service

import (
    "context"
    "fmt"
    "time"

    "github.com/google/uuid"
    "github.com/valpere/trytrago/application/dto/request"
    "github.com/valpere/trytrago/application/dto/response"
    "github.com/valpere/trytrago/application/mapper"
    "github.com/valpere/trytrago/domain/database"
    "github.com/valpere/trytrago/domain/database/repository"
    "github.com/valpere/trytrago/domain/logging"
    "github.com/valpere/trytrago/domain/model"
)

// translationService implements the TranslationService interface
type translationService struct {
    repo   repository.Repository
    logger logging.Logger
}

// NewTranslationService creates a new instance of TranslationService
func NewTranslationService(repo repository.Repository, logger logging.Logger) TranslationService {
    return &translationService{
        repo:   repo,
        logger: logger.With(logging.String("service", "translation")),
    }
}

// CreateTranslation implements TranslationService.CreateTranslation
func (s *translationService) CreateTranslation(ctx context.Context, meaningID uuid.UUID, req *request.CreateTranslationRequest) (*response.TranslationResponse, error) {
    s.logger.Debug("creating translation",
        logging.String("meaningID", meaningID.String()),
        logging.String("languageID", req.LanguageID),
    )

    // Verify the meaning exists - first get all entries
    params := repository.ListParams{
        Limit: 100, // Reasonable limit
    }
    entries, err := s.repo.ListEntries(ctx, params)
    if err != nil {
        s.logger.Error("failed to list entries to find meaning",
            logging.Error(err),
            logging.String("meaningID", meaningID.String()),
        )
        return nil, fmt.Errorf("failed to find meaning: %w", err)
    }

    // Find the meaning
    var meaning *database.Meaning
    var entry *database.Entry

    for i := range entries {
        e := &entries[i]
        for j := range e.Meanings {
            m := &e.Meanings[j]
            if m.ID == meaningID {
                meaning = m
                entry = e
                break
            }
        }
        if meaning != nil {
            break
        }
    }

    if meaning == nil {
        return nil, database.ErrEntryNotFound
    }

    // Create translation
    now := time.Now().UTC()
    translation := &database.Translation{
        ID:         uuid.New(),
        MeaningID:  meaningID,
        LanguageID: req.LanguageID,
        Text:       req.Text,
        CreatedAt:  now,
        UpdatedAt:  now,
    }

    // Add to meaning's translations
    meaning.Translations = append(meaning.Translations, *translation)

    // Update the meaning with the new translation
    if err := s.repo.UpdateEntry(ctx, entry); err != nil {
        s.logger.Error("failed to save translation",
            logging.Error(err),
            logging.String("meaningID", meaningID.String()),
        )
        return nil, fmt.Errorf("failed to save translation: %w", err)
    }

    // Create response
    resp := mapper.TranslationToResponse(translation)
    return resp, nil
}

// UpdateTranslation implements TranslationService.UpdateTranslation
func (s *translationService) UpdateTranslation(ctx context.Context, id uuid.UUID, req *request.UpdateTranslationRequest) (*response.TranslationResponse, error) {
    s.logger.Debug("updating translation", logging.String("id", id.String()))

    // Find the translation by checking all meanings
    // This is inefficient but necessary with the current model structure
    // In a real implementation, we would need a direct repository method to get a translation

    // Get all entries that might contain this translation
    params := repository.ListParams{
        Limit: 100, // Reasonable limit, adjust based on actual requirements
    }
    entries, err := s.repo.ListEntries(ctx, params)
    if err != nil {
        s.logger.Error("failed to list entries to find translation",
            logging.Error(err),
            logging.String("translationID", id.String()),
        )
        return nil, fmt.Errorf("failed to find translation: %w", err)
    }

    // Search for the translation
    var foundTranslation *database.Translation
    var foundEntry *database.Entry

    for i := range entries {
        entry := &entries[i]
        for j := range entry.Meanings {
            meaning := &entry.Meanings[j]
            for k := range meaning.Translations {
                translation := &meaning.Translations[k]
                if translation.ID == id {
                    foundTranslation = translation
                    foundEntry = entry
                    break
                }
            }
            if foundTranslation != nil {
                break
            }
        }
        if foundTranslation != nil {
            break
        }
    }

    if foundTranslation == nil {
        return nil, database.ErrEntryNotFound
    }

    // Update the translation fields
    foundTranslation.Text = req.Text
    foundTranslation.UpdatedAt = time.Now().UTC()

    // Save the updated entry
    if err := s.repo.UpdateEntry(ctx, foundEntry); err != nil {
        s.logger.Error("failed to update translation",
            logging.Error(err),
            logging.String("translationID", id.String()),
        )
        return nil, fmt.Errorf("failed to update translation: %w", err)
    }

    // Create response
    resp := mapper.TranslationToResponse(foundTranslation)
    return resp, nil
}

// DeleteTranslation implements TranslationService.DeleteTranslation
func (s *translationService) DeleteTranslation(ctx context.Context, id uuid.UUID) error {
    s.logger.Debug("deleting translation", logging.String("id", id.String()))

    // Similar to UpdateTranslation, we need to find the translation first
    // Then remove it from its parent meaning

    // Get all entries that might contain this translation
    params := repository.ListParams{
        Limit: 100, // Reasonable limit, adjust based on actual requirements
    }
    entries, err := s.repo.ListEntries(ctx, params)
    if err != nil {
        s.logger.Error("failed to list entries to find translation for deletion",
            logging.Error(err),
            logging.String("translationID", id.String()),
        )
        return fmt.Errorf("failed to find translation: %w", err)
    }

    // Search for the translation
    var foundTranslation *database.Translation
    var foundMeaning *database.Meaning
    var foundEntry *database.Entry
    var translationIndex int

    for i := range entries {
        entry := &entries[i]
        for j := range entry.Meanings {
            meaning := &entry.Meanings[j]
            for k := range meaning.Translations {
                translation := &meaning.Translations[k]
                if translation.ID == id {
                    foundTranslation = translation
                    foundMeaning = meaning
                    foundEntry = entry
                    translationIndex = k
                    break
                }
            }
            if foundTranslation != nil {
                break
            }
        }
        if foundTranslation != nil {
            break
        }
    }

    if foundTranslation == nil {
        return database.ErrEntryNotFound
    }

    // Remove the translation from the meaning
    foundMeaning.Translations = append(
        foundMeaning.Translations[:translationIndex],
        foundMeaning.Translations[translationIndex+1:]...,
    )

    // Save the updated entry
    if err := s.repo.UpdateEntry(ctx, foundEntry); err != nil {
        s.logger.Error("failed to delete translation",
            logging.Error(err),
            logging.String("translationID", id.String()),
        )
        return fmt.Errorf("failed to delete translation: %w", err)
    }

    return nil
}

// ListTranslations implements TranslationService.ListTranslations
func (s *translationService) ListTranslations(ctx context.Context, meaningID uuid.UUID, langID string) (*response.TranslationListResponse, error) {
    s.logger.Debug("listing translations",
        logging.String("meaningID", meaningID.String()),
        logging.String("languageID", langID),
    )

    // Get the meaning with its translations - first get all entries
    params := repository.ListParams{
        Limit: 100, // Reasonable limit
    }
    entries, err := s.repo.ListEntries(ctx, params)
    if err != nil {
        s.logger.Error("failed to list entries to find meaning",
            logging.Error(err),
            logging.String("meaningID", meaningID.String()),
        )
        return nil, fmt.Errorf("failed to find meaning: %w", err)
    }

    // Find the meaning
    var meaning *database.Meaning

    for i := range entries {
        e := &entries[i]
        for j := range e.Meanings {
            m := &e.Meanings[j]
            if m.ID == meaningID {
                meaning = m
                break
            }
        }
        if meaning != nil {
            break
        }
    }

    if meaning == nil {
        return nil, database.ErrEntryNotFound
    }

    // Filter translations by language if specified
    var translations []database.Translation
    if langID != "" {
        for _, t := range meaning.Translations {
            if t.LanguageID == langID {
                translations = append(translations, t)
            }
        }
    } else {
        translations = meaning.Translations
    }

    // Create response
    resp := &response.TranslationListResponse{
        Translations: make([]*response.TranslationResponse, len(translations)),
        Total:        len(translations),
        Limit:        100,
        Offset:       0,
    }

    for i, t := range translations {
        resp.Translations[i] = mapper.TranslationToResponse(&t)
    }

    return resp, nil
}

// AddTranslationComment implements TranslationService.AddTranslationComment
func (s *translationService) AddTranslationComment(ctx context.Context, translationID uuid.UUID, req *request.CreateCommentRequest) (*response.CommentResponse, error) {
    s.logger.Debug("adding comment to translation",
        logging.String("translationID", translationID.String()),
        logging.String("userID", req.UserID.String()),
    )

    // Find the translation by ID
    // Note: In a real implementation, you would have a direct repository method to get a translation by ID

    // Get all entries that might contain this translation
    params := repository.ListParams{
        Limit: 100, // Reasonable limit
    }
    entries, err := s.repo.ListEntries(ctx, params)
    if err != nil {
        s.logger.Error("failed to list entries to find translation",
            logging.Error(err),
            logging.String("translationID", translationID.String()),
        )
        return nil, fmt.Errorf("failed to find translation: %w", err)
    }

    // Search for the translation
    var foundTranslation *database.Translation

    for i := range entries {
        entry := entries[i]
        for j := range entry.Meanings {
            meaning := entry.Meanings[j]
            for k := range meaning.Translations {
                translation := &meaning.Translations[k]
                if translation.ID == translationID {
                    foundTranslation = translation
                    break
                }
            }
            if foundTranslation != nil {
                break
            }
        }
        if foundTranslation != nil {
            break
        }
    }

    if foundTranslation == nil {
        return nil, database.ErrEntryNotFound
    }

    // Create a new comment
    comment := model.Comment{
        ID:         uuid.New(),
        UserID:     req.UserID,
        TargetType: "translation",
        TargetID:   translationID,
        Content:    req.Content,
        CreatedAt:  time.Now().UTC(),
        UpdatedAt:  time.Now().UTC(),
    }

    // In a real implementation, you would save this comment to a dedicated comments table
    // For now, we'll create a mock response

    // Create user for the comment (would come from a user repository in a real implementation)
    user := &model.User{
        ID:       req.UserID,
        Username: "user" + req.UserID.String()[0:8], // Mock username
    }

    // Create response
    resp := &response.CommentResponse{
        ID:      comment.ID,
        Content: comment.Content,
        User: response.UserSummary{
            ID:       user.ID,
            Username: user.Username,
        },
        CreatedAt: comment.CreatedAt,
        UpdatedAt: comment.UpdatedAt,
    }

    return resp, nil
}

// ToggleTranslationLike implements TranslationService.ToggleTranslationLike
func (s *translationService) ToggleTranslationLike(ctx context.Context, translationID uuid.UUID, userID uuid.UUID) error {
    s.logger.Debug("toggling like on translation",
        logging.String("translationID", translationID.String()),
        logging.String("userID", userID.String()),
    )

    // Find the translation by ID
    params := repository.ListParams{
        Limit: 100, // Reasonable limit
    }
    entries, err := s.repo.ListEntries(ctx, params)
    if err != nil {
        s.logger.Error("failed to list entries to find translation",
            logging.Error(err),
            logging.String("translationID", translationID.String()),
        )
        return fmt.Errorf("failed to find translation: %w", err)
    }

    // Search for the translation
    var foundTranslation *database.Translation

    for i := range entries {
        entry := entries[i]
        for j := range entry.Meanings {
            meaning := entry.Meanings[j]
            for k := range meaning.Translations {
                translation := &meaning.Translations[k]
                if translation.ID == translationID {
                    foundTranslation = translation
                    break
                }
            }
            if foundTranslation != nil {
                break
            }
        }
        if foundTranslation != nil {
            break
        }
    }

    if foundTranslation == nil {
        return database.ErrEntryNotFound
    }

    // In a real implementation, you would:
    // 1. Check if the user has already liked this translation
    // 2. If yes, remove the like
    // 3. If no, add a new like
    // 4. Update the likes count for the translation

    // Create or toggle like (would be saved to a database in a real implementation)
    like := model.Like{
        ID:         uuid.New(),
        UserID:     userID,
        TargetType: "translation",
        TargetID:   translationID,
        CreatedAt:  time.Now().UTC(),
    }

    // Placeholder for logging - in a real implementation, this would be saved
    s.logger.Info("like processed",
        logging.String("likeID", like.ID.String()),
        logging.String("translationID", translationID.String()),
        logging.String("userID", userID.String()),
    )

    return nil
}
