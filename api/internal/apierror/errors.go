package apierror

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Error codes
const (
	// Validation errors
	CodeValidationFailed = "VALIDATION_FAILED"
	CodeInvalidRequest   = "INVALID_REQUEST"
	CodeInvalidThreadID  = "INVALID_THREAD_ID"
	CodeMissingAudioFile = "MISSING_AUDIO_FILE"
	CodeFileTooLarge     = "FILE_TOO_LARGE"

	// Auth errors
	CodeInvalidCredentials = "AUTH_INVALID_CREDENTIALS"
	CodeEmailTaken         = "AUTH_EMAIL_TAKEN"
	CodeSessionExpired     = "AUTH_SESSION_EXPIRED"
	CodeUnauthorized       = "AUTH_UNAUTHORIZED"

	// Payment errors
	CodeInsufficientCredits = "INSUFFICIENT_CREDITS"

	// Not found errors
	CodeThreadNotFound   = "THREAD_NOT_FOUND"
	CodeUserNotFound     = "USER_NOT_FOUND"
	CodeResourceNotFound = "RESOURCE_NOT_FOUND"

	// External service errors
	CodeExternalServiceError  = "EXTERNAL_SERVICE_ERROR"
	CodeAudioProcessingFailed = "AUDIO_PROCESSING_FAILED"

	// Internal errors
	CodeInternalError = "INTERNAL_ERROR"
)

// AppError represents a structured API error
type AppError struct {
	Code    string      `json:"code"`
	Message string      `json:"error"`
	Status  int         `json:"-"`
	Details interface{} `json:"details,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	return e.Message
}

// RespondWithError writes a standardized error response to the gin context
func RespondWithError(c *gin.Context, err *AppError) {
	c.JSON(err.Status, err)
}

// Validation errors

func ValidationFailed(message string) *AppError {
	return &AppError{
		Code:    CodeValidationFailed,
		Message: message,
		Status:  http.StatusBadRequest,
	}
}

func InvalidRequest(message string) *AppError {
	return &AppError{
		Code:    CodeInvalidRequest,
		Message: message,
		Status:  http.StatusBadRequest,
	}
}

func InvalidThreadID() *AppError {
	return &AppError{
		Code:    CodeInvalidThreadID,
		Message: "Invalid thread ID",
		Status:  http.StatusBadRequest,
	}
}

func MissingAudioFile() *AppError {
	return &AppError{
		Code:    CodeMissingAudioFile,
		Message: "Audio file is required",
		Status:  http.StatusBadRequest,
	}
}

func FileTooLarge(maxSize int64) *AppError {
	return &AppError{
		Code:    CodeFileTooLarge,
		Message: "Audio file too large",
		Status:  http.StatusRequestEntityTooLarge,
		Details: map[string]int64{"maxSize": maxSize},
	}
}

// Auth errors

func InvalidCredentials() *AppError {
	return &AppError{
		Code:    CodeInvalidCredentials,
		Message: "Invalid email or password",
		Status:  http.StatusUnauthorized,
	}
}

func EmailTaken() *AppError {
	return &AppError{
		Code:    CodeEmailTaken,
		Message: "Email already registered",
		Status:  http.StatusConflict,
	}
}

func SessionExpired() *AppError {
	return &AppError{
		Code:    CodeSessionExpired,
		Message: "Session expired",
		Status:  http.StatusUnauthorized,
	}
}

func Unauthorized(message string) *AppError {
	if message == "" {
		message = "Authentication required"
	}
	return &AppError{
		Code:    CodeUnauthorized,
		Message: message,
		Status:  http.StatusUnauthorized,
	}
}

// Payment errors

func InsufficientCredits(creditsNeeded int) *AppError {
	return &AppError{
		Code:    CodeInsufficientCredits,
		Message: "Insufficient credits",
		Status:  http.StatusPaymentRequired,
		Details: map[string]int{"creditsNeeded": creditsNeeded},
	}
}

// Not found errors

func ThreadNotFound() *AppError {
	return &AppError{
		Code:    CodeThreadNotFound,
		Message: "Thread not found",
		Status:  http.StatusNotFound,
	}
}

func UserNotFound() *AppError {
	return &AppError{
		Code:    CodeUserNotFound,
		Message: "User not found",
		Status:  http.StatusNotFound,
	}
}

func ResourceNotFound(resource string) *AppError {
	message := "Resource not found"
	if resource != "" {
		message = resource + " not found"
	}
	return &AppError{
		Code:    CodeResourceNotFound,
		Message: message,
		Status:  http.StatusNotFound,
	}
}

// External service errors

func ExternalServiceError() *AppError {
	return &AppError{
		Code:    CodeExternalServiceError,
		Message: "Something went wrong. Please try again.",
		Status:  http.StatusInternalServerError,
	}
}

func AudioProcessingFailed() *AppError {
	return &AppError{
		Code:    CodeAudioProcessingFailed,
		Message: "Unable to process audio. Please try again.",
		Status:  http.StatusInternalServerError,
	}
}

// Internal errors

func InternalError(message string) *AppError {
	if message == "" {
		message = "An unexpected error occurred"
	}
	return &AppError{
		Code:    CodeInternalError,
		Message: message,
		Status:  http.StatusInternalServerError,
	}
}
