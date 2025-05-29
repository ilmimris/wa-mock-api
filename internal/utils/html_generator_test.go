package utils

import (
	"html/template"
	"os" // Added for TestGenerateHTML
	"path/filepath" // Added for TestGenerateHTML
	"strings"
	"testing"
)

func TestFormatContentHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected template.HTML
	}{
		{"empty", "", ""},
		{"plain text", "Hello world", "Hello world"},
		{"bold", "*bold text*", "<strong>bold text</strong>"},
		{"italic", "_italic text_", "<em>italic text</em>"},
		{"strikethrough", "~strikethrough text~", "<del>strikethrough text</del>"},
		{"monospace", "```monospace text```", "<code>monospace text</code>"},
		{"newline", "line1\nline2", "line1<br>line2"},
		{"combined", "*bold* and _italic_", "<strong>bold</strong> and <em>italic</em>"},
		{"html escaping", "<script>alert('xss')</script>", "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;"},
		{"advanced bold", "This is *bold* text.", "This is <strong>bold</strong> text."},
		{"advanced italic", "This is _italic_ text.", "This is <em>italic</em> text."},
		{"no format at word boundary", "This_is_not_italic", "This_is_not_italic"},
		{"no format at word boundary bold", "This*is*not*bold", "This*is*not*bold"},
		{"mixed content", "Hello *world* _this_ is a ~test~ ```code```\nNew line with <tag>", "Hello <strong>world</strong> <em>this</em> is a <del>test</del> <code>code</code><br>New line with &lt;tag&gt;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatContentHTML(tt.input); got != tt.expected {
				t.Errorf("formatContentHTML(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestProcessChatData(t *testing.T) {
	tests := []struct {
		name     string
		input    RawChatData
		expected ChatData
	}{
		{
			name: "empty messages",
			input: RawChatData{
				ChatName:     "Test Chat",
				Messages:     []RawMessage{},
				Width:        100,
				HeaderLineText: "Test Chat Header",
				LastSeen:     "today",
			},
			expected: ChatData{
				ChatName:     "Test Chat",
				Messages:     []Message{},
				Width:        100,
				HeaderLineText: "Test Chat Header",
				LastSeen:     "today",
			},
		},
		{
			name: "single message with author",
			input: RawChatData{
				ChatName: "Single Author Chat",
				Messages: []RawMessage{
					{ID: "1", Author: "John", Content: "Hello *John*", Timestamp: "10:00", Type: "message"},
				},
                Width: 800, HeaderLineText: "Chat with John", LastSeen: "yesterday",
			},
			expected: ChatData{
				ChatName: "Single Author Chat",
				Messages: []Message{
					{ID: "1", Author: "John", Content: "Hello <strong>John</strong>", Timestamp: "10:00", Type: "message", FormattedContent: "Hello <strong>John</strong>"},
				},
                Width: 800, HeaderLineText: "Chat with John", LastSeen: "yesterday",
			},
		},
		{
			name: "message without author (sent by self)",
			input: RawChatData{
				ChatName: "Self Chat",
				Messages: []RawMessage{
					{ID: "2", Author: "", Content: "_my message_", Timestamp: "10:01", Type: "message"},
				},
                Width: 800, HeaderLineText: "My Notes", LastSeen: "online",
			},
			expected: ChatData{
				ChatName: "Self Chat",
				Messages: []Message{
					{ID: "2", Author: "", Content: "<em>my message</em>", Timestamp: "10:01", Type: "message", FormattedContent: "<em>my message</em>"},
				},
                Width: 800, HeaderLineText: "My Notes", LastSeen: "online",
			},
		},
        {
            name: "system message",
            input: RawChatData{
                ChatName: "System Notifications",
                Messages: []RawMessage{
                    {ID: "3", Content: "User joined", Timestamp: "10:02", Type: "system"},
                },
                Width: 800, HeaderLineText: "System", LastSeen: "N/A",
            },
            expected: ChatData{
                ChatName: "System Notifications",
                Messages: []Message{
                    {ID: "3", Author: "", Content: "User joined", Timestamp: "10:02", Type: "system", FormattedContent: "User joined"},
                },
                Width: 800, HeaderLineText: "System", LastSeen: "N/A",
            },
        },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ProcessChatData(tt.input)
			// Compare field by field for easier debugging
			if got.ChatName != tt.expected.ChatName {
				t.Errorf("ProcessChatData().ChatName = %v, want %v", got.ChatName, tt.expected.ChatName)
			}
			if got.Width != tt.expected.Width {
				t.Errorf("ProcessChatData().Width = %v, want %v", got.Width, tt.expected.Width)
			}
            if got.HeaderLineText != tt.expected.HeaderLineText {
				t.Errorf("ProcessChatData().HeaderLineText = %v, want %v", got.HeaderLineText, tt.expected.HeaderLineText)
			}
            if got.LastSeen != tt.expected.LastSeen {
				t.Errorf("ProcessChatData().LastSeen = %v, want %v", got.LastSeen, tt.expected.LastSeen)
			}
			if len(got.Messages) != len(tt.expected.Messages) {
				t.Fatalf("ProcessChatData().Messages length = %v, want %v", len(got.Messages), len(tt.expected.Messages))
			}
			for i, msg := range got.Messages {
				expMsg := tt.expected.Messages[i]
				if msg.ID != expMsg.ID {
					t.Errorf("Messages[%d].ID = %v, want %v", i, msg.ID, expMsg.ID)
				}
				if msg.Author != expMsg.Author {
					t.Errorf("Messages[%d].Author = %v, want %v", i, msg.Author, expMsg.Author)
				}
				if msg.Content != expMsg.Content {
					t.Errorf("Messages[%d].Content = %v, want %v", i, msg.Content, expMsg.Content)
				}
				if msg.Timestamp != expMsg.Timestamp {
					t.Errorf("Messages[%d].Timestamp = %v, want %v", i, msg.Timestamp, expMsg.Timestamp)
				}
				if msg.Type != expMsg.Type {
					t.Errorf("Messages[%d].Type = %v, want %v", i, msg.Type, expMsg.Type)
				}
                // Note: FormattedContent is not explicitly checked here as it's an internal detail
                // and its correctness is reflected in the `Content` field of `Message` after processing.
			}
		})
	}
}

func TestGenerateHTML(t *testing.T) {
	// Create a dummy template file for testing
	dummyTemplatePath := "test_whatsapp_chat.html"
	dummyTemplateContent := `
<!DOCTYPE html>
<html>
<head><title>{{.ChatName}}</title><style>body{width:{{.Width}}px;}</style></head>
<body>
<h1>{{.HeaderLineText}}</h1>
<p>Last seen: {{.LastSeen}}</p>
<div class="chat-messages">
{{range .Messages}}
    <div class="{{messageClass .}}">
        {{if .Author}}<p class="author">{{.Author}}</p>{{end}}
        <p class="content">{{.Content}}</p>
        <span class="time">{{formatTimestamp .Timestamp}}</span>
    </div>
{{else}}
    <p>No messages.</p>
{{end}}
</div>
</body></html>`
	// Write this dummy template to a file
	// In a real test setup, you might use ioutil.WriteFile or os.Create
	// For simplicity here, we assume this file exists or use a helper to create it.
	// This test depends on the actual template file.
	// To make it more hermetic, we could mock the file reading or use a template string directly.

	// For now, let's test with a simplified ChatData and check for substrings.
	// A more robust test would involve creating the dummy template file.
	// As a shortcut, if the actual `templates/whatsapp-chat.html` is available in the test context,
	// we can use it, but that makes the unit test less isolated.

	// Let's assume `templates/whatsapp-chat.html` is accessible for this test.
	// If not, this test will fail or need adjustment.
	// A better way is to pass template content directly or mock file reads.

	// For this test, we will focus on checking if GenerateHTML runs and produces *some* output
	// containing expected parts, rather than exact HTML matching.
	// The path to template will be relative from where `go test` is run (package directory).
	// So, "../../templates/whatsapp-chat.html" if running from internal/utils.
	// Or, use an absolute path or ensure the test runner provides the correct context.
	// For now, assuming "templates/whatsapp-chat.html" is resolvable from the package dir.
	// This needs to be fixed if tests are run from repo root.
	// Let's assume the `GenerateHTML` function takes a path relative to the `templates` dir.
	// And the `templates` dir is at the root of the project.
	// The `GenerateHTML` uses `filepath.Abs`, so it should resolve correctly if the `templates` dir is present.
	// For robustness, tests should create their own temporary template files.

	// Creating a temporary template file for this test:
	tempDir := t.TempDir()
	tempTemplateFile, err := os.Create(filepath.Join(tempDir, "temp_chat.html"))
	if err != nil {
		t.Fatalf("Failed to create temp template file: %v", err)
	}
	_, err = tempTemplateFile.WriteString(dummyTemplateContent)
	if err != nil {
		t.Fatalf("Failed to write to temp template file: %v", err)
	}
	tempTemplateFile.Close()
	testTemplatePath := tempTemplateFile.Name()


	chatData := ChatData{
		ChatName:     "My Test Chat",
		HeaderLineText: "Test Chat Header",
		LastSeen:     "just now",
		Width:        500,
		Messages: []Message{
			{ID: "1", Author: "Alice", Content: template.HTML("Hello <b>Alice</b>"), Timestamp: "10:00", Type: "message"},
			{ID: "2", Author: "", Content: template.HTML("My reply"), Timestamp: "10:01", Type: "message"},
			{ID: "3", Author: "", Content: template.HTML("A system message"), Timestamp: "10:02", Type: "system"},
		},
	}

	htmlOutput, err := GenerateHTML(chatData, testTemplatePath)
	if err != nil {
		t.Fatalf("GenerateHTML() error = %v", err)
	}

	if !strings.Contains(htmlOutput, "<title>My Test Chat</title>") {
		t.Errorf("GenerateHTML() output does not contain expected ChatName in title")
	}
    if !strings.Contains(htmlOutput, "width:500px;") {
		t.Errorf("GenerateHTML() output does not contain expected Width")
	}
	if !strings.Contains(htmlOutput, "<h1>Test Chat Header</h1>") {
		t.Errorf("GenerateHTML() output does not contain expected HeaderLineText")
	}
	if !strings.Contains(htmlOutput, "<p>Last seen: just now</p>") {
		t.Errorf("GenerateHTML() output does not contain expected LastSeen")
	}
	if !strings.Contains(htmlOutput, "Hello <b>Alice</b>") {
		t.Errorf("GenerateHTML() output does not contain message content for Alice")
	}
	if !strings.Contains(htmlOutput, "My reply") {
		t.Errorf("GenerateHTML() output does not contain message content for self reply")
	}
    if !strings.Contains(htmlOutput, "message received") { // Alice's message
		t.Errorf("GenerateHTML() output does not contain expected class for received message")
	}
    if !strings.Contains(htmlOutput, "message sent") { // Self reply
		t.Errorf("GenerateHTML() output does not contain expected class for sent message")
	}
    if !strings.Contains(htmlOutput, "message system-message") { // System message
		t.Errorf("GenerateHTML() output does not contain expected class for system message")
	}
    if !strings.Contains(htmlOutput, "<span class=\"time\">10:00</span>") {
		t.Errorf("GenerateHTML() output does not contain formatted timestamp")
	}

	// Test with no messages
	chatDataNoMessages := ChatData{
		ChatName: "Empty Chat",
        HeaderLineText: "Empty Chat Header",
		Messages: []Message{},
	}
	htmlOutputNoMessages, err := GenerateHTML(chatDataNoMessages, testTemplatePath)
	if err != nil {
		t.Fatalf("GenerateHTML() with no messages error = %v", err)
	}
	if !strings.Contains(htmlOutputNoMessages, "<p>No messages.</p>") {
		t.Errorf("GenerateHTML() output for no messages does not contain 'No messages.' text")
	}
}

// Helper to create dummy template file - not used above, but could be an alternative
// For GenerateHTML, we need to import "os" and "path/filepath" if using temporary files.
// Already did that for the test.

// Add tests for helper functions like formatTimestamp, messageClass if they have complex logic not covered.
// formatTimestamp is already used by GenerateHTML test indirectly.
// messageClass is also used indirectly.

func TestFormatTimestamp(t *testing.T) {
    tests := []struct {
        name     string
        timestamp string
        expected string
    }{
        {"already formatted", "10:30", "10:30"},
        {"RFC3339", "2023-10-27T15:04:05Z", "15:04"},
        {"custom format D/M/YYYY, HH:MM", "27/10/2023, 16:05", "16:05"},
        {"custom format M/D/YYYY, HH:MM", "10/27/2023, 17:06", "17:06"},
        {"SQL timestamp", "2023-10-27 18:07:00", "18:07"},
        {"invalid format", "invalid-timestamp", "invalid-timestamp"},
		{"empty", "", ""},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := formatTimestamp(tt.timestamp); got != tt.expected {
                t.Errorf("formatTimestamp(%q) = %q, want %q", tt.timestamp, got, tt.expected)
            }
        })
    }
}

func TestMessageClass(t *testing.T) {
    tests := []struct {
        name     string
        message  Message
        expected string
    }{
        {"sent text message", Message{Author: "", Type: "message"}, "message sent"},
        {"received text message", Message{Author: "John", Type: "message"}, "message received"},
        {"system message", Message{Type: "system"}, "message system-message"},
        {"sent image message", Message{Author: "", Type: "image"}, "message sent image-message"},
        {"received video message", Message{Author: "Jane", Type: "video"}, "message received video-message"},
		{"sent document message", Message{Author: "", Type: "document"}, "message sent document-message"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := messageClass(tt.message); got != tt.expected {
                t.Errorf("messageClass() for %v = %q, want %q", tt.message, got, tt.expected)
            }
        })
    }
}
```

I'll need to import `os` and `path/filepath` for the `TestGenerateHTML` function to correctly create and use a temporary template file. I will add these imports.
