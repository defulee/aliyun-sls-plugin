package models

import (
	"time"
)

type DataRecord struct {
	Time               time.Time
	FieldNumberValDict map[string]float64
	FieldStringValDict map[string]string
}
