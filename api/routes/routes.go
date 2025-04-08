package routes

import (
	"github.com/gin-gonic/gin"
	"sema/api/handlers"
	"sema/repository"
	"sema/services/authentication"
	"sema/api/middleware"
)

func SetupRoutes(router *gin.Engine, authService *authentication.AuthService, repo *repository.FirestoreRepository) {
	// Public routes (No authentication required)

	/* Routes for registering */
	router.POST("/api/auth/verify", handlers.VerifyTokenHandler(authService))
	router.GET("/register", handlers.RegisterHandler)

	// Protected routes (Require authentication)
	home := router.Group("/")
	home.Use(middleware.AuthMiddleware(authService))  

	report := router.Group("/report/:reportID")
	report.Use(middleware.AuthMiddleware(authService))
	report.Use(middleware.AuthUserinReport(authService, repo))


	reportAdmin := router.Group("/report/:reportID/")
	reportAdmin.Use(middleware.AuthMiddleware(authService))
	reportAdmin.Use(middleware.AuthAdmininReport(authService, repo))




	// Home & report routes (protected)
	home.GET("/", handlers.HomeHandler(repo))
	home.POST("/api/reports", handlers.CreateReportHandler(repo))
	home.DELETE("/api/deleteaccount", handlers.DeleteAccount(authService, repo))


	report.GET("/", handlers.ReportHandler(repo))
	report.GET("/section/:sectionID", handlers.WebSocketHandler(repo))
	report.GET("/api/isadmin", handlers.IsAdmin(repo))

	reportAdmin.POST("/api/addusertoreport", handlers.AddUserToReport(authService, repo))
	reportAdmin.GET("/api/generateReport", handlers.GenerateReportHandler(repo))
	reportAdmin.DELETE("/api/removeuser", handlers.RemoveUserFromReport(authService, repo))
	reportAdmin.POST("/api/renamereport", handlers.RenameReport(repo))
	reportAdmin.DELETE("/api/deletereport", handlers.DeleteReport(repo))
	reportAdmin.GET("/logs", handlers.ReportLogsHandler(repo))

}

// Disabled middleware protection for load testing purposes
func LoadTestSetupRoutes(router *gin.Engine, authService *authentication.AuthService, repo *repository.FirestoreRepository) {
	// Public routes (No authentication required)

	/* Routes for registering */
	router.POST("/api/auth/verify", handlers.VerifyTokenHandler(authService))
	router.GET("/register", handlers.RegisterHandler)

	// Protected routes (Require authentication)
	home := router.Group("/")
	// home.Use(middleware.AuthMiddleware(authService))  

	report := router.Group("/report/:reportID")
	// report.Use(middleware.AuthMiddleware(authService))
	// report.Use(middleware.AuthUserinReport(authService, repo))


	reportAdmin := router.Group("/report/:reportID/")
	// reportAdmin.Use(middleware.AuthMiddleware(authService))
	// reportAdmin.Use(middleware.AuthAdmininReport(authService, repo))




	// Home & report routes (protected)
	home.GET("/", handlers.HomeHandler(repo))
	home.POST("/api/reports", handlers.CreateReportHandler(repo))
	home.DELETE("/api/deleteaccount", handlers.DeleteAccount(authService, repo))


	report.GET("/", handlers.ReportHandler(repo))
	report.GET("/section/:sectionID", handlers.WebSocketHandler(repo))
	report.GET("/api/isadmin", handlers.IsAdmin(repo))

	reportAdmin.POST("/api/addusertoreport", handlers.AddUserToReport(authService, repo))
	reportAdmin.GET("/api/generateReport", handlers.GenerateReportHandler(repo))
	reportAdmin.DELETE("/api/removeuser", handlers.RemoveUserFromReport(authService, repo))
	reportAdmin.POST("/api/renamereport", handlers.RenameReport(repo))
	reportAdmin.DELETE("/api/deletereport", handlers.DeleteReport(repo))
	reportAdmin.GET("/logs", handlers.ReportLogsHandler(repo))

}
