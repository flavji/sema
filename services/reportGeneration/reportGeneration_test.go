package reportGeneration_test

import (
	"os"
	"testing"
	"sema/services/reportGeneration"
)

func TestGeneratePDF(t *testing.T) {
	report := []map[string]interface{}{
		{
			"sectionTitle": "Introduction",
			"subsections": []map[string]interface{}{
				{
					"title": "Overview",
					"content": `{"delta":{"delta":{"ops":[{"insert":"This is a test paragraph.\n"}]}}}`,
				},
				{
					"title": "Scope",
					"content": `{"delta":{"delta":{"ops":[{"insert":"Scope details go here.\n"}]}}}`,
				},
			},
		},
	}

	err := reportGeneration.GeneratePDF("TestReport", report)
	if err != nil {
		t.Errorf("GeneratePDF returned error: %v", err)
	}

	if _, err := os.Stat("TestReport.pdf"); os.IsNotExist(err) {
		t.Error("PDF file was not created")
	} else {
		_ = os.Remove("TestReport.pdf") // clean up
	}
}

