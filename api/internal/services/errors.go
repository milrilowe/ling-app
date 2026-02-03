package services

import "errors"

// Validation errors
var (
	ErrAudioTooShort = errors.New("audio too short")
	ErrAudioTooLong  = errors.New("audio too long")
	ErrAudioInvalid  = errors.New("audio invalid")
)
