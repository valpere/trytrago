package mapper

import (
	"github.com/valpere/trytrago/application/dto/request"
	"github.com/valpere/trytrago/application/dto/response"
	"github.com/valpere/trytrago/domain/model"
)

// UserToResponse maps a domain User model to a UserResponse DTO
func UserToResponse(user *model.User) *response.UserResponse {
	if user == nil {
		return nil
	}

	return &response.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Avatar:    user.Avatar,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

// UserToSummary maps a domain User model to a UserSummary DTO
func UserToSummary(user *model.User) *response.UserSummary {
	if user == nil {
		return nil
	}

	return &response.UserSummary{
		ID:       user.ID,
		Username: user.Username,
		Avatar:   user.Avatar,
	}
}

// CreateUserRequestToModel maps a CreateUserRequest DTO to a domain User model
func CreateUserRequestToModel(req *request.CreateUserRequest) *model.User {
	if req == nil {
		return nil
	}

	return &model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password, // Note: This should be hashed before saving
		Role:     model.RoleUser,
		IsActive: true,
	}
}

// UpdateUserRequestToModel updates a domain User model with data from an UpdateUserRequest DTO
func UpdateUserRequestToModel(user *model.User, req *request.UpdateUserRequest) {
	if user == nil || req == nil {
		return
	}

	if req.Username != "" {
		user.Username = req.Username
	}

	if req.Email != "" {
		user.Email = req.Email
	}

	if req.Password != "" {
		user.Password = req.Password // Note: This should be hashed before saving
	}

	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}
}

// CommentToResponse maps a domain Comment model to a CommentResponse DTO
func CommentToResponse(comment *model.Comment) *response.CommentResponse {
	if comment == nil {
		return nil
	}

	resp := &response.CommentResponse{
		ID:        comment.ID,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt,
		UpdatedAt: comment.UpdatedAt,
	}

	// If user is available in the domain model, map it to UserSummary
	if user, ok := comment.User.(*model.User); ok && user != nil {
		resp.User = *UserToSummary(user)
	}

	return resp
}

// LikeToResponse maps a domain Like model to a LikeResponse DTO
func LikeToResponse(like *model.Like) *response.LikeResponse {
	if like == nil {
		return nil
	}

	resp := &response.LikeResponse{
		ID:         like.ID,
		UserID:     like.UserID,
		TargetType: like.TargetType,
		TargetID:   like.TargetID,
		CreatedAt:  like.CreatedAt,
	}

	// If user is available in the domain model, map it to UserSummary
	if user, ok := like.User.(*model.User); ok && user != nil {
		resp.User = *UserToSummary(user)
	}

	return resp
}

// CommentListToResponse maps a slice of domain Comment models to a slice of CommentResponse DTOs
func CommentListToResponse(comments []model.Comment) []*response.CommentResponse {
	if len(comments) == 0 {
		return []*response.CommentResponse{}
	}

	result := make([]*response.CommentResponse, len(comments))
	for i, comment := range comments {
		result[i] = CommentToResponse(&comment)
	}

	return result
}

// LikeListToResponse maps a slice of domain Like models to a slice of LikeResponse DTOs
func LikeListToResponse(likes []model.Like) []*response.LikeResponse {
	if len(likes) == 0 {
		return []*response.LikeResponse{}
	}

	result := make([]*response.LikeResponse, len(likes))
	for i, like := range likes {
		result[i] = LikeToResponse(&like)
	}

	return result
}
