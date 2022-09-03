package models

import (
	"time"
)

type DataRecord struct {
	Time         time.Time
	Number       *float64
	FieldValDict map[string]string
}
