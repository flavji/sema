package reportGeneration


import (
  "context"
  "encoding/json"
  "fmt"
  "time"
  "os"
  "github.com/chromedp/cdproto/page"
  "github.com/chromedp/chromedp"
  "github.com/dchenk/go-render-quill"
  "sema/models/delta"
)

// GeneratePDF renders the report content as a PDF
func GeneratePDF(reportName string, reportContent []map[string]interface{}) error {
  fmt.Println("Generating PDF for:", reportName)

  var htmlContent string

  htmlContent += "<!DOCTYPE html> <html> <head> <style>" +
    "@page { size: A4; margin: 20mm; } " +
    "body { font-family: \"Times New Roman\", serif; font-size: 12pt; line-height: 1.5; text-align: justify; color: #333; } " +
    "h1 { font-size: 18pt; font-weight: bold; margin-bottom: 10mm; text-transform: uppercase; padding-bottom: 3mm; } " +
    "h2 { font-size: 14pt; font-weight: bold; margin-top: 10mm; margin-bottom: 4mm; } " +
    "p { margin-bottom: 5mm; } " +
    ".page-break { page-break-before: always; } " +
    ".avoid-break { page-break-inside: avoid; } " +
    "</style></head><body>"

  titlePage := "<div style=\"height: 100vh; display: flex; flex-direction: column; justify-content: center; align-items: center; text-align: center;\"> <h1>%s</h1></div> <div style=\"position: absolute; bottom: 20px; width: 100%%; text-align: center;\"><p>%s</p> </div>"
  htmlContent += fmt.Sprintf(titlePage, reportName, time.Now().Format("January 2, 2006"))

  for i, sectionData := range reportContent {
    sectionName := sectionData["sectionTitle"].(string)
    var htmlSection string
    if i == 0 {
      htmlSection = "<h1>%d    %s</h1>"
    } else {
      htmlSection = "<div class = \"page-break\"> <h1>%d %s</h1> </div>"
    }
    htmlContent += fmt.Sprintf(htmlSection, i+1, sectionName)

    subSectionData, ok := sectionData["subsections"].([]map[string]interface{})
    if !ok {
      continue
    }

    for j, subsection := range subSectionData {
      subSectionTitle := subsection["title"].(string)
      htmlSubsection := "<div class = \"avoid-break\"> <h2>%d.%d    %s</h2>"
			htmlContent += fmt.Sprintf(htmlSubsection, i+1, j+1, subSectionTitle)

			subSectionContent, ok := subsection["content"].(string)
			if !ok || subSectionContent == "" {
				return fmt.Errorf("missing or invalid content for subsection: %v", subsection["title"])
			}

			var parsedDelta delta.Delta
			err := json.Unmarshal([]byte(subSectionContent), &parsedDelta)
			if err != nil {
				return fmt.Errorf("missing or invalid content for subsection: %v, in section: %s", subsection["title"],  sectionName)
			}

			opsJSON, err := json.Marshal(parsedDelta.Delta.Delta.Ops)
			if err != nil {
				return fmt.Errorf("error marshalling Ops: %v", err)
			}

			html, err := quill.Render(opsJSON)
			if err != nil {
				return fmt.Errorf("error converting Delta to HTML: %v", err)
			}


			if len(html) < 23 {
				return fmt.Errorf("missing or invalid content for subsection: %v, in section: %s", subsection["title"],  sectionName)
			}
			fmt.Println(len(html))

			htmlContent += fmt.Sprintf("%s </div>", html)
		}
	}

	htmlContent += "</body></html>"

	return saveHTMLToPDF(reportName+".pdf", htmlContent)
}

func saveHTMLToPDF(pdfPath, htmlContent string) error {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var buf []byte
	err := chromedp.Run(ctx,
	chromedp.Navigate("about:blank"),
	chromedp.ActionFunc(func(ctx context.Context) error {
		return chromedp.Evaluate(fmt.Sprintf(`document.documentElement.innerHTML = '%s';`, htmlContent), nil).Do(ctx)
	}),
	chromedp.WaitVisible("body", chromedp.ByQuery),
	chromedp.Sleep(5*time.Second),
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
	return fmt.Errorf("failed to render PDF: %v", err)
}

err = os.WriteFile(pdfPath, buf, 0644)
if err != nil {
	return fmt.Errorf("failed to save PDF: %v", err)
}

fmt.Println("PDF successfully generated:", pdfPath)
return nil
}
