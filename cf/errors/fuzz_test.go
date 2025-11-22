// +build gofuzz

package errors

import (
	"testing"
)

// FuzzNew tests the New function with random input
func FuzzNew(f *testing.F) {
	// Seed corpus with interesting inputs
	f.Add("")
	f.Add("simple error")
	f.Add("error with special chars: !@#$%^&*()")
	f.Add("unicode: 你好世界 שלום עולם")
	f.Add("very long error: " + string(make([]byte, 10000)))
	f.Add("\n\t\r\x00")
	f.Add("SQL injection attempt: '; DROP TABLE users--")
	f.Add("<script>alert('xss')</script>")

	f.Fuzz(func(t *testing.T, msg string) {
		// Should never panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("New panicked with input %q: %v", msg, r)
			}
		}()

		err := New(msg)

		// Verify error is created
		if err == nil {
			t.Errorf("New(%q) returned nil", msg)
		}

		// Verify error message is preserved
		if err != nil && err.Error() != msg {
			t.Errorf("New(%q).Error() = %q, want %q", msg, err.Error(), msg)
		}
	})
}

// FuzzHttpError tests HttpError creation with random status codes and messages
func FuzzHttpError(f *testing.F) {
	// Seed corpus
	f.Add(404, "NOT_FOUND", "Resource not found")
	f.Add(500, "SERVER_ERROR", "Internal server error")
	f.Add(0, "", "")
	f.Add(-1, "NEGATIVE", "negative status")
	f.Add(999999, "HUGE", "huge status code")

	f.Fuzz(func(t *testing.T, statusCode int, code, description string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("NewHttpError panicked: %v", r)
			}
		}()

		err := NewHttpError(statusCode, code, description)

		if err == nil {
			return // OK to return nil for some inputs
		}

		// If error is created, it should implement error interface
		_ = err.Error()

		// Check if it's an HttpError
		if httpErr, ok := err.(HttpError); ok {
			// Verify status code is preserved
			if httpErr.StatusCode() != statusCode {
				t.Errorf("StatusCode() = %d, want %d", httpErr.StatusCode(), statusCode)
			}

			// Verify error code is preserved
			if httpErr.ErrorCode() != code {
				t.Errorf("ErrorCode() = %q, want %q", httpErr.ErrorCode(), code)
			}
		}
	})
}

// FuzzNewWithSlice tests error aggregation with random error slices
func FuzzNewWithSlice(f *testing.F) {
	f.Add(3) // Number of errors to create

	f.Fuzz(func(t *testing.T, numErrors int) {
		// Limit to reasonable range
		if numErrors < 0 || numErrors > 100 {
			return
		}

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("NewWithSlice panicked: %v", r)
			}
		}()

		// Create random error slice
		var errs []error
		for i := 0; i < numErrors; i++ {
			errs = append(errs, New("error"))
		}

		result := NewWithSlice(errs)

		// Should return nil for empty slice
		if len(errs) == 0 && result != nil {
			t.Errorf("NewWithSlice([]) should return nil")
		}

		// Should return error for non-empty slice
		if len(errs) > 0 && result == nil {
			t.Errorf("NewWithSlice(errs) returned nil for %d errors", numErrors)
		}
	})
}
