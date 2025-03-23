package reportGeneration

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
	"html"
	"os"
	"github.com/chromedp/cdproto/page"

	"github.com/chromedp/chromedp"
	"github.com/dchenk/go-render-quill"
	"sema/models/delta"
)

// GeneratePDF renders the report content as a PDF
func GeneratePDF(reportName string, reportContent []map[string]interface{}) {
	fmt.Println("Generating PDF for:", reportName)

	var htmlContent string

	for _, sectionData := range reportContent {
		sectionName := sectionData["sectionTitle"].(string)
		htmlContent += fmt.Sprintf("<h1>%s</h1>", sectionName)

		subSectionData, ok := sectionData["subsections"].([]map[string]interface{})
		if !ok {
			continue
		}

		for _, subsection := range subSectionData {
			subSectionTitle := subsection["title"].(string)
			htmlContent += fmt.Sprintf("<h2>%s</h2>", subSectionTitle)

			subSectionContent, ok := subsection["content"].(string)
			if !ok {
				continue
			}

			var parsedDelta delta.Delta
			err := json.Unmarshal([]byte(subSectionContent), &parsedDelta)
			if err != nil {
				continue
			}

			opsJSON, err := json.Marshal(parsedDelta.Delta.Delta.Ops)
			if err != nil {
				log.Fatalf("Error marshalling Ops: %v", err)
			}

			html, err := quill.Render(opsJSON)
			if err != nil {
				log.Fatalf("Error converting Delta to HTML: %v", err)
			}

			htmlContent += fmt.Sprintf("<p>%s</p>", html)
		}
	}

	// clean up html

	// Save HTML content to a PDF
	saveHTMLToPDF(reportName+".pdf", htmlContent)
}


func saveHTMLToPDF(pdfPath, htmlContent string) {
	// Create a new context for chromedp
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// HTML template to wrap the raw HTML content
	escapedHTML := html.EscapeString(htmlContent)
	fmt.Println(escapedHTML)

	// Construct the full HTML with escaped content
		var buf []byte
	err := chromedp.Run(ctx,
		// Open a blank page
		chromedp.Navigate("about:blank"),
		// Inject the raw HTML into the page body
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Inject the HTML content safely by escaping it
			err := chromedp.Evaluate(fmt.Sprintf(`document.documentElement.innerHTML = "%s";`, htmlContent), nil).Do(ctx)
			if err != nil {
				return err
			}
			return nil
		}),
		// Wait a bit for content to load before generating the PDF
		chromedp.Sleep(2 * time.Second),
		// Print the page to a PDF and capture the PDF content in 'buf'
		chromedp.ActionFunc(func(ctx context.Context) error {
			// The buf variable must be passed correctly for the PDF content
			var localBuf []byte
			localBuf, _, err := page.PrintToPDF().WithPrintBackground(true).Do(ctx)
			if err != nil {
				return err
			}
			buf = localBuf // Assign the PDF content to the outer buf variable
			return nil
		}),
	)
	if err != nil {
		log.Fatalf("Failed to render PDF: %v", err)
	}

	// Write the PDF to a file
	err = os.WriteFile(pdfPath, buf, 0644)
	if err != nil {
		log.Fatalf("Failed to save PDF: %v", err)
	}

	fmt.Println("PDF successfully generated:", pdfPath)
}
