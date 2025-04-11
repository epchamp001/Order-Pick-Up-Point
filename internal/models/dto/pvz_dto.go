package dto

import "time"

// PvzDTO godoc
// @Description Represents a PVZ (pickup point) with its information.
type PvzDTO struct {
	Id               string    `json:"id,omitempty" example:"pvz789"`
	RegistrationDate time.Time `json:"registrationDate,omitempty" example:"2025-04-09T12:00:00Z"`
	City             string    `json:"city" example:"Moscow"`
}

// PvzGet200ResponseInnerReceptionsInner godoc
// @Description Represents a reception within a PVZ, including its associated products.
type PvzGet200ResponseInnerReceptionsInner struct {
	Reception *ReceptionDTO `json:"reception,omitempty"`
	Products  []ProductDTO  `json:"products,omitempty"`
}

// PvzGet200ResponseInner godoc
// @Description Response model for retrieving PVZ information, including receptions and products.
type PvzGet200ResponseInner struct {
	Pvz        *PvzDTO                                 `json:"pvz,omitempty"`
	Receptions []PvzGet200ResponseInnerReceptionsInner `json:"receptions,omitempty"`
}

// CreatePvzPostRequest godoc
// @Description Request payload for creating a new PVZ.
type CreatePvzPostRequest struct {
	City string `json:"city" binding:"required" example:"Moscow"`
}

// CreatePvzResponse godoc
// @Description Response returned after successfully creating a PVZ.
type CreatePvzResponse struct {
	PvzId string `json:"id" example:"pvz789"`
}
