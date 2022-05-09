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
	log.DefaultLogger.Info("ParsePayload", "payload.Query", payload.Query, "format", payload.Format,
		"from", strconv.FormatInt(payload.From, 10), "to", strconv.FormatInt(payload.To, 10), "MaxDataPoints", payload.MaxDataPoints)

	return &payload, nil
}
