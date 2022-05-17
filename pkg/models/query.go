package models

import (
	"encoding/json"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"strconv"
	"strings"
)

type QueryPayload struct {
	Query         string `json:"queryText"`
	Format        string `json:"format"`
	TimeField     string `json:"timeField"`
	Timezone      string `json:"timezone"`
	TimeFormat    string `json:"timeFormat"`
	From          int64  `json:"from"`
	To            int64  `json:"to"`
	MaxDataPoints int64  `json:"maxDataPoints,omitempty"`
	Hide          bool   `json:"hide,omitempty"`
}

func ParsePayload(query backend.DataQuery) (*QueryPayload, error) {
	var payload QueryPayload
	payload.Hide = false

	// Unmarshal the JSON into QueryPayload.
	err := json.Unmarshal(query.JSON, &payload)
	if err != nil {
		return nil, err
	}

	payload.From = query.TimeRange.From.UnixMilli() / 1000
	payload.To = query.TimeRange.To.UnixMilli() / 1000
	if len(payload.Format) == 0 {
		payload.Format = "Table"
	}
	if len(payload.TimeField) == 0 {
		payload.TimeField = "time"
	}
	if len(payload.Timezone) == 0 {
		payload.Timezone = "Asia/Shanghai"
	}
	if len(payload.TimeFormat) == 0 {
		payload.TimeFormat = "2006-01-02 15:04:05"
	} else {
		payload.TimeFormat = strings.ReplaceAll(payload.TimeFormat, "yyyy", "2006")
		payload.TimeFormat = strings.ReplaceAll(payload.TimeFormat, "MM", "01")
		payload.TimeFormat = strings.ReplaceAll(payload.TimeFormat, "dd", "02")
		payload.TimeFormat = strings.ReplaceAll(payload.TimeFormat, "HH", "15")
		payload.TimeFormat = strings.ReplaceAll(payload.TimeFormat, "hh", "03")
		payload.TimeFormat = strings.ReplaceAll(payload.TimeFormat, "mm", "04")
		payload.TimeFormat = strings.ReplaceAll(payload.TimeFormat, "ss", "05")
	}

	log.DefaultLogger.Info("ParsePayload", "Query", payload.Query, "Format", payload.Format,
		"TimeField", payload.TimeField, "Timezone", payload.Timezone, "TimeFormat", payload.TimeFormat,
		"From", strconv.FormatInt(payload.From, 10), "To", strconv.FormatInt(payload.To, 10),
		"MaxDataPoints", payload.MaxDataPoints, "Hide", payload.Hide)

	return &payload, nil
}
