package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"go-whatsapp-screenshot/internal/services"
	"go-whatsapp-screenshot/internal/utils"
)

// ScreenshotRequest defines the structure for the JSON request body.
type ScreenshotRequest struct {
	Messages          []RequestMessage            `json:"messages"`
	ChatName          string                      `json:"chatName"` // Used as HeaderLineText
	LastSeen          string                      `json:"lastSeen"`
	OutputFileName    string                      `json:"outputFileName"`    // Optional: for Content-Disposition
	ScreenshotOptions *services.ScreenshotOptions `json:"screenshotOptions"` // Optional: to override defaults
}

// RequestMessage is a simplified message structure from the input JSON.
// We will map this to utils.Message.
type RequestMessage struct {
	SessionID      json.Number `json:"session_id,omitempty"` // Using json.Number for flexibility
	Timestamp      string      `json:"timestamp"`
	Sender         string      `json:"sender"` // Maps to Author in utils.Message
	Content        string      `json:"content"`
	AWBNumber      string      `json:"awb_number,omitempty"`
	RecipientName  string      `json:"recipient_name,omitempty"`
	RecipientPhone string      `json:"recipient_phone,omitempty"`
	// We can add a Type field here if the client can specify it,
	// otherwise, we'll infer or default it. For now, assume "message" type.
	// ID can be generated if not provided.
}

// ScreenshotHandler handles requests to generate a screenshot of a chat.
func ScreenshotHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ScreenshotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding JSON request: %v", err)
		http.Error(w, fmt.Sprintf("Invalid JSON payload: %v", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// --- Prepare ChatData for HTML generation ---
	rawChatData := utils.RawChatData{
		ChatName:       req.ChatName, // Will be used as HeaderLineText in the template
		HeaderLineText: req.ChatName,
		LastSeen:       req.LastSeen,
		Messages:       make([]utils.RawMessage, len(req.Messages)),
		// Width can be set from screenshot options or a default
	}

	for i, rm := range req.Messages {
		// Assuming "Bot" sender means message is from "self" (sent, no author displayed in bubble)
		// and other senders are "received" (author displayed).
		// This logic might need adjustment based on how "sender" field is used.
		// If `rm.Sender` is "Bot" or "User" (representing self), Author should be empty for styling.
		author := rm.Sender
		if strings.ToLower(rm.Sender) == "bot" || strings.ToLower(rm.Sender) == "user" { // Example "self" identifiers
			author = "" // Mark as "sent" by the user viewing the chat
		}

		rawChatData.Messages[i] = utils.RawMessage{
			ID:        fmt.Sprintf("msg%d_%s", i, time.Now().Format("150405")), // Generate a simple unique ID
			Author:    author,
			Content:   rm.Content,
			Timestamp: rm.Timestamp,
			Type:      "message", // Default to "message". Could be enhanced if client sends type.
		}
	}

	// --- Apply Screenshot Options ---
	// Use provided options or defaults.
	// The HTML template has a {{width}} placeholder for the body.
	// This should ideally come from screenshot options if available.
	activeScreenshotOptions := services.ScreenshotOptions{
		Width:      1280,                     // Default width
		Height:     720,                      // Default height (less critical if selector/fullpage is used)
		Selector:   services.DefaultSelector, // Default selector
		IsFullPage: false,                    // Default: capture selector, not full page
		Format:     "png",                    // Default format
		Quality:    90,                       // Default JPEG quality
		Timeout:    30 * time.Second,
	}

	if req.ScreenshotOptions != nil {
		if req.ScreenshotOptions.Width > 0 {
			activeScreenshotOptions.Width = req.ScreenshotOptions.Width
		}
		if req.ScreenshotOptions.Height > 0 {
			activeScreenshotOptions.Height = req.ScreenshotOptions.Height
		}
		if req.ScreenshotOptions.Selector != "" {
			activeScreenshotOptions.Selector = req.ScreenshotOptions.Selector
		}
		activeScreenshotOptions.IsFullPage = req.ScreenshotOptions.IsFullPage
		if req.ScreenshotOptions.Format != "" {
			activeScreenshotOptions.Format = strings.ToLower(req.ScreenshotOptions.Format)
		}
		if req.ScreenshotOptions.Quality > 0 && activeScreenshotOptions.Format == "jpeg" {
			activeScreenshotOptions.Quality = req.ScreenshotOptions.Quality
		}
		if req.ScreenshotOptions.Timeout > 0 {
			activeScreenshotOptions.Timeout = req.ScreenshotOptions.Timeout
		}
	}

	// Set the width for the HTML template from the screenshot options
	rawChatData.Width = activeScreenshotOptions.Width

	// Process raw chat data to format messages (bold, italics, etc.)
	processedChatData := utils.ProcessChatData(rawChatData)

	// --- Generate HTML ---
	htmlStr, err := utils.GenerateHTML(processedChatData, "templates/whatsapp-chat.html")
	if err != nil {
		log.Printf("Error generating HTML: %v", err)
		http.Error(w, "Failed to generate HTML content", http.StatusInternalServerError)
		return
	}

	// --- Take Screenshot ---
	screenshotBytes, err := services.TakeScreenshotFromHTML(htmlStr, activeScreenshotOptions)
	if err != nil {
		log.Printf("Error taking screenshot: %v", err)
		http.Error(w, "Failed to take screenshot", http.StatusInternalServerError)
		return
	}

	// --- Return Image ---
	contentType := "image/png"
	if activeScreenshotOptions.Format == "jpeg" {
		contentType = "image/jpeg"
	}
	w.Header().Set("Content-Type", contentType)

	// Optional: Set Content-Disposition to suggest a filename
	if req.OutputFileName != "" {
		// Basic sanitization for filename
		safeFileName := strings.ReplaceAll(req.OutputFileName, "\"", "_")
		safeFileName = strings.ReplaceAll(safeFileName, "/", "_")
		safeFileName = strings.ReplaceAll(safeFileName, "\\", "_")
		if !strings.HasSuffix(strings.ToLower(safeFileName), "."+activeScreenshotOptions.Format) {
			safeFileName += "." + activeScreenshotOptions.Format
		}
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", safeFileName))
	} else {
		defaultFilename := "whatsapp-chat-screenshot." + activeScreenshotOptions.Format
		w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", defaultFilename))
	}

	if _, err := w.Write(screenshotBytes); err != nil {
		log.Printf("Error writing screenshot to response: %v", err)
		// http.Error can't be used here as headers might have been written
	}
	log.Println("Screenshot request processed successfully.")
}
