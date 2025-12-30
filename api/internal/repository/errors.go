package repository

import "errors"

// Repository errors - these are returned by repository methods
// and can be checked by services to handle specific cases.
var (
	ErrNotFound = errors.New("record not found")
)
