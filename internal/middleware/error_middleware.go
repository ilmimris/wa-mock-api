package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime/debug"
)

// ErrorResponse is a generic JSON response for errors.
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"` // Optional additional details
}

// RecoveryHandler is a middleware that recovers from panics, logs the error,
// and returns a JSON error response.
func RecoveryHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v\n%s", err, debug.Stack())

				// Prevent further writes to response if headers already sent
				if w.Header().Get("Content-Type") == "" {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.WriteHeader(http.StatusInternalServerError)
				} else {
					// Headers already sent, cannot set status code or content type.
					// Log this situation.
					log.Println("Headers already sent, cannot set JSON error response for panic.")
					return
				}

				response := ErrorResponse{
					Error:   "Internal Server Error",
					Details: "A critical error occurred. Please try again later.",
				}
				if jsonErr := json.NewEncoder(w).Encode(response); jsonErr != nil {
					log.Printf("Failed to write JSON error response: %v", jsonErr)
				}
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// CustomError is an interface that custom error types can implement
// to provide their own HTTP status code and error response.
type CustomError interface {
	error
	StatusCode() int
	JSONResponse() ErrorResponse
}

// ErrorHandlingMiddleware is a middleware that handles errors returned by handlers.
// This version assumes handlers might return a CustomError.
// For simplicity, if handlers don't explicitly return errors in a way this middleware can catch
// (e.g., by writing to response and returning nil), then this middleware primarily relies on RecoveryHandler for panics.
// A more advanced setup might involve a custom http.Handler func signature that returns an error.
func ErrorHandlingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This is a placeholder for where more sophisticated error handling could go
		// if handlers returned errors directly in a way that could be intercepted here
		// before `next.ServeHTTP` or if `next.ServeHTTP` itself returned an error.
		// Since standard http.Handler doesn't return errors, panic recovery is the main mechanism.

		// Example: if using a custom handler that returns `error`:
		// if err := next.(CustomHandlerType).ServeHTTPWithErro(w,r); err != nil {
		//    handleError(w, r, err) // handleError would write JSON response
		//    return
		// }
		next.ServeHTTP(w, r)
	})
}

// Helper function to write error responses (can be used by more specific error handlers too)
func WriteJSONError(w http.ResponseWriter, statusCode int, errResponse ErrorResponse) {
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	}
	w.WriteHeader(statusCode) // Set status code *before* writing body for some clients/frameworks
	if encodeErr := json.NewEncoder(w).Encode(errResponse); encodeErr != nil {
		log.Printf("Error encoding JSON error response: %v", encodeErr)
		// Fallback if encoding fails, though headers might be partially written
		http.Error(w, `{"error":"Failed to encode error response"}`, http.StatusInternalServerError)
	}
}
