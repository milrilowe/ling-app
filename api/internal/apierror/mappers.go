package apierror

import (
	"errors"

	"ling-app/api/internal/repository"
	"ling-app/api/internal/services"
	"ling-app/api/internal/services/auth"
)

// FromAuthError maps auth service errors to AppError
func FromAuthError(err error) *AppError {
	switch {
	case errors.Is(err, auth.ErrInvalidCredentials):
		return InvalidCredentials()
	case errors.Is(err, auth.ErrEmailTaken):
		return EmailTaken()
	case errors.Is(err, auth.ErrUserNotFound):
		return UserNotFound()
	case errors.Is(err, auth.ErrSessionNotFound):
		return SessionExpired()
	default:
		return InternalError("")
	}
}

// FromCreditsError maps credits service errors to AppError
func FromCreditsError(err error, creditsNeeded int) *AppError {
	switch {
	case errors.Is(err, services.ErrInsufficientCredits):
		return InsufficientCredits(creditsNeeded)
	case errors.Is(err, services.ErrCreditsNotFound):
		return ResourceNotFound("Credits")
	default:
		return InternalError("")
	}
}

// FromRepositoryError maps repository errors to AppError
func FromRepositoryError(err error, resource string) *AppError {
	if errors.Is(err, repository.ErrNotFound) {
		return ResourceNotFound(resource)
	}
	return InternalError("")
}

// IsNotFoundError checks if error is a not found error
func IsNotFoundError(err error) bool {
	return errors.Is(err, repository.ErrNotFound)
}
