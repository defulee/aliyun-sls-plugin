package models

import (
	"time"
)

type DataRecord struct {
	Time   time.Time
	Values map[string]float64
}
