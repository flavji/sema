package main

import (
		"log"

		"sema/api/handlers"

		"github.com/gin-gonic/gin"
		"sema/api/routes"
		"sema/repository"
)

const (
	projectID = "sema-7c193"
)


func main() {


r := gin.Default()
		repo, err := repository.NewFirestoreRepository(projectID, "../config/firebase_credentials.json")
		if err != nil {
			log.Fatalf("Failed to initialize Firestore: %v", err)
		}

		r.Static("/static", "../static")
		// r.StaticFS("/static", gin.Dir("../static", false)) // do not cache
		r.LoadHTMLGlob("../templates/*")

		apiRoutes := r.Group("/api")
		{
			apiRoutes.POST("/reports", handlers.CreateReportHandler(repo))
			apiRoutes.POST("/generateReport", handlers.GenerateReportHandler(repo))
		}

		routes.SetupWebSocketRoutes(r, repo)
		r.GET("/", handlers.HomeHandler)
		r.GET("/report/:reportID", handlers.ReportHandler(repo))

		log.Println("Server running on port 8080")
		r.Run(":8080")

	}
