package repository

import (
	"context"
	"fmt"
	"log"
	"strings"

	"sema/models/reportTemplates"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
)

type FirestoreRepository struct {
	client *firestore.Client
	ctx    context.Context
}

func NewFirestoreRepository(projectID, credentialsPath string) (*FirestoreRepository, error) {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, projectID, option.WithCredentialsFile(credentialsPath))
	if err != nil {
		return nil, err
	}
	return &FirestoreRepository{client: client, ctx: ctx}, nil
}

func (r *FirestoreRepository) GetTemplate(templateID string) (*reportTemplates.ReportTemplate, error) {
	var template *reportTemplates.ReportTemplate


	doc, err := r.client.Collection("templates").Doc(templateID).Get(r.ctx)
	if err != nil {
		return nil, err
	}

	doc.DataTo(&template)

	return template, nil

}



func (r *FirestoreRepository) CreateReport(reportName, reportID, templateID string) error {
	template, err := r.GetTemplate(templateID)
	if err != nil {
		return err
	}

	newReportDoc := r.client.Collection("reports").Doc(reportID)

	// Create report document
	_, err = newReportDoc.Set(r.ctx, map[string]interface{}{
		"reportID" : reportID,
		"templateID" : templateID,
		"reportName": reportName,
		"users":      "",
		"admins":     "",
	})
	if err != nil {
		log.Fatalf("Failed to create report document: %v", err)
	}

	// Iterate over sections with an index
	for sectionIndex, section := range template.Sections {
		fmt.Printf("Adding Section: %s\n", section.Title)


		// hash section title
		sectionDocTitle := sanitizeFirebaseDocName(section.Title)
		sectionDocRef := newReportDoc.Collection("sections").Doc(sectionDocTitle)

		// Store section with its order index
		_, err := sectionDocRef.Set(r.ctx, map[string]interface{}{
			"title": section.Title,
			"order": sectionIndex, // Preserve order
		})
		if err != nil {
			log.Fatalf("Failed to add section: %v", err)
		}

		// Iterate over subsections with an index
		for subsectionIndex, subsection := range section.Subsections {
			fmt.Printf("Adding Subsection: %s under Section: %s\n", subsection, section.Title)

			//subsectionDocRef := sectionDocRef.Collection("subsections").NewDoc()
			// hash the subsection title to generate a id 
			subSectionDocTitle := sanitizeFirebaseDocName(subsection)
			subsectionDocRef := sectionDocRef.Collection("subsections").Doc(subSectionDocTitle)

			// Store subsection with its order index
			_, err := subsectionDocRef.Set(r.ctx, map[string]interface{}{
				"title":   subsection,
				"content": "",
				"order":   subsectionIndex, // Preserve order
			})
			if err != nil {
				log.Fatalf("Failed to add subsection: %v", err)
			}
		}
	}

	return nil
}


func (r *FirestoreRepository) GetReportFieldTemplateID(reportID string) (string, error) {
	reportDocRef := r.client.Collection("reports").Doc(reportID)

	// Fetch the report document
	reportSnap, err := reportDocRef.Get(r.ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get report: %v", err)
	}

	reportData := reportSnap.Data()
	templateID, ok := reportData["templateID"].(string)
	if !ok {
		return "", fmt.Errorf("templateID not found in report")
	}
	return templateID, nil
}




func (r *FirestoreRepository) FetchReportSectionContents(reportID, sectionTitle string) (map[string]string, error) {
	sectionDocRef := r.client.Collection("reports").Doc(reportID).Collection("sections").Doc(sectionTitle)
	subsectionsSnap, err := sectionDocRef.Collection("subsections").Documents(r.ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch subsections: %v", err)
	}

	contents := make(map[string]string)
	for _, doc := range subsectionsSnap {
		data := doc.Data()
		title, _ := data["title"].(string)
		content, _ := data["content"].(string)
		contents[title] = content
	}
	return contents, nil
}

func (r *FirestoreRepository) UpdateReportSectionContents(reportID, sectionTitle, subsectionTitle, newContent string) error {
	sanitieSectionTitle := sanitizeFirebaseDocName(sectionTitle)
	sanitizeSubSectionTitle := sanitizeFirebaseDocName(subsectionTitle)
	subsectionDocRef := r.client.Collection("reports").Doc(reportID).Collection("sections").Doc(sanitieSectionTitle).Collection("subsections").Doc(sanitizeSubSectionTitle)
	_, err := subsectionDocRef.Update(r.ctx, []firestore.Update{
		{Path: "content", Value: newContent},
	})
	if err != nil {
		return fmt.Errorf("failed to update subsection content: %v", err)
	}
	return nil
}

func (r *FirestoreRepository) FetchReportContent(reportID string) (string ,[]map[string]interface{}, error) {
    docRef := r.client.Collection("reports").Doc(reportID)

		reportDoc, err := docRef.Get(r.ctx)
    if err != nil {
				return "",nil, fmt.Errorf("failed to fetch report: %w", err)
    }
		
		reportData := reportDoc.Data()
    reportName, ok := reportData["reportName"].(string)
    if !ok {
        return "", nil, fmt.Errorf("report name missing or invalid")
    }

    // Get the sections collection ordered by the "order" field
    sectionsQuery := docRef.Collection("sections").OrderBy("order", firestore.Asc)
    sectionsDocs, err := sectionsQuery.Documents(r.ctx).GetAll()
    if err != nil {
        return "", nil, fmt.Errorf("failed to fetch sections: %w", err)
    }

    var orderedReportContent []map[string]interface{}

    // Iterate over the sections to ensure they're in order and process subsections
    for _, sectionDoc := range sectionsDocs {
        sectionData := sectionDoc.Data()
        sectionTitle := sectionData["title"].(string)

        // Get subsections collection ordered by "order" field
        subsectionsQuery := sectionDoc.Ref.Collection("subsections").OrderBy("order", firestore.Asc)
        subsectionsDocs, err := subsectionsQuery.Documents(r.ctx).GetAll()
        if err != nil {
						return "", nil, fmt.Errorf("failed to fetch subsections for section %s: %w", sectionTitle, err)
        }

        subsections := []map[string]interface{}{}
        for _, subsectionDoc := range subsectionsDocs {
            subsectionData := subsectionDoc.Data()
            subsectionContent := subsectionData["content"] // Get the "content" field
            subsectionTitle := subsectionData["title"].(string)

            // Append the subsection as a map with title and content
            subsections = append(subsections, map[string]interface{}{
                "title":   subsectionTitle,
                "content": subsectionContent,
            })
        }

        // Add the section and its subsections to the report content as a map
        orderedReportContent = append(orderedReportContent, map[string]interface{}{
            "sectionTitle": sectionTitle,
            "subsections":  subsections,
        })
    }

    return reportName, orderedReportContent, nil
}
	
func sanitizeFirebaseDocName(name string) string {
    return strings.ReplaceAll(name, "/", "_")
}


