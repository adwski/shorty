// Package model holds common data structures.
package model

// StatsResponse is a response message for storage stats endpoint.
type StatsResponse struct {
	URLs  int `json:"urls"`
	Users int `json:"users"`
}
