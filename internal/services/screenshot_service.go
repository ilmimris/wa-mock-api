package services

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// DefaultSelector is the CSS selector for the main chat container.
const DefaultSelector = ".chat-container" // Matches the class in the HTML template
const defaultTimeout = 30 * time.Second   // Default timeout for screenshot operations

// ScreenshotOptions defines configuration for taking a screenshot.
type ScreenshotOptions struct {
	Width      int           // Viewport width
	Height     int           // Viewport height (less relevant for full page or specific element if it dictates size)
	Selector   string        // CSS selector for the element to capture. If empty and not IsFullPage, captures viewport.
	Quality    int           // JPEG quality (1-100). Only used if Format is "jpeg" and IsFullPage is true.
	Format     string        // "jpeg" or "png". Currently, only FullScreenshot explicitly supports JPEG via quality. Others default to PNG.
	IsFullPage bool          // Whether to capture the full scrollable page.
	Timeout    time.Duration // Optional timeout for the operation. Defaults to `defaultTimeout`.
}

// TakeScreenshotFromHTML generates a screenshot from an HTML string using chromedp.
func TakeScreenshotFromHTML(htmlContent string, options ScreenshotOptions) ([]byte, error) {
	// Apply default options
	if options.Width == 0 {
		options.Width = 1280 // Default width
	}
	if options.Height == 0 {
		// For full page or element screenshots, height is often determined by content.
		// For viewport, it's important. Defaulting to a reasonable value.
		options.Height = 720
	}
	// Note: Selector default is handled in the task logic if needed.
	// If options.Selector is empty and not IsFullPage, it becomes a viewport shot.
	// If a default selector like DefaultSelector is desired for "element" mode when options.Selector is empty,
	// it should be set here or before calling this function.
	// For now, empty selector + not IsFullPage = viewport.

	if options.Format == "" {
		options.Format = "png" // Default to PNG
	}
	if options.Quality == 0 && strings.ToLower(options.Format) == "jpeg" {
		options.Quality = 90 // Default JPEG quality
	} else if options.Quality < 1 || options.Quality > 100 {
		if strings.ToLower(options.Format) == "jpeg" { // Only apply quality clamping for JPEG
			log.Printf("Warning: Quality %d is out of range (1-100). Using default 90 for JPEG.", options.Quality)
			options.Quality = 90
		}
	}

	currentTimeout := options.Timeout
	if currentTimeout == 0 {
		currentTimeout = defaultTimeout
	}

	// Create allocator options
	allocOpts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),            // Often required in containerized environments
		chromedp.Flag("disable-dev-shm-usage", true), // Also common in containers
		chromedp.Flag("enable-logging", "stderr"),    // Enable browser logging
		chromedp.Flag("v", "1"),                      // Verbosity level for browser logs
	)

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), allocOpts...)
	defer cancelAlloc()

	// Create a new browser context
	// Add listener for console logs from the browser
	browserCtx, cancelBrowser := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf), chromedp.WithDebugf(log.Printf), chromedp.WithErrorf(log.Printf))
	defer cancelBrowser()

	// Create a timeout context for the entire operation
	ctx, cancelOperation := context.WithTimeout(browserCtx, currentTimeout)
	defer cancelOperation()

	var buf []byte
	tasks := chromedp.Tasks{
		// Navigate to a blank page first, then set content. This is more robust for large HTML.
		chromedp.Navigate("about:blank"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			lctx, lcancel := context.WithTimeout(ctx, 10*time.Second) // Timeout for this specific action
			defer lcancel()
			frameTree, err := page.GetFrameTree().Do(lctx)
			if err != nil {
				return fmt.Errorf("could not get frame tree: %w", err)
			}
			err = page.SetDocumentContent(frameTree.Frame.ID, htmlContent).Do(lctx)
			if err != nil {
				return fmt.Errorf("could not set document content: %w", err)
			}
			log.Println("Document content set successfully.")
			return nil
		}),
		chromedp.EmulateViewport(int64(options.Width), int64(options.Height)),
	}

	finalFormat := "png" // Default actual format

	if options.IsFullPage {
		log.Printf("Capturing full page screenshot (requested format: %s)", options.Format)
		qualityForFull := 0 // This means PNG for FullScreenshot
		if strings.ToLower(options.Format) == "jpeg" {
			qualityForFull = options.Quality
			finalFormat = "jpeg"
		}
		tasks = append(tasks, chromedp.FullScreenshot(&buf, qualityForFull))
	} else if options.Selector != "" {
		log.Printf("Capturing element screenshot (selector: '%s', format: png)", options.Selector)
		// chromedp.Screenshot captures the element as PNG.
		// Wait for the element to be visible and then capture.
		tasks = append(tasks, chromedp.WaitVisible(options.Selector, chromedp.ByQuery))
		tasks = append(tasks, chromedp.ActionFunc(func(ctx context.Context) error {
			log.Printf("Element '%s' is visible.", options.Selector)
			return nil
		}))
		tasks = append(tasks, chromedp.Screenshot(options.Selector, &buf, chromedp.ByQuery))
		finalFormat = "png"
	} else { // Fallback: capture viewport
		log.Printf("Capturing viewport screenshot (format: png)")
		// chromedp.CaptureScreenshot captures the viewport as PNG.
		tasks = append(tasks, chromedp.CaptureScreenshot(&buf))
		finalFormat = "png"
	}

	if err := chromedp.Run(ctx, tasks); err != nil {
		// Check for specific error types, e.g., context deadline exceeded
		if strings.Contains(err.Error(), "context deadline exceeded") {
			return nil, fmt.Errorf("screenshot operation timed out after %s: %w", currentTimeout, err)
		}
		return nil, fmt.Errorf("failed to take screenshot: %w", err)
	}

	if len(buf) == 0 {
		errMsg := "screenshot buffer is empty"
		if options.Selector != "" {
			errMsg = fmt.Sprintf("%s, ensure selector '%s' exists, is visible, and has non-zero dimensions", errMsg, options.Selector)
		}
		return nil, fmt.Errorf(errMsg)
	}

	log.Printf("Screenshot captured successfully, size: %d bytes, actual format: %s", len(buf), finalFormat)
	return buf, nil
}

// Note: For element or viewport screenshots to be in JPEG format with specific quality,
// a more complex chromedp.ActionFunc would be needed to call the underlying
// page.CaptureScreenshot CDP command with specific format and quality parameters.
// Example:
// tasks = append(tasks, chromedp.ActionFunc(func(ctx context.Context) error {
//		var err error
//		*res, err = page.CaptureScreenshot().
//			WithFormat(page.CaptureScreenshotFormatJpeg). // or page.CaptureScreenshotFormatPng
//			WithQuality(int64(options.Quality)). // 0-100, only for JPEG
//			// WithClip(clip) // For specific area or element
//			Do(ctx)
//		return err
//	}))
// This is currently not implemented for element/viewport to keep it simpler for this iteration.
// The `chromedp.FullScreenshot` handles JPEG quality directly.
// The `chromedp.Screenshot` (element) and `chromedp.CaptureScreenshot` (viewport) output PNG by default with simple usage.
