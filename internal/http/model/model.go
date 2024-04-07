// Package model contains http api related data types.
package model

// ShortenRequest is single URL shorten request.
type ShortenRequest struct {
	URL string `json:"url"`
}

// ShortenResponse is a single URL shorten response.
type ShortenResponse struct {
	Result string `json:"result"`
}
