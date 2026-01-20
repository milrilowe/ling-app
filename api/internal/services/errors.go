package services

import "errors"

// Validation errors
var (
	ErrAudioTooShort = errors.New("audio too short")
	ErrAudioInvalid  = errors.New("audio invalid")
)
