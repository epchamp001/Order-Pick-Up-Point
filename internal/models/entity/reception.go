package entity

import "time"

type Reception struct {
	ID       string    `json:"id"`
	DateTime time.Time `json:"date_time"`
	PvzID    string    `json:"pvz_id"`
	Status   string    `json:"status"`
}
