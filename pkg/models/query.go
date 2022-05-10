package models

import (
	"encoding/json"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"strconv"
)

type QueryPayload struct {
	Query         string `json:"queryText"`
	Format        string `json:"format"`
	TimeField     string `json:"timeField"`
	Timezone      string `json:"timeFormat"`
	From          int64  `json:"from"`
	To            int64  `json:"to"`
	MaxDataPoints int64  `json:"-"`
}

func ParsePayload(query backend.DataQuery) (*QueryPayload, error) {
	var payload QueryPayload

	// Unmarshal the JSON into QueryPayload.
	err := json.Unmarshal(query.JSON, &payload)
	if err != nil {
		return nil, err
	}

	payload.From = query.TimeRange.From.UnixMilli() / 1000
	payload.To = query.TimeRange.To.UnixMilli() / 1000
	payload.MaxDataPoints = query.MaxDataPoints
	if len(payload.Format) == 0 {
		payload.Format = "Table"
	}
	if len(payload.TimeField) == 0 {
		payload.TimeField = "time"
	}
	if len(payload.Timezone) == 0 {
		payload.Timezone = "Asia/Shanghai"
	}
	log.DefaultLogger.Info("ParsePayload", "payload.Query", payload.Query,
		"format", payload.Format, "TimeField", payload.TimeField, "Timezone", payload.Timezone,
		"from", strconv.FormatInt(payload.From, 10), "to", strconv.FormatInt(payload.To, 10),
		"MaxDataPoints", payload.MaxDataPoints)

	return &payload, nil
}
