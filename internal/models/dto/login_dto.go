package dto

// LoginPostRequest godoc
// @Description Request payload for login using email and password.
type LoginPostRequest struct {
	Email    string `json:"email" binding:"required" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"strongpassword123"`
}
