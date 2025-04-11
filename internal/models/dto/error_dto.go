package dto

// Error godoc
// @Description Standard error response containing an error message.
type Error struct {
	Message string `json:"message" example:"invalid request body"`
}
