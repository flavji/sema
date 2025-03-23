package routes

import (
	"github.com/gin-gonic/gin"
	"sema/api/handlers"
	"sema/repository"
)


func SetupWebSocketRoutes(router *gin.Engine, repo *repository.FirestoreRepository) {
	router.GET("/report/:reportID/section/:sectionID", handlers.WebSocketHandler(repo))
}
