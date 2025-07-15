package utils

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"log"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Message represents a single chat message.
type Message struct {
	ID        string        `json:"id"`
	Author    string        `json:"author,omitempty"`
	Content   template.HTML `json:"content"` // Changed to template.HTML
	Timestamp string        `json:"timestamp"`
	Type      string        `json:"type"` // "message", "system", "image", "video", "audio", "sticker", "contact", "document"
	MediaURL  string        `json:"mediaUrl,omitempty"`
	FileName  string        `json:"fileName,omitempty"`
	FileSize  string        `json:"fileSize,omitempty"`
	// Fields derived or used by template funcs, not directly from input JSON for content formatting
	FormattedContent template.HTML `json:"-"` // Content after WhatsApp formatting
}

// ChatData represents the overall chat data for the template.
type ChatData struct {
	ChatName       string    `json:"chatName"` // Retained for potential use, maps to HeaderLineText
	Messages       []Message `json:"messages"`
	Width          int       `json:"width"`          // For body style
	HeaderLineText string    `json:"headerLineText"` // For chat header
	LastSeen       string    `json:"lastSeen"`       // For chat header
}

// RawMessage is used for decoding the input JSON where content is still a string.
type RawMessage struct {
	ID        string `json:"id"`
	Author    string `json:"author,omitempty"`
	Content   string `json:"content"` // Content as string
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"`
	MediaURL  string `json:"mediaUrl,omitempty"`
	FileName  string `json:"fileName,omitempty"`
	FileSize  string `json:"fileSize,omitempty"`
}

// RawChatData is used for decoding the input JSON.
type RawChatData struct {
	ChatName       string       `json:"chatName"`
	Messages       []RawMessage `json:"messages"`
	Width          int          `json:"width"`
	HeaderLineText string       `json:"headerLineText"`
	LastSeen       string       `json:"lastSeen"`
}

// formatContentHTML converts WhatsApp style text to HTML.
// It replicates the logic from the JavaScript function convertWhatsAppToHTMLAdvanced.
func formatContentHTML(text string) template.HTML {
	// 1. Escape HTML characters
	escapedText := text
	escapedText = strings.ReplaceAll(escapedText, "&", "&amp;")
	escapedText = strings.ReplaceAll(escapedText, "<", "&lt;")
	escapedText = strings.ReplaceAll(escapedText, ">", "&gt;")
	escapedText = strings.ReplaceAll(escapedText, "\"", "&quot;")
	escapedText = strings.ReplaceAll(escapedText, "'", "&#39;")

	// 2. Apply formatting
	// Bold: *text*
	reBold := regexp.MustCompile(`(^|[^\w])\*([^\s*][^*]*[^\s*]|\S)\*([^\w]|$)`)
	html := reBold.ReplaceAllString(escapedText, "$1<strong>$2</strong>$3")

	// Italic: _text_
	reItalic := regexp.MustCompile(`(^|[^\w])_([^\s_][^_]*[^\s_]|\S)_([^\w]|$)`)
	html = reItalic.ReplaceAllString(html, "$1<em>$2</em>$3")

	// Monospace: ```text```
	reMonospace := regexp.MustCompile("```([^`]+)```")
	html = reMonospace.ReplaceAllString(html, "<code>$1</code>")

	// Strikethrough: ~text~
	reStrikethrough := regexp.MustCompile(`(^|[^\w])~([^\s~][^~]*[^\s~]|\S)~([^\w]|$)`)
	html = reStrikethrough.ReplaceAllString(html, "$1<del>$2</del>$3")

	// Line breaks
	html = strings.ReplaceAll(html, "\n", "<br>")

	return template.HTML(html)
}

// ProcessChatData converts RawChatData to ChatData, including content formatting.
func ProcessChatData(rawData RawChatData) ChatData {
	processedMessages := make([]Message, len(rawData.Messages))
	for i, rawMsg := range rawData.Messages {
		processedMessages[i] = Message{
			ID:               rawMsg.ID,
			Author:           rawMsg.Author,
			Content:          formatContentHTML(rawMsg.Content), // Format content here
			Timestamp:        rawMsg.Timestamp,
			Type:             rawMsg.Type,
			MediaURL:         rawMsg.MediaURL,
			FileName:         rawMsg.FileName,
			FileSize:         rawMsg.FileSize,
			FormattedContent: formatContentHTML(rawMsg.Content), // Also store it here if needed separately
		}
	}
	return ChatData{
		ChatName:       rawData.ChatName, // Or rawData.HeaderLineText if that's the primary source
		Messages:       processedMessages,
		Width:          rawData.Width,
		HeaderLineText: rawData.HeaderLineText,
		LastSeen:       rawData.LastSeen,
	}
}

// GenerateHTML generates HTML from processed chat data using a template.
func GenerateHTML(processedData ChatData, templatePath string) (string, error) {
	absTemplatePath, err := filepath.Abs(templatePath)
	if err != nil {
		log.Printf("Error getting absolute path for template: %v", err)
		return "", err
	}
	log.Printf("Attempting to read template file from: %s", absTemplatePath)

	tmplContent, err := ioutil.ReadFile(absTemplatePath)
	if err != nil {
		log.Printf("Error reading template file %s: %v", absTemplatePath, err)
		return "", err
	}

	// Create a new template and parse the template content
	// The actual message content formatting is now done in ProcessChatData.
	// The template will directly use the .Content field which is already template.HTML
	tmpl, err := template.New(filepath.Base(absTemplatePath)).Funcs(template.FuncMap{
		"formatTimestamp": formatTimestamp,
		"isSystemMessage": isSystemMessage,
		"isMediaMessage":  isMediaMessage,
		"isTextMessage":   isTextMessage,
		"isImage":         func(m Message) bool { return m.Type == "image" },
		"isVideo":         func(m Message) bool { return m.Type == "video" },
		"isAudio":         func(m Message) bool { return m.Type == "audio" },
		"isSticker":       func(m Message) bool { return m.Type == "sticker" },
		"isContact":       func(m Message) bool { return m.Type == "contact" },
		"isDocument":      func(m Message) bool { return m.Type == "document" },
		"hasAuthor":       func(m Message) bool { return m.Author != "" && m.Type == "message" },
		"messageClass":    messageClass,
		"mediaIconClass":  mediaIconClass,
		// No need for a content formatting func here if Message.Content is pre-formatted to template.HTML
	}).Parse(string(tmplContent))
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, processedData); err != nil {
		log.Printf("Error executing template: %v", err)
		return "", err
	}

	return buf.String(), nil
}

func formatTimestamp(timestamp string) string {
	if strings.Contains(timestamp, ":") && !strings.Contains(timestamp, " ") && !strings.Contains(timestamp, "T") {
		return timestamp // Already HH:MM
	}
	layouts := []string{
		time.RFC3339,              // "2006-01-02T15:04:05Z07:00"
		"2/1/2006, 15:04",         // "D/M/YYYY, HH:MM"
		"1/2/2006, 15:04",         // "M/D/YYYY, HH:MM"
		"2006-01-02 15:04:05",     // Common SQL timestamp
		"2006-01-02T15:04:05.000", // Timestamp with milliseconds
	}
	for _, layout := range layouts {
		t, err := time.Parse(layout, timestamp)
		if err == nil {
			return t.Format("15:04") // HH:MM
		}
	}
	log.Printf("Could not parse timestamp: %s, returning as is.", timestamp)
	return timestamp
}

func isSystemMessage(msg Message) bool {
	return msg.Type == "system"
}

func isMediaMessage(msg Message) bool {
	// Broader definition including documents
	return msg.Type == "image" || msg.Type == "video" || msg.Type == "audio" || msg.Type == "sticker" || msg.Type == "document"
}

func isTextMessage(msg Message) bool {
	// A message is text if it's of type "message" and not any known media type.
	// This relies on MediaURL/FileName being empty for pure text messages.
	return msg.Type == "message" && msg.MediaURL == "" && msg.FileName == "" && !isMediaMessage(msg)
}

func messageClass(msg Message) string {
	baseClass := "message"
	// Determine if the message is sent or received.
	// The original template CSS implies "sent" messages don't explicitly show author in the bubble,
	// but are right-aligned. "received" messages are left-aligned and may show author.
	// Let's assume: if Author is empty OR Author is a special value indicating "self", it's sent.
	// This logic might need adjustment based on actual data.
	// For now, if Author is empty, it's 'sent'. If Author is present, it's 'received'.
	// System messages are distinct.
	if msg.Type == "system" {
		return "message system-message"
	}

	if msg.Author == "" { // Assuming no author means it's a "sent" message by the user
		baseClass += " sent"
	} else {
		baseClass += " received"
	}

	// Append type-specific classes
	if msg.Type == "image" {
		return baseClass + " image-message"
	}
	if msg.Type == "video" {
		return baseClass + " video-message"
	}
	if msg.Type == "audio" {
		return baseClass + " audio-message"
	}
	if msg.Type == "sticker" {
		return baseClass + " sticker-message"
	}
	if msg.Type == "contact" {
		return baseClass + " contact-message"
	}
	if msg.Type == "document" {
		return baseClass + " document-message"
	}
	// If it's a plain text message (type "message" without specific media)
	if msg.Type == "message" && !isMediaMessage(msg) {
		// no special class other than .message .sent or .message .received
	}
	return baseClass
}

func mediaIconClass(msg Message) string {
	// These are just placeholder class names. Actual icons would need CSS/SVGs.
	switch msg.Type {
	case "audio":
		return "icon-audio" // Example class
	case "document":
		return "icon-document" // Example class
	case "video":
		return "icon-video" // Example class
	default:
		return "" // No specific icon class for images, stickers, text
	}
}
