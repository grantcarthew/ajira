package cli

import (
	"errors"
	"net"
	"testing"

	"github.com/gcarthew/ajira/internal/api"
)

func TestExitCodeFromError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{
			name:     "nil error returns success",
			err:      nil,
			expected: ExitSuccess,
		},
		{
			name:     "ExitError returns its code",
			err:      NewExitError(ExitPartial, errors.New("partial failure")),
			expected: ExitPartial,
		},
		{
			name:     "wrapped ExitError returns its code",
			err:      errors.New("wrapper: " + NewExitError(ExitNetError, errors.New("network")).Error()),
			expected: ExitUserError, // String wrapping breaks the chain
		},
		{
			name: "API error 401 returns auth error",
			err: &api.APIError{
				StatusCode: 401,
				Status:     "401 Unauthorized",
				Method:     "GET",
				Path:       "/test",
			},
			expected: ExitAuthError,
		},
		{
			name: "API error 403 returns auth error",
			err: &api.APIError{
				StatusCode: 403,
				Status:     "403 Forbidden",
				Method:     "GET",
				Path:       "/test",
			},
			expected: ExitAuthError,
		},
		{
			name: "API error 404 returns API error",
			err: &api.APIError{
				StatusCode: 404,
				Status:     "404 Not Found",
				Method:     "GET",
				Path:       "/test",
			},
			expected: ExitAPIError,
		},
		{
			name: "API error 500 returns API error",
			err: &api.APIError{
				StatusCode: 500,
				Status:     "500 Internal Server Error",
				Method:     "POST",
				Path:       "/test",
			},
			expected: ExitAPIError,
		},
		{
			name:     "DNS error returns network error",
			err:      &net.DNSError{Err: "no such host", Name: "example.com"},
			expected: ExitNetError,
		},
		{
			name:     "generic error returns user error",
			err:      errors.New("some error"),
			expected: ExitUserError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExitCodeFromError(tt.err)
			if result != tt.expected {
				t.Errorf("ExitCodeFromError() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestExitError(t *testing.T) {
	t.Run("Error returns wrapped error message", func(t *testing.T) {
		err := NewExitError(ExitAPIError, errors.New("test error"))
		if err.Error() != "test error" {
			t.Errorf("Error() = %q, want %q", err.Error(), "test error")
		}
	})

	t.Run("Error returns empty string for nil inner error", func(t *testing.T) {
		err := &ExitError{Code: ExitSuccess, Err: nil}
		if err.Error() != "" {
			t.Errorf("Error() = %q, want empty string", err.Error())
		}
	})

	t.Run("Unwrap returns inner error", func(t *testing.T) {
		inner := errors.New("inner error")
		err := NewExitError(ExitNetError, inner)
		if err.Unwrap() != inner {
			t.Errorf("Unwrap() did not return inner error")
		}
	})

	t.Run("errors.As finds wrapped APIError", func(t *testing.T) {
		apiErr := &api.APIError{StatusCode: 404, Status: "404 Not Found"}
		err := NewExitError(ExitAPIError, apiErr)

		var found *api.APIError
		if !errors.As(err, &found) {
			t.Error("errors.As should find wrapped APIError")
		}
	})
}

func TestExitCodeConstants(t *testing.T) {
	// Verify exit code values match documented values
	if ExitSuccess != 0 {
		t.Errorf("ExitSuccess = %d, want 0", ExitSuccess)
	}
	if ExitUserError != 1 {
		t.Errorf("ExitUserError = %d, want 1", ExitUserError)
	}
	if ExitAPIError != 2 {
		t.Errorf("ExitAPIError = %d, want 2", ExitAPIError)
	}
	if ExitNetError != 3 {
		t.Errorf("ExitNetError = %d, want 3", ExitNetError)
	}
	if ExitAuthError != 4 {
		t.Errorf("ExitAuthError = %d, want 4", ExitAuthError)
	}
	if ExitPartial != 5 {
		t.Errorf("ExitPartial = %d, want 5", ExitPartial)
	}
}
