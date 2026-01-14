package cli

import (
	"errors"
	"net"

	"github.com/gcarthew/ajira/internal/api"
)

// Exit codes for the CLI.
const (
	ExitSuccess   = 0 // Successful execution
	ExitUserError = 1 // User/input error (invalid args, missing required values)
	ExitAPIError  = 2 // API error (4xx/5xx responses, except auth)
	ExitNetError  = 3 // Network/connection error
	ExitAuthError = 4 // Authentication error (401, 403)
	ExitPartial   = 5 // Partial failure in batch operations
)

// ExitError wraps an error with an exit code.
type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return ""
}

func (e *ExitError) Unwrap() error {
	return e.Err
}

// NewExitError creates a new ExitError with the given code and error.
func NewExitError(code int, err error) *ExitError {
	return &ExitError{Code: code, Err: err}
}

// ExitCodeFromError determines the appropriate exit code for an error.
// Returns ExitSuccess (0) if err is nil.
func ExitCodeFromError(err error) int {
	if err == nil {
		return ExitSuccess
	}

	// Check for ExitError first
	var exitErr *ExitError
	if errors.As(err, &exitErr) {
		return exitErr.Code
	}

	// Check for API errors
	var apiErr *api.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.StatusCode {
		case 401, 403:
			return ExitAuthError
		default:
			return ExitAPIError
		}
	}

	// Check for network errors
	var netErr net.Error
	if errors.As(err, &netErr) {
		return ExitNetError
	}

	// Check for DNS errors
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return ExitNetError
	}

	// Default to user error for unrecognised errors
	return ExitUserError
}
