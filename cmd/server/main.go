package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"go-whatsapp-screenshot/internal/handlers"
	"go-whatsapp-screenshot/internal/middleware"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port
	}

	// Create a new ServeMux
	mux := http.NewServeMux()

	// Wrap the screenshot handler with the recovery middleware
	screenshotHandlerWithMiddleware := middleware.RecoveryHandler(http.HandlerFunc(handlers.ScreenshotHandler))
	mux.Handle("/screenshot", screenshotHandlerWithMiddleware)

	// Health check endpoint (not wrapped by recovery handler for simplicity, or could be)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})

	log.Printf("Server starting on port %s\n", port)
	
	// Use the mux with http.ListenAndServe
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Error starting server: %s\n", err)
	}
}

// screenshotHandler is now defined in internal/handlers/screenshot_handler.go
// We keep this file clean by delegating handler logic.
// func screenshotHandler(w http.ResponseWriter, r *http.Request) {
// 	log.Printf("Received request for %s %s\n", r.Method, r.URL.Path)
// 	if r.Method != http.MethodPost {
// 		http.Error(w, "Only POST requests are allowed", http.StatusMethodNotAllowed)
// 		return
// 	}
// 	// Dummy success message
// 	fmt.Fprintf(w, "Screenshot request received")
// }
}
