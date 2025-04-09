package repository_test

import (
	"context"
	"os"
	"testing"

	"cloud.google.com/go/firestore"
	"github.com/stretchr/testify/assert"
	"sema/repository"
)

func setupTestRepo(t *testing.T) *repository.FirestoreRepository {
	os.Setenv("FIRESTORE_EMULATOR_HOST", "127.0.0.1:8080")

	ctx := context.Background()
	client, err := firestore.NewClient(ctx, "test-project")
	if err != nil {
		t.Fatalf("Failed to create Firestore client: %v", err)
	}

	_, _ = client.Collection("templates").Doc("template123").Set(ctx, map[string]interface{}{
		"Sections": []map[string]interface{}{
			{
				"Title":       "Introduction",
				"Subsections": []string{"Overview", "Scope"},
			},
		},
	})

	return repository.NewTestFirestoreRepository(client, ctx)
}

func TestCreateReport(t *testing.T) {
	repo := setupTestRepo(t)
	err := repo.CreateReport("Test Report", "test-report-123", "template123", "test@example.com")
	assert.NoError(t, err)
	doc, err := repo.Client.Collection("reports").Doc("test-report-123").Get(repo.Ctx)
	assert.NoError(t, err)
	assert.Equal(t, "Test Report", doc.Data()["reportName"])
}

func TestLinkReportWithUser(t *testing.T) {
	repo := setupTestRepo(t)
	err := repo.LinkReportWithUser("user123", "test-report-123", true, true)
	assert.NoError(t, err)
	doc, err := repo.Client.Collection("users").Doc("user123").Collection("linkedReports").Doc("test-report-123").Get(repo.Ctx)
	assert.NoError(t, err)
	assert.Equal(t, true, doc.Data()["privilege"])
}

func TestBufferAndFlushLogs(t *testing.T) {
	repo := setupTestRepo(t)
	reportID := "log-test-report"
	repo.BufferLog(reportID, "log message", "user@example.com")
	repo.FlushLogs(reportID)
	docs, err := repo.Client.Collection("reports").Doc(reportID).Collection("logs").Documents(repo.Ctx).GetAll()
	assert.NoError(t, err)
	assert.Greater(t, len(docs), 0)
}

func TestGetTemplate(t *testing.T) {
	repo := setupTestRepo(t)
	tmpl, err := repo.GetTemplate("template123")
	assert.NoError(t, err)
	assert.NotNil(t, tmpl)
}

func TestGetReportFieldTemplateID(t *testing.T) {
	repo := setupTestRepo(t)
	repo.CreateReport("Template Report", "template-report-id", "template123", "test@example.com")
	tmplID, err := repo.GetReportFieldTemplateID("template-report-id")
	assert.NoError(t, err)
	assert.Equal(t, "template123", tmplID)
}

func TestFetchReportSectionContentsAndUpdate(t *testing.T) {
	repo := setupTestRepo(t)
	repo.CreateReport("Test", "section-test-id", "template123", "test@example.com")
	err := repo.UpdateReportSectionContents("section-test-id", "Introduction", "Overview", "Updated Content")
	assert.NoError(t, err)
	contents, err := repo.FetchReportSectionContents("section-test-id", "Introduction")
	assert.NoError(t, err)
	assert.Equal(t, "Updated Content", contents["Overview"])
}

func TestFetchReportContent(t *testing.T) {
	repo := setupTestRepo(t)
	repo.CreateReport("Full Report", "full-report-id", "template123", "test@example.com")
	title, sections, err := repo.FetchReportContent("full-report-id")
	assert.NoError(t, err)
	assert.Equal(t, "Full Report", title)
	assert.Greater(t, len(sections), 0)
}

func TestUserReportLinkingAndFetching(t *testing.T) {
	repo := setupTestRepo(t)
	repo.CreateReport("User Linked Report", "linked-report-id", "template123", "user@test.com")
	repo.LinkReportWithUser("userUID", "linked-report-id", true, true)
	reports, err := repo.GetUserReportLinks("userUID")
	assert.NoError(t, err)
	assert.NotEmpty(t, reports)
}

func TestUserAdminChecks(t *testing.T) {
	repo := setupTestRepo(t)
	repo.CreateReport("Admin Report", "admin-report-id", "template123", "admin@test.com")
	repo.LinkReportWithUser("adminUID", "admin-report-id", true, true)
	isIn, _ := repo.IsUserInReport("adminUID", "admin-report-id")
	isAdmin, _ := repo.IsAdminInReport("adminUID", "admin-report-id")
	assert.True(t, isIn)
	assert.True(t, isAdmin)
}

func TestRemoveUserFromReport(t *testing.T) {
	repo := setupTestRepo(t)
	repo.CreateReport("Removable Report", "remove-report-id", "template123", "user@test.com")
	repo.LinkReportWithUser("userToRemove", "remove-report-id", false, false)
	err := repo.RemoveUserFromReport("userToRemove", "remove-report-id")
	assert.NoError(t, err)
}

func TestRenameReport(t *testing.T) {
	repo := setupTestRepo(t)
	repo.CreateReport("Rename Me", "rename-report-id", "template123", "renamer@test.com")
	err := repo.RenameReport("rename-report-id", "New Name")
	assert.NoError(t, err)
	doc, _ := repo.Client.Collection("reports").Doc("rename-report-id").Get(repo.Ctx)
	assert.Equal(t, "New Name", doc.Data()["reportName"])
}

func TestDeleteReport(t *testing.T) {
	repo := setupTestRepo(t)
	repo.CreateReport("Delete Me", "delete-report-id", "template123", "deleter@test.com")
	err := repo.DeleteReport("delete-report-id")
	assert.NoError(t, err)
}

func TestDestroyUser(t *testing.T) {
	repo := setupTestRepo(t)
	repo.CreateReport("Own Report", "owned-report-id", "template123", "owner@test.com")
	repo.LinkReportWithUser("ownerUID", "owned-report-id", true, true)
	err := repo.DestroyUser("ownerUID")
	assert.NoError(t, err)
}

func TestFetchLogsForReport(t *testing.T) {
	repo := setupTestRepo(t)
	reportID := "logs-report-id"
	repo.BufferLog(reportID, "first log", "tester@test.com")
	repo.FlushLogs(reportID)
	logs, err := repo.FetchLogsForReport(reportID)
	assert.NoError(t, err)
	assert.NotEmpty(t, logs)
}

