package models

type QueryInfo struct {
	Query string `json:"query"`
	Xcol  string `json:"xcol"`
	Ycol  string `json:"ycol"`
}
