package main

import (
	"log"
	"github.com/gin-gonic/gin"
	"sema/api/routes"
	"sema/repository"
	"sema/services/authentication"
	"sema/services/firebase"
)

const (
	projectID = "sema-7c193"
)


func main() {


	r := gin.Default()

	firebaseApp, err := firebase.NewFirebaseApp("../../config/firebase_credentials.json")
	if err != nil {
		log.Fatalf("Failed to initialize Firebase: %v", err)
	}

	// Initialize Firestore Repo using the shared Firebase App
	repo, err := repository.NewFirestoreRepository(firebaseApp, projectID)
	if err != nil {
		log.Fatalf("Failed to initialize Firestore: %v", err)
	}
	defer repo.Client.Close() // Ensure Firestore client is closed on exit

	// Initialize Auth Service using the shared Firebase App
	authService, err := authentication.NewAuthService(firebaseApp)
	if err != nil {
		log.Fatalf("Failed to initialize Firebase Auth: %v", err)
	}		


	r.Static("/static", "../../static")
	r.LoadHTMLGlob("../../templates/*")

	routes.SetupRoutes(r, authService, repo)

	log.Println("Server running on port 8080")
	r.Run(":8080")

}
