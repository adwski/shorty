// Package shortener is shortened URL management service.
// It provides API to retrieve, store and delete URLs.
//
// Deletion can be done only with batch operation and by design it is delayed
// and executed via Flusher queue.
//
// It also provides user level isolation of URLs.
package shortener
