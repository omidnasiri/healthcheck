package model

import "gorm.io/gorm"

type Endpoint struct {
	gorm.Model
	URL         string
	Interval    int // in seconds
	Retries     int // retries before submitting failure
	LastStatus  bool
	ActiveCheck bool
	CheckLogs   []CheckLog
}
