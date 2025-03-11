package mapper

import (
	"github.com/valpere/trytrago/application/dto/response"
	"github.com/valpere/trytrago/domain/database"
)

// TranslationToResponse maps a domain Translation model to a TranslationResponse DTO
func TranslationToResponse(translation *database.Translation) *response.TranslationResponse {
	if translation == nil {
		return nil
	}

	return &response.TranslationResponse{
		ID:         translation.ID,
		MeaningID:  translation.MeaningID,
		LanguageID: translation.LanguageID,
		Text:       translation.Text,
		LikesCount: 0, // To be implemented with actual count
		CreatedAt:  translation.CreatedAt,
		UpdatedAt:  translation.UpdatedAt,
	}
}

// TranslationListToResponse maps a slice of domain Translation models to a slice of TranslationResponse DTOs
func TranslationListToResponse(translations []database.Translation) []*response.TranslationResponse {
	if len(translations) == 0 {
		return []*response.TranslationResponse{}
	}

	result := make([]*response.TranslationResponse, len(translations))
	for i, translation := range translations {
		result[i] = TranslationToResponse(&translation)
	}

	return result
}
