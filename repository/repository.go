package repository

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"sema/models/reportTemplates"
	"sema/services/firebase"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)


type ReportRepository interface {
	IsUserInReport(uid, reportID string) (bool, error)
	IsAdminInReport(uid, reportID string) (bool, error)
	GetUserReportLinks(uid string) ([]Report, error)
	LinkReportWithUser(uID, reportID string, privilege bool, ownership bool) error
	GetReportFieldTemplateID(reportID string) (string, error)
	GetTemplate(templateID string) (*reportTemplates.ReportTemplate, error)
	CreateReport(reportName, reportID, templateID, userEmail string) error
	FetchReportContent(reportID string) (string, []map[string]interface{}, error)
	FetchLogsForReport(reportID string) ([]string, error)
	RemoveUserFromReport(uID, reportID string) error
	RenameReport(reportID, reportName string) error
	DeleteReport(reportID string) error
	DestroyUser(uID string) error
	FetchReportSectionContents(reportID, sectionTitle string) (map[string]string, error)
	UpdateReportSectionContents(reportID, sectionTitle, subsectionTitle, newContent string) error
	BufferLog(reportID, message, user string)
}

type FirestoreRepository struct {
	Client     *firestore.Client
	Ctx        context.Context
	logBuffers map[string][]map[string]interface{}
	logMu      sync.Mutex
}


// NewFirestoreRepository initializes Firestore using shared FirebaseApp
func NewFirestoreRepository(firebaseApp *firebase.FirebaseApp, projectID string) (*FirestoreRepository, error) {
	ctx := context.Background()

	firestoreClient, err := firebaseApp.App.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("error initializing Firestore: %v", err)
	}

	repo := &FirestoreRepository{
		Client:     firestoreClient,
		Ctx:        ctx,
		logBuffers: make(map[string][]map[string]interface{}),
	}

	// Start periodic log flushing
	repo.StartLogFlusher(30 * time.Second)

	return repo, nil
}

func (r *FirestoreRepository) BufferLog(reportID, message, user string) {
	r.logMu.Lock()
	defer r.logMu.Unlock()

	entry := map[string]interface{}{
		"timestamp": time.Now(),
		"message":   user + " " + message,
		"userID":    user,
	}
	r.logBuffers[reportID] = append(r.logBuffers[reportID], entry)

	if len(r.logBuffers[reportID]) >= 10 {
		go r.FlushLogs(reportID)
	}
}

func (r *FirestoreRepository) FlushLogs(reportID string) {
	r.logMu.Lock()
	logs := r.logBuffers[reportID]
	r.logBuffers[reportID] = nil
	r.logMu.Unlock()

	if len(logs) == 0 {
		return
	}

	batch := r.Client.Batch()
	logCol := r.Client.Collection("reports").Doc(reportID).Collection("logs")

	for _, entry := range logs {
		doc := logCol.NewDoc()
		batch.Set(doc, entry)
	}

	_, err := batch.Commit(r.Ctx)
	if err != nil {
		log.Printf("Failed to flush logs for report %s: %v", reportID, err)
	} else {
		log.Printf("Flushed %d logs for report %s", len(logs), reportID)
	}
}

func (r *FirestoreRepository) StartLogFlusher(interval time.Duration) {
	go func() {
		for {
			time.Sleep(interval)
			r.logMu.Lock()
			for reportID := range r.logBuffers {
				if len(r.logBuffers[reportID]) > 0 {
					go r.FlushLogs(reportID)
				}
			}
			r.logMu.Unlock()
		}
	}()
}

func (r *FirestoreRepository) GetTemplate(templateID string) (*reportTemplates.ReportTemplate, error) {
	var template *reportTemplates.ReportTemplate



	doc, err := r.Client.Collection("templates").Doc(templateID).Get(r.Ctx)
	if err != nil {
		return nil, err
	}

	doc.DataTo(&template)

	return template, nil

}



func (r *FirestoreRepository) CreateReport(reportName, reportID, templateID, userEmail string) error {
	template, err := r.GetTemplate(templateID)
	if err != nil {
		return err
	}

	newReportDoc := r.Client.Collection("reports").Doc(reportID)

	// Create report document
	_, err = newReportDoc.Set(r.Ctx, map[string]interface{}{
		"reportID" : reportID,
		"templateID" : templateID,
		"reportName": reportName,
		"creationTime": time.Now(),	
	})
	if err != nil {
		log.Fatalf("Failed to create report document: %v", err)
	}

	// ✅ Add a log entry

	logEntry := map[string]interface{}{
		"timestamp": time.Now(),
		"message":   fmt.Sprintf("Report created by user %s", userEmail),
		"userID":    userEmail,
	}

	_, err = newReportDoc.Collection("logs").NewDoc().Set(r.Ctx, logEntry)
	if err != nil {
		return fmt.Errorf("failed to create report log: %w", err)
	}

	// Iterate over sections with an index
	for sectionIndex, section := range template.Sections {
		fmt.Printf("Adding Section: %s\n", section.Title)


		// hash section title
		sectionDocTitle := sanitizeFirebaseDocName(section.Title)
		sectionDocRef := newReportDoc.Collection("sections").Doc(sectionDocTitle)

		// Store section with its order index
		_, err := sectionDocRef.Set(r.Ctx, map[string]interface{}{
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
			_, err := subsectionDocRef.Set(r.Ctx, map[string]interface{}{
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
	reportDocRef := r.Client.Collection("reports").Doc(reportID)

	// Fetch the report document
	reportSnap, err := reportDocRef.Get(r.Ctx)
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
	sectionDocRef := r.Client.Collection("reports").Doc(reportID).Collection("sections").Doc(sectionTitle)
	subsectionsSnap, err := sectionDocRef.Collection("subsections").Documents(r.Ctx).GetAll()
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
	subsectionDocRef := r.Client.Collection("reports").Doc(reportID).Collection("sections").Doc(sanitieSectionTitle).Collection("subsections").Doc(sanitizeSubSectionTitle)
	_, err := subsectionDocRef.Update(r.Ctx, []firestore.Update{
		{Path: "content", Value: newContent},
	})
	if err != nil {
		return fmt.Errorf("failed to update subsection content: %v", err)
	}
	return nil
}

func (r *FirestoreRepository) FetchReportContent(reportID string) (string ,[]map[string]interface{}, error) {
	docRef := r.Client.Collection("reports").Doc(reportID)

	reportDoc, err := docRef.Get(r.Ctx)
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
	sectionsDocs, err := sectionsQuery.Documents(r.Ctx).GetAll()
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
		subsectionsDocs, err := subsectionsQuery.Documents(r.Ctx).GetAll()
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


func (r *FirestoreRepository) LinkReportWithUser(uID, reportID string, privilege bool, ownership bool) error {
	// Get a reference to the subcollection "linkedReports" inside the user document
	reportDocRef := r.Client.Collection("users").Doc(uID).Collection("linkedReports").Doc(reportID)

	// Get the document snapshot to check if it exists and if the privilege is already set to true
	docSnapshot, _ := reportDocRef.Get(r.Ctx)

	if docSnapshot.Exists() {
		// If the document exists, check the current privilege level
		existingPrivilege, exists := docSnapshot.Data()["privilege"].(bool)
		if exists && existingPrivilege {
			// If the privilege is already true, no further changes should be made
			fmt.Println("Privilege level is already true, no changes allowed.")
			return nil
		}
	}

	// If the document doesn't exist or privilege is not true, update the document
	_, err := reportDocRef.Set(r.Ctx, map[string]interface{}{
		"privilege": privilege,
		"owner": ownership,
	})

	if err != nil {
		return fmt.Errorf("failed to link report with user: %v", err)
	}

	fmt.Println("Linked report with user in subcollection")
	return nil
}




type Report struct {
	ReportID    string    `json:"reportID"`
	ReportTitle  string    `json:"reportTitle"`
	CreationTime time.Time `json:"creationTime"`
}

func (r *FirestoreRepository) GetUserReportLinks(uID string) ([]Report, error) {
	// Get a reference to the "linkedReports" subcollection inside the user document
	reportsCollection := r.Client.Collection("users").Doc(uID).Collection("linkedReports")

	// Query all documents in the subcollection
	docs, err := reportsCollection.Documents(r.Ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get linked reports for user: %v", err)
	}

	// Initialize a slice to store the reports
	var reports []Report

	for _, doc := range docs {
		reportID := doc.Ref.ID // Report ID is the document ID

		// Fetch the actual report document from the "reports" collection
		reportDocRef := r.Client.Collection("reports").Doc(reportID)
		reportDoc, err := reportDocRef.Get(r.Ctx)
		if err != nil {
			log.Printf("failed to get report document for reportID: %v\n", reportID)
			log.Printf("removing reportID: %v, from users: %s, colleciton\n", reportID, uID)
			// remove the report collection here 
			// Delete broken link
			_, delErr := doc.Ref.Delete(r.Ctx)
			if delErr != nil {
				log.Printf("Failed to delete broken report link: %v", delErr)
			}

			continue // Skips broken report

		}

		// Extract relevant fields from the report document
		reportData := reportDoc.Data()

		creationTime, ok := reportData["creationTime"].(time.Time)
		if !ok {
			return nil, fmt.Errorf("invalid creationTime format in report document for reportID: %v", reportID)
		}

		reportName, ok := reportData["reportName"].(string)
		if !ok {
			return nil, fmt.Errorf("missing reportName in report document for reportID: %v", reportID)
		}

		// Append the report to the reports slice
		reports = append(reports, Report{
			ReportID:     reportID,
			ReportTitle:  reportName,
			CreationTime: creationTime,
		})
	}
	// Sort the reports slice by creationTime (descending)
	sort.Slice(reports, func(i, j int) bool {
		return reports[i].CreationTime.After(reports[j].CreationTime)
	})

	return reports, nil
}

func (r *FirestoreRepository) IsUserInReport(uID, reportID string) (bool, error) {
	// Get a reference to the specific report document in the user's "linkedReports" subcollection
	reportDocRef := r.Client.Collection("users").Doc(uID).Collection("linkedReports").Doc(reportID)

	// Check if the document exists
	doc, err := reportDocRef.Get(r.Ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return false, nil // Report is not linked to the user
		}
		return false, fmt.Errorf("failed to check if user is in report: %v", err)
	}

	return doc.Exists(), nil
}


func (r *FirestoreRepository) IsAdminInReport(uID, reportID string) (bool, error) {
	// Get a reference to the specific report document in the user's "linkedReports" subcollection
	reportDocRef := r.Client.Collection("users").Doc(uID).Collection("linkedReports").Doc(reportID)

	// Try to get the document
	doc, err := reportDocRef.Get(r.Ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return false, nil // Report is not linked to the user
		}
		return false, fmt.Errorf("failed to check if user is in report: %w", err)
	}

	// Check privilege flag
	privilege, ok := doc.Data()["privilege"].(bool)
	if !ok {
		return false, nil // Privilege not set or wrong type
	}

	return privilege, nil
}


func (r *FirestoreRepository) RemoveUserFromReport(uID, reportID string) error {
	reportDocRef := r.Client.Collection("users").Doc(uID).Collection("linkedReports").Doc(reportID)

	// Get the document snapshot
	docSnapshot, err := reportDocRef.Get(r.Ctx)
	if err != nil {
		return fmt.Errorf("failed to get report document: %w", err)
	}

	if docSnapshot.Exists() {
		if privilege, ok := docSnapshot.Data()["privilege"].(bool); ok && privilege {
			fmt.Println("User is Admin, cannot be removed")
			return nil
		}
	}

	// Delete the document to remove user from report
	_, err = reportDocRef.Delete(r.Ctx)
	if err != nil {
		return fmt.Errorf("failed to delete report document: %w", err)
	}

	fmt.Println("Removed user from report")
	return nil
}

func (r *FirestoreRepository) RenameReport(reportID, reportName string) error {
	reportDoc := r.Client.Collection("reports").Doc(reportID)

	_, err := reportDoc.Set(r.Ctx, map[string]interface{}{
		"reportName": reportName,
	}, firestore.MergeAll) // Merge to only update "reportName"

	if err != nil {
		return fmt.Errorf("failed to rename report: %w", err)
	}

	return nil
}

func (r *FirestoreRepository) DeleteReport(reportID string) error {
	reportDoc := r.Client.Collection("reports").Doc(reportID)

	// Get all sections
	sectionsIter := reportDoc.Collection("sections").Documents(r.Ctx)
	defer sectionsIter.Stop()

	for {
		sectionDoc, err := sectionsIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to iterate sections: %w", err)
		}

		// Get all subsections in this section
		subsectionsIter := sectionDoc.Ref.Collection("subsections").Documents(r.Ctx)
		for {
			subDoc, err := subsectionsIter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return fmt.Errorf("failed to iterate subsections: %w", err)
			}
			_, err = subDoc.Ref.Delete(r.Ctx)
			if err != nil {
				return fmt.Errorf("failed to delete subsection: %w", err)
			}
		}

		// Delete section after its subsections
		_, err = sectionDoc.Ref.Delete(r.Ctx)
		if err != nil {
			return fmt.Errorf("failed to delete section: %w", err)
		}
	}

	logEntry := map[string]interface{}{
		"timestamp": time.Now(),
		"message":   "Report was deleted",
	}

	_, err := reportDoc.Collection("logs").NewDoc().Set(r.Ctx, logEntry)
	if err != nil {
		return fmt.Errorf("failed to log report deletion: %w", err)
	}

	// Finally delete the report doc
	_, err = reportDoc.Delete(r.Ctx)
	if err != nil {
		return fmt.Errorf("failed to delete report: %w", err)
	}

	fmt.Println("Deleted report and all nested sections/subsections")
	return nil
}


func (r *FirestoreRepository) DestroyUser(uID string) error {


	reportsCollection := r.Client.Collection("users").Doc(uID).Collection("linkedReports")

	// Query all documents in the subcollection
	docs, err := reportsCollection.Documents(r.Ctx).GetAll()
	if err != nil {
		return fmt.Errorf("failed to get linked reports for user: %v", err)
	}

	for _, doc := range docs {
		isOwner, exists := doc.Data()["owner"].(bool)
		if exists && isOwner {
			reportID := doc.Ref.ID
			if err := r.DeleteReport(reportID); err != nil {
				return fmt.Errorf("failed to delete owned report %s: %w", reportID, err)
			}
		}

		// Delete the linked report reference
		_, err := doc.Ref.Delete(r.Ctx)
		if err != nil {
			return fmt.Errorf("failed to delete linked report reference: %w", err)
		}
	}

	// Delete the user document itself
	_, err = r.Client.Collection("users").Doc(uID).Delete(r.Ctx)
	if err != nil {
		return fmt.Errorf("failed to delete user document: %v", err)
	}

	fmt.Println("Successfully deleted user and related data")
	return nil
}


func (r *FirestoreRepository) FetchLogsForReport(reportID string) ([]string, error) {
	logsCollection := r.Client.Collection("reports").Doc(reportID).Collection("logs")

	// Get all log documents
	docs, err := logsCollection.OrderBy("timestamp", firestore.Asc).Documents(r.Ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch logs for report: %w", err)
	}

	var logs []string
	for _, doc := range docs {
		message, ok := doc.Data()["message"].(string)
		if !ok {
			continue // skip invalid entries
		}

		// Optionally include timestamp
		timestamp, _ := doc.Data()["timestamp"].(time.Time)
		logs = append(logs, fmt.Sprintf("[%s] %s", timestamp.Format("2006-01-02 15:04:05"), message))
	}

	return logs, nil
}


func sanitizeFirebaseDocName(name string) string {
	return strings.ReplaceAll(name, "/", "_")
}



// NewTestFirestoreRepository is only for testing.
func NewTestFirestoreRepository(client *firestore.Client, ctx context.Context) *FirestoreRepository {
	repo := &FirestoreRepository{
		Client:     client,
		Ctx:        ctx,
		logBuffers: make(map[string][]map[string]interface{}),
	}
	repo.StartLogFlusher(1 * time.Second) // optional during tests
	return repo
}
