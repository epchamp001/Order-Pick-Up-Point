package dto

// DummyLoginPostRequest godoc
// @Description Request payload for dummy login. Provide a desired user role ("client", "employee", "moderator") to obtain a JWT token.
type DummyLoginPostRequest struct {
	Role string `json:"role" example:"employee"`
}

// TokenResponse godoc
// @Description Response containing a JWT token.
type TokenResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}
