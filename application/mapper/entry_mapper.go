package mapper

import (
	"github.com/valpere/trytrago/application/dto/response"
	"github.com/valpere/trytrago/domain/database"
)

// EntryToResponse maps a domain Entry model to an EntryResponse DTO
func EntryToResponse(entry *database.Entry) *response.EntryResponse {
	if entry == nil {
		return nil
	}

	resp := &response.EntryResponse{
		ID:            entry.ID,
		Word:          entry.Word,
		Type:          string(entry.Type),
		Pronunciation: entry.Pronunciation,
		CreatedAt:     entry.CreatedAt,
		UpdatedAt:     entry.UpdatedAt,
	}

	// Map meanings if available
	if len(entry.Meanings) > 0 {
		resp.Meanings = make([]response.MeaningResponse, len(entry.Meanings))
		
		for i, meaning := range entry.Meanings {
			resp.Meanings[i] = *MeaningToResponse(&meaning)
		}
	}

	return resp
}

// MeaningToResponse maps a domain Meaning model to a MeaningResponse DTO
func MeaningToResponse(meaning *database.Meaning) *response.MeaningResponse {
	if meaning == nil {
		return nil
	}

	resp := &response.MeaningResponse{
		ID:          meaning.ID,
		EntryID:     meaning.EntryID,
		Description: meaning.Description,
		CreatedAt:   meaning.CreatedAt,
		UpdatedAt:   meaning.UpdatedAt,
		LikesCount:  0, // To be implemented with actual count
	}

	// Map examples if available
	if len(meaning.Examples) > 0 {
		resp.Examples = make([]response.ExampleResponse, len(meaning.Examples))
		
		for i, example := range meaning.Examples {
			resp.Examples[i] = response.ExampleResponse{
				ID:        example.ID,
				Text:      example.Text,
				Context:   example.Context,
				CreatedAt: example.CreatedAt,
				UpdatedAt: example.UpdatedAt,
			}
		}
	}

	// Map translations if available
	if len(meaning.Translations) > 0 {
		resp.Translations = make([]response.TranslationResponse, len(meaning.Translations))
		
		for i, translation := range meaning.Translations {
			resp.Translations[i] = *TranslationToResponse(&translation)
		}
	}

	return resp
}

// ExampleToResponse maps a domain Example model to an ExampleResponse DTO
func ExampleToResponse(example *database.Example) *response.ExampleResponse {
	if example == nil {
		return nil
	}

	return &response.ExampleResponse{
		ID:        example.ID,
		Text:      example.Text,
		Context:   example.Context,
		CreatedAt: example.CreatedAt,
		UpdatedAt: example.UpdatedAt,
	}
}
