package entity

import "time"

type Product struct {
	ID          string    `json:"id"`
	DateTime    time.Time `json:"date_time"`
	Type        string    `json:"type"`
	ReceptionID string    `json:"reception_id"`
}
