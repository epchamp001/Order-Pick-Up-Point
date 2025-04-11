package dto

import "time"

// ReceptionDTO godoc
// @Description Represents a reception record for goods.
type ReceptionDTO struct {
	Id       string    `json:"id,omitempty" example:"recv456"`
	DateTime time.Time `json:"dateTime" example:"2025-04-09T15:04:05Z"`
	PvzId    string    `json:"pvzId" example:"pvz789"`
	Status   string    `json:"status" example:"in_progress"`
}

// CloseReceptionResponse godoc
// @Description Response returned after successfully closing a reception.
type CloseReceptionResponse struct {
	ReceptionId string `json:"receptionId" example:"recv456"`
}

// CreateReceptionRequest godoc
// @Description Request payload for creating a new reception.
// The DateTime field is of type time.Time; if not provided, the current time is used.
type CreateReceptionRequest struct {
	PvzId    string    `json:"pvzId" binding:"required" example:"pvz789"`
	DateTime time.Time `json:"dateTime" example:"2025-04-09T15:04:05Z"`
}

// CreateReceptionResponse godoc
// @Description Response returned after a reception is created.
type CreateReceptionResponse struct {
	ReceptionId string `json:"receptionId" example:"recv456"`
}
