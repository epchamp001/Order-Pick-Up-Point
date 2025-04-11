package entity

import "time"

type Pvz struct {
	ID               string    `json:"id"`
	RegistrationDate time.Time `json:"registration_date"`
	City             string    `json:"city"`
}
