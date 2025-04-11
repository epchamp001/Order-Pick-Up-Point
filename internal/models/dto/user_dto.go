package dto

// User godoc
// @Description Represents a user in the system.
type User struct {
	Id    string `json:"id,omitempty" example:"user123"`
	Email string `json:"email" example:"user@example.com"`
	Role  string `json:"role" example:"moderator"`
}
