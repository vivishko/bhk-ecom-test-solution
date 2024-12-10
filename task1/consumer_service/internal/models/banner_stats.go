package models

import "time"

type BannerStat struct {
	BannerID      int       		`db:"banner_id"`
	HourTimestamp time.Time 		`db:"hour_timestamp"`
	Counts        map[string]int 	`db:"counts"` 
}