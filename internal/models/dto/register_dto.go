package dto

// RegisterPostRequest godoc
// @Description Request payload for user registration.
type RegisterPostRequest struct {
	Email    string `json:"email" binding:"required" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"strongpassword123"`
	Role     string `json:"role" binding:"required" example:"moderator"`
}

// RegisterResponse godoc
// @Description Response returned after successful user registration.
type RegisterResponse struct {
	UserID string `json:"user_id" example:"user123"`
}
