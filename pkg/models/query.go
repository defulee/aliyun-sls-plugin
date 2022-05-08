package models

type QueryPayload struct {
	Query         string `json:"queryText"`
	Format        string `json:"format"`
	From          int64  `json:"from"`
	To            int64  `json:"to"`
	MaxDataPoints int64  `json:"-"`
}
