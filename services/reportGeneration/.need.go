package reportGeneration

import (
	"context"
	"encoding/json"
	"fmt"
	
	"log"
	"time"
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
	
	htmlContent += 
	`data:text/html, <!DOCTYPE html>
	 <html>	
		<head> 
			<style> 
				@page { size : A4; margin: 20mm; } 
							</style> 
		</head>
		<body>
		`
		/*
		.page-break { page-break-before: always; } 
				.avoid-break { page-break-inside: avoid; }
				*/

			
		

	for _, sectionData := range reportContent {
		sectionName := sectionData["sectionTitle"].(string)
		// <div class = "page-break">
		htmlSection := ` 
	<div>
				<h1>%s</h1>
			</div>
		`
		htmlContent += fmt.Sprintf(htmlSection, sectionName)

		subSectionData, ok := sectionData["subsections"].([]map[string]interface{})
		if !ok {
			continue
		}

		for _, subsection := range subSectionData {
			subSectionTitle := subsection["title"].(string)

				//<div class = "avoid-break"> 
			htmlSubsection := `
				<div>
				<h2>%s</h2>

			`
			htmlContent += fmt.Sprintf(htmlSubsection, subSectionTitle)

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

			htmlContent += fmt.Sprintf("%s </div>", html)
		}
	}

	htmlContent += `</body></html>`
	fmt.Println(htmlContent)

	// Save HTML content to a PDF
	saveHTMLToPDF(reportName+".pdf", htmlContent)
}

    

func saveHTMLToPDF(pdfPath, htmlContent string) {
    // Create a new context for chromedp
    ctx, cancel := chromedp.NewContext(context.Background())
    defer cancel()

    // HTML content should not be escaped
    var buf []byte
    err := chromedp.Run(ctx,
        // Open a blank page
				// chromedp.Navigate("data:text/html," + url.PathEscape(htmlContent)),
				


				// Inject the raw HTML into the page body
				chromedp.ActionFunc(func(ctx context.Context) error {
					// Inject raw HTML content without escaping it
					err := chromedp.Evaluate(fmt.Sprintf(`document.documentElement.innerHTML = '%s';`, htmlContent), nil).Do(ctx)
					if err != nil {
						return err
					}
					return nil
				}),
				// Wait for the page content to load properly, ensure images load
				chromedp.WaitVisible("body", chromedp.ByQuery), // Wait until body is visible
				chromedp.Sleep(5 * time.Second), // Additional wait for complex resources like images

				// Print the page to a PDF and capture the PDF content in 'buf'
				chromedp.ActionFunc(func(ctx context.Context) error {
					var localBuf []byte
					localBuf, _, err := page.PrintToPDF().WithPrintBackground(true).Do(ctx)
					if err != nil {
						return err
					}
					buf = localBuf
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

