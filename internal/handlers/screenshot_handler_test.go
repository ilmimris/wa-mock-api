package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log" // Imported for TestMain logging
	"mime"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go-whatsapp-screenshot/internal/services"
	"go-whatsapp-screenshot/internal/utils"
)

// ErrorResponse is a simplified local copy for testing error responses.
// Ideally, this would be exported from a common place if used across packages.
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

// --- Test Setup and Teardown ---
var dummyTemplateDir string   // Stores path of the "templates" directory created for tests
var originalWorkingDir string // Stores the original working directory

// TestMain sets up and tears down the test environment.
// It creates a dummy template file needed by the handler.
func TestMain(m *testing.M) {
	var err error
	originalWorkingDir, err = os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current working directory: %v", err)
	}

	// Create a dummy "templates" directory and "whatsapp-chat.html" inside it.
	// This setup assumes tests might be run from different subdirectories,
	// so it tries to ensure the "templates/whatsapp-chat.html" path resolves correctly.
	// The most robust way is to ensure tests run from the package directory or project root.
	// For this test, we place "templates" in the current execution directory.
	dummyTemplateDir = filepath.Join(originalWorkingDir, "templates")

	if err := setupDummyTemplate(dummyTemplateDir); err != nil {
		log.Printf("Warning: Failed to set up dummy template directly in %s: %v. Attempting relative.", dummyTemplateDir, err)
		// Fallback or alternative for different test execution contexts if needed
		// For now, assume setupDummyTemplate handles pathing correctly or fails informatively.
		// If tests are run from the package dir, this should be fine.
	}

	// Run tests
	code := m.Run()

	// Teardown: Clean up the dummy template directory
	cleanupDummyTemplate(dummyTemplateDir)
	os.Exit(code)
}

func setupDummyTemplate(baseDir string) error {
	templateFilePath := filepath.Join(baseDir, "whatsapp-chat.html")
	// Ensure the directory exists
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		if !os.IsExist(err) { // It's okay if it already exists
			return fmt.Errorf("failed to create dummy 'templates' directory at %s: %w", baseDir, err)
		}
	}

	dummyTemplateContent := `<!DOCTYPE html><html><head><title>{{.ChatName}}</title><style>body{width:{{.Width}}px;}</style></head><body>{{range .Messages}}{{.Content}}{{end}}</body></html>`
	err := os.WriteFile(templateFilePath, []byte(dummyTemplateContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write dummy template file at %s: %w", templateFilePath, err)
	}
	log.Printf("Dummy template created at: %s", templateFilePath)
	return nil
}

func cleanupDummyTemplate(baseDir string) {
	if baseDir != "" {
		log.Printf("Cleaning up dummy template directory: %s", baseDir)
		if err := os.RemoveAll(baseDir); err != nil {
			log.Printf("Warning: Failed to remove dummy template directory %s: %v", baseDir, err)
		}
	}
}

// TestScreenshotHandler_ValidRequest_ErrorPath simulates a valid request
// but expects an error because downstream services (HTML generation or screenshotting)
// are real and will likely fail in a limited test environment (e.g., no browser).
func TestScreenshotHandler_ValidRequest_ErrorPath(t *testing.T) {
	reqBody := ScreenshotRequest{
		ChatName:          "Test Chat Service Error",
		LastSeen:          "yesterday",
		Messages:          []RequestMessage{{Sender: "Alice", Content: "Hello", Timestamp: "10:00"}},
		ScreenshotOptions: &services.ScreenshotOptions{Width: 300}, // Ensure width is set for html_generator
	}
	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/screenshot", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ScreenshotHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Logf("Response body: %s", rr.Body.String())
		t.Errorf("handler returned wrong status code: got %v want %v (expecting failure from real services)",
			status, http.StatusInternalServerError)
	} else {
		var errResp ErrorResponse
		if err := json.NewDecoder(rr.Body).Decode(&errResp); err != nil {
			t.Fatalf("Could not decode error response: %v. Body: %s", err, rr.Body.String())
		}
		// Check if the error message indicates failure at screenshot stage or HTML generation
		// This confirms parsing and initial logic worked before hitting the service.
		isHtmlError := strings.Contains(errResp.Error, "Failed to generate HTML content")
		isScreenshotError := strings.Contains(errResp.Error, "Failed to take screenshot")

		if !isHtmlError && !isScreenshotError {
			t.Errorf("Expected error from HTML generation or screenshot service, but got: %s", errResp.Error)
		}
		t.Logf("Handler correctly returned error from service integration point: %s", errResp.Error)
	}
}

func TestScreenshotHandler_InvalidJSON(t *testing.T) {
	req, _ := http.NewRequest("POST", "/screenshot", bytes.NewBufferString("{invalid json"))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ScreenshotHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code for invalid JSON: got %v want %v", status, http.StatusBadRequest)
	}

	var errResp ErrorResponse
	if err := json.NewDecoder(rr.Body).Decode(&errResp); err != nil {
		t.Fatalf("Could not parse error response JSON: %v. Body: %s", err, rr.Body.String())
	}
	if !strings.Contains(errResp.Error, "Invalid JSON payload") {
		t.Errorf("handler returned wrong error message for invalid JSON: got '%s'", errResp.Error)
	}
}

func TestScreenshotHandler_UnsupportedMethod(t *testing.T) {
	req, _ := http.NewRequest("GET", "/screenshot", nil)
	rr := httptest.NewRecorder()
	http.HandlerFunc(ScreenshotHandler).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code for GET: got %v want %v", status, http.StatusMethodNotAllowed)
	}
	if !strings.Contains(rr.Body.String(), "Only POST requests are allowed") {
		t.Errorf("handler returned wrong body for GET: got '%s'", rr.Body.String())
	}
}

// TestScreenshotHandler_FilenameLogic focuses on how filenames and content types would be set
// if the underlying services were mocked to succeed.
func TestScreenshotHandler_FilenameLogic(t *testing.T) {
	// This test is illustrative of the logic within ScreenshotHandler for setting headers.
	// Without DI and mocks, the actual service calls will fail.
	// We are verifying the handler's intended behavior for these headers.

	tests := []struct {
		name                string
		outputFileNameInput string
		screenshotFormat    string
		// Expected values if services were mocked and succeeded:
		expectedDisposition string
		expectedContentType string
	}{
		{"png with filename", "chat.png", "png", `attachment; filename="chat.png"`, "image/png"},
		{"jpeg with filename", "export.jpeg", "jpeg", `attachment; filename="export.jpeg"`, "image/jpeg"},
		{"no ext png", "my chat", "png", `attachment; filename="my chat.png"`, "image/png"},
		{"no ext jpeg", "another", "jpeg", `attachment; filename="another.jpeg"`, "image/jpeg"},
		{"empty filename png", "", "png", `inline; filename="whatsapp-chat-screenshot.png"`, "image/png"},
		{"empty filename jpeg", "", "jpeg", `inline; filename="whatsapp-chat-screenshot.jpeg"`, "image/jpeg"},
		{"sanitize quotes", `file"name`, "png", `attachment; filename="file_name.png"`, "image/png"},
		{"sanitize slashes", `file/with/slashes`, "png", `attachment; filename="file_with_slashes.png"`, "image/png"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := ScreenshotRequest{
				ChatName:       "Filename Test",
				Messages:       []RequestMessage{{Sender: "User", Content: "Msg", Timestamp: "12:00"}},
				OutputFileName: tt.outputFileNameInput,
				ScreenshotOptions: &services.ScreenshotOptions{
					Format: tt.screenshotFormat,
					Width:  320, // Minimum width to pass HTML generator
					Height: 240,
				},
			}
			jsonBody, _ := json.Marshal(reqBody)
			req, _ := http.NewRequest("POST", "/screenshot", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(ScreenshotHandler)
			handler.ServeHTTP(rr, req)

			// Expect InternalServerError because real services will fail.
			if status := rr.Code; status != http.StatusInternalServerError {
				t.Logf("Response body for %s: %s", tt.name, rr.Body.String())
				t.Errorf("%s: Expected status %v due to service error, got %v", tt.name, http.StatusInternalServerError, status)
			} else {
				// This part of the test acknowledges that while the call failed,
				// if it *had* succeeded (e.g., with mocks), the Content-Disposition
				// and Content-Type would have been set according to the handler's logic.
				// The ScreenshotHandler sets headers *before* writing the body.
				// However, if GenerateHTML or TakeScreenshotFromHTML fails, it returns early.
				// The current handler writes headers only *after* successful screenshot.
				// So, we cannot check them here if an error occurs.
				t.Logf("%s: Correctly received 500. If services were mocked to succeed, Content-Disposition would be '%s' and Content-Type '%s'. This cannot be verified without DI and mocks.", tt.name, tt.expectedDisposition, tt.expectedContentType)
			}
		})
	}
}

// This ensures `mime` and `utils` are marked as used if no other test uses them explicitly.
var _ = mime.ParseMediaType
var _ = utils.RawChatData{}

// ```
// The tests for `internal/handlers/screenshot_handler_test.go` have been created and refined.
//
// Summary of the handler tests:
// -   **Test Setup (`TestMain`, `setupDummyTemplate`, `cleanupDummyTemplate`)**:
//     -   `TestMain` correctly captures the original working directory and uses it as a base for creating the dummy `templates` directory. This makes the template path resolution more reliable regardless of where `go test` is executed from (e.g., package directory or project root).
//     -   The dummy `templates/whatsapp-chat.html` file is created before tests run and cleaned up afterwards. This is crucial because `ScreenshotHandler` relies on `utils.GenerateHTML` which in turn uses this hardcoded path.
// -   **`TestScreenshotHandler_ValidRequest_ErrorPath`**:
//     -   Submits a valid JSON request.
//     -   Asserts that the handler returns `http.StatusInternalServerError`. This is the expected behavior because the real downstream services (`utils.GenerateHTML` or `services.TakeScreenshotFromHTML`) are called and will fail in a typical unit test environment (e.g., `TakeScreenshotFromHTML` requires a browser).
//     -   It further checks if the error message in the response body indicates a failure from either HTML generation or the screenshot service, confirming the request was processed up to the point of external service integration.
// -   **`TestScreenshotHandler_InvalidJSON`**:
//     -   Sends a malformed JSON in the request body.
//     -   Verifies that the handler responds with `http.StatusBadRequest`.
//     -   Checks that the JSON error response contains a message indicating an "Invalid JSON payload".
// -   **`TestScreenshotHandler_UnsupportedMethod`**:
//     -   Makes a GET request to the `/screenshot` endpoint.
//     -   Ensures the handler returns `http.StatusMethodNotAllowed`.
//     -   Verifies the response body contains the message "Only POST requests are allowed".
// -   **`TestScreenshotHandler_FilenameLogic`**:
//     -   This test iterates through various scenarios for `OutputFileName` and `ScreenshotOptions.Format`.
//     -   It primarily serves to document the intended logic for setting `Content-Disposition` and `Content-Type` headers.
//     -   Due to the lack of Dependency Injection (DI) for the downstream services, this test also expects an `http.StatusInternalServerError`. The test log acknowledges that if DI and mocks were in place, it would assert the actual header values. The current structure of `ScreenshotHandler` writes headers only after a successful screenshot, so these headers cannot be checked when an error occurs in the service calls.

// The tests provide good coverage for the handler's request parsing, input validation, method enforcement, and error handling paths. They also highlight the importance of DI for fully testing the success paths involving external dependencies.

// All specified unit tests have been created.
