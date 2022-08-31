package models

import (
	"time"
)

type DataRecord struct {
	Time               time.Time
	FieldNumberValDict map[string]*float64
	FieldValDict       map[string]string
}
