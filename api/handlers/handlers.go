package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sema/models/delta"
	"sema/models/joinMessage"
	"sema/models/syncMessage"
	"sema/models/updateRepoMessage"
	"sema/repository"
	"time"

	"sema/services/authentication"
	"sema/services/reportGeneration"
	"sema/services/websockets"

	"github.com/gin-gonic/gin"
)

var websocketmanager = websockets.SpawnWebSocketManager()

type ReportRequest struct { 
	Type string `json:"type" binding:"required"`
	Name string `json:"name" binding:"required"`
}

type ReportGenerationRequest struct {
	Type string `json:"type" binding:"required"`
	ReportID string `json:"reportID" binding:"required"`

}

func generateID() string {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

func HomeHandler(repo repository.ReportRepository) gin.HandlerFunc {

	return func(c *gin.Context) {
	uID, ok := c.Get("uid")
		if !ok {
			// Handle error if the uid is not found in the context
			c.JSON(http.StatusBadRequest, gin.H{"error": "uid not found"})
			return
		}

		// Assert uID to string
		uidStr, ok := uID.(string)
		if !ok {
			// Handle error if the uid cannot be asserted to a string
			c.JSON(http.StatusBadRequest, gin.H{"error": "uid is not a string"})
			return
		}


		reports, _ := repo.GetUserReportLinks(uidStr)

		fmt.Println(reports)

		reportsJSON, err := json.Marshal(reports)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to serialize reports")
			return
		}

			fmt.Println("1")
		// Pass the JSON string to the template
		c.HTML(http.StatusOK, "index.html", gin.H{
			"reports": string(reportsJSON), // Ensure it's a valid JSON string
		})

	}
}

func AddUserToReport(authService authentication.AuthServiceInterface, repo repository.ReportRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the email from the request body
		var requestData struct {
			Email   string `json:"email"`
			ReportID string `json:"reportId"`
			Privilege bool `json:"privilege"`

		}


		// Bind the incoming JSON data to the struct
		if err := c.ShouldBindJSON(&requestData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
			return
		}

		// Ensure the email is provided
		if requestData.Email == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
			return
		}

		// Get the UID from the email using the AuthService
		uid, err := authService.GetUIDFromEmail(requestData.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get UID from email, is user registered?"})
			return
		}

		// Add the user to the report (Assuming repo has a method to add a user by UID and ReportID)
		err = repo.LinkReportWithUser(uid, requestData.ReportID, requestData.Privilege, false) //
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add user to report"})
			return
		}

		// Return success response
		fmt.Println("Added User")
		c.JSON(http.StatusOK, gin.H{"message": "User successfully added to report"})
	}
}

func VerifyTokenHandler(authService authentication.AuthServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token is missing"})
			return
		}

		// Remove "Bearer " prefix if present
		token := authHeader[len("Bearer "):]

		// Verify the token
		decodedToken, err := authService.VerifyToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		// Token is valid, return user info (optional)
		c.JSON(http.StatusOK, gin.H{
			"uid":    decodedToken.UID,
			"status": "token is valid",
		})
	}
}


func CreateReportHandler(repo repository.ReportRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ReportRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Success": false, "error": "Invalid Request"})
			return
		}

		reportID := generateID() 

		uID, ok := c.Get("uid")
		if !ok {
			// Handle error if the uid is not found in the context
			c.JSON(http.StatusBadRequest, gin.H{"error": "uid not found"})
			return
		}

		// Assert uID to string
		uidStr, ok := uID.(string)
		if !ok {
			// Handle error if the uid cannot be asserted to a string
			c.JSON(http.StatusBadRequest, gin.H{"error": "uid is not a string"})
			return
		}

		userEmail, ok := c.Get("email")
		if !ok {
			// Handle error if the uid is not found in the context
			c.JSON(http.StatusBadRequest, gin.H{"error": "uid not found"})
			return
		}

		// Assert uID to string
		userEmailStr, ok := userEmail.(string)
		if !ok {
			// Handle error if the uid cannot be asserted to a string
			c.JSON(http.StatusBadRequest, gin.H{"error": "uid is not a string"})
			return
		}



		repo.CreateReport(req.Name, reportID, req.Type, userEmailStr) 
		err := repo.LinkReportWithUser(uidStr, reportID, true, true) // Always give admin to creator 
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to link report with user"})
			return
		}



		c.JSON(http.StatusOK, gin.H{"Success": true, "reportID": reportID})
	}
}

func RegisterHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "register.html", nil)
}

func ReportHandler(repo repository.ReportRepository) gin.HandlerFunc {

	/* Get Sections + Subsections from report or Get template report is based off of */
	return func(c *gin.Context) {	
		fmt.Println("3")
		reportID := c.Param("reportID")
		templateID, _ := repo.GetReportFieldTemplateID(reportID)
		template, _ := repo.GetTemplate(templateID)
		templateJSON, _ := json.Marshal(template.Sections)
		fmt.Println(string(templateJSON)) // Log to ensure it's correct
		c.HTML(http.StatusOK, "report.html", gin.H {
			"template" : string(templateJSON),
		}) 

	}
}

func GenerateReportHandler(repo repository.ReportRepository) gin.HandlerFunc {
  return func(c *gin.Context) {
    reportID := c.DefaultQuery("reportID", "")
    if reportID == "" {
      c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Missing reportID"})
      return
    }

    log.Println("Generating Report:", reportID)

    // Fetch report content
    reportName, reportContent, err := repo.FetchReportContent(reportID)
    if err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to fetch report content"})
      return
    }

    // Generate PDF
    pdfFileName := reportName + ".pdf"
    err = reportGeneration.GeneratePDF(reportName, reportContent)
    if err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("Failed to generate PDF: %v", err)})
      return
    }

    // Set response headers
    c.Header("Content-Disposition", "attachment; filename="+pdfFileName)
    c.Header("Content-Type", "application/pdf")

    // Serve the file
    c.File(pdfFileName)

    // Clean up the file after serving
    go func(fileName string) {
      time.Sleep(1 * time.Second)
      err := os.Remove(fileName)
      if err != nil {
        log.Printf("Failed to delete PDF file: %v", err)
      } else {
        log.Printf("Deleted PDF file: %s", fileName)
      }
    }(pdfFileName)
  }
}

func ReportLogsHandler(repo repository.ReportRepository) gin.HandlerFunc {
  return func(c *gin.Context) {
    reportID := c.Param("reportID")

    logs, err := repo.FetchLogsForReport(reportID)
    if err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch logs"})
      return
    }

    c.HTML(http.StatusOK, "logs.html", gin.H{
      "reportID": reportID,
      "logs":     logs,
    })
  }
}


func RemoveUserFromReport(authService authentication.AuthServiceInterface, repo repository.ReportRepository) gin.HandlerFunc {

	return func(c *gin.Context) {
		reportID := c.Param("reportID")
		// Get the email from the request body
		var requestData struct {
			Email   string `json:"email"`
		}


		// Bind the incoming JSON data to the struct
		if err := c.ShouldBindJSON(&requestData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
			return
		}

		// Ensure the email is provided
		if requestData.Email == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
			return
		}

		// Get the UID from the email using the AuthService
		uid, err := authService.GetUIDFromEmail(requestData.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get UID from email, is user registered?"})
			return
		}

		// Add the user to the report (Assuming repo has a method to add a user by UID and ReportID)

		err = repo.RemoveUserFromReport(uid, reportID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove user from report"})
			return
		}

		// Return success response
		fmt.Println("Removed User")
		c.JSON(http.StatusOK, gin.H{"message": "User successfully removed from report"})

	}
}


func RenameReport(repo repository.ReportRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var requestData struct {
			ReportName   string `json:"reportname"`
		}

		 if err := c.ShouldBindJSON(&requestData); err != nil {
        c.JSON(400, gin.H{"error": "Invalid JSON"})
        return
			} 
		

		reportID := c.Param("reportID")
		err := repo.RenameReport(reportID, requestData.ReportName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to rename report"})
			return
		}

		fmt.Println("Report Renamed to: ", requestData.ReportName)
		c.JSON(http.StatusOK, gin.H{"message": "report was renamed"})	
	}
}


func DeleteReport(repo repository.ReportRepository) gin.HandlerFunc {

	return func(c *gin.Context) {

		reportID := c.Param("reportID")
		err := repo.DeleteReport(reportID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete report"})
			return
		}

		fmt.Println("Removed User")
		c.JSON(http.StatusOK, gin.H{"message": "User successfully removed from report"})	
	}

}


func DeleteAccount(authService authentication.AuthServiceInterface, repo repository.ReportRepository) gin.HandlerFunc {
	return func(c *gin.Context) {

		uID, exists := c.Get("uid")
		if !exists {
			fmt.Println("UID not found in context")
			c.Redirect(http.StatusFound, "/")
			c.Abort()
			return
		}


		uidStr, ok := uID.(string)
		if !ok {
			// Handle error if the uid cannot be asserted to a string
			c.JSON(http.StatusBadRequest, gin.H{"error": "uid is not a string"})
			return
		}



		// Delete user from Firestore
		if err := repo.DestroyUser(uidStr); err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user from Firestore"})
			return
		}

		// Delete user from Firebase Auth
		if err := authService.DestroyUser(uidStr); err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user from Firebase"})
			return
		}


		// Respond with success
		c.JSON(http.StatusOK, gin.H{"message": "User successfully removed"})
		fmt.Println("Removed User")
	}
}


func IsAdmin(repo repository.ReportRepository) gin.HandlerFunc {
  return func(c *gin.Context) {
    reportID := c.Param("reportID")
    uID, exists := c.Get("uid")    
    if !exists {
      fmt.Println("UID not found in context")
      c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
      return
    }

		uidStr, ok := uID.(string)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "uid is not a string"})
			return
		}

		isAdmin, err := repo.IsAdminInReport(uidStr, reportID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check admin status"})
			return
		}

		
		log.Println("is user admin: ", isAdmin)
		c.JSON(http.StatusOK, gin.H{"isAdmin": isAdmin})
	}
}

func WebSocketHandler(repo repository.ReportRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := websockets.Upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println("Error upgrading websocket: ", err)
			return
		}
		defer conn.Close()

		reportID := c.Param("reportID")

		id := c.Request.URL.String()
		fmt.Println(id)

		websocketmanager.OpenConnection(id, conn)
		defer websocketmanager.CloseConnection(id, conn)

		userEmailVal, ok := c.Get("email")
		if !ok {
			log.Println("Email not found in context")
			return
		}
		userEmail, ok := userEmailVal.(string)
		if !ok {
			log.Println("Email in context is not a string")
			return
		}

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println("Error reading message: ", err)
				websocketmanager.CloseConnection(id, conn)
				break
			}

			var message map[string]interface{}
			if err := json.Unmarshal(msg, &message); err != nil {
				log.Println("Error unmarshalling message:", err)
				log.Println("Raw JSON:", string(msg))
				continue
			}

			switch message["type"] {
			case "join":
				log.Println("A client joined: ", id)
				repo.BufferLog(reportID, "joined a report section", userEmail)

				if websocketmanager.GetNumofConns(id) == 1 {
					var joinMsg joinMessage.JoinMessage
					if err := json.Unmarshal(msg, &joinMsg); err != nil {
						log.Println("Error unmarshalling sync message:", err)
						log.Println("Raw JSON:", string(msg))
						continue
					}

					contents, err := repo.FetchReportSectionContents(joinMsg.ReportID, joinMsg.Section)
					if err != nil {
						log.Println("Error fetching contents from database: ", err)
						return
					}

					for _, content := range contents {
						if content == "" {
							continue
						}
						bContent := []byte(content)
						var delta delta.Delta
						if err := json.Unmarshal(bContent, &delta); err != nil {
							log.Println("Error unmarshalling delta:", err)
							log.Println("Raw JSON:", string(bContent))
							continue
						}
						conn.WriteJSON(delta)
					}
				} else {
					websocketmanager.RequestSectionContents(id, conn)
				}

			case "sync":
				log.Println("Received JSON sync message:", string(msg))
				repo.BufferLog(reportID, "synced content", userEmail)

				var syncMsg syncMessage.SyncMessage
				if err := json.Unmarshal(msg, &syncMsg); err != nil {
					log.Println("Error unmarshalling sync message:", err)
					log.Println("Raw JSON:", string(msg))
					continue
				}

				for editorID, content := range syncMsg.Contents {
					websocketmanager.SendToIDExpectConn(id, content, conn)
					deltaJSON, err := json.Marshal(content)
					if err != nil {
						log.Fatal("Error marshalling Delta:", err)
						log.Println("Raw JSON:", string(msg))
					}
					repo.UpdateReportSectionContents(syncMsg.ReportID, syncMsg.Section, editorID, string(deltaJSON))
				}

			case "updateRepo":
				log.Println("Received JSON updateRepo message:", string(msg))
				repo.BufferLog(reportID, "triggered repository update", userEmail)

				var updateRepoMsg updateRepoMessage.UpdateRepoMessage
				if err := json.Unmarshal(msg, &updateRepoMsg); err != nil {
					log.Println("Error unmarshalling updateRepo message:", err)
					log.Println("Raw JSON:", string(msg))
					continue
				}

				for editorID, content := range updateRepoMsg.Contents {
					deltaJSON, err := json.Marshal(content)
					if err != nil {
						log.Fatal("Error marshalling Delta:", err)
						log.Println("Raw JSON:", string(msg))
					}
					repo.UpdateReportSectionContents(updateRepoMsg.ReportID, updateRepoMsg.Section, editorID, string(deltaJSON))
				}

			case "delta":
				log.Println("Received JSON delta message:", string(msg))
				repo.BufferLog(reportID, "sent a delta update: " + string(msg), userEmail)

				var delta delta.Delta
				if err := json.Unmarshal(msg, &delta); err != nil {
					log.Println("Error unmarshalling delta:", err)
					log.Println("Raw JSON:", string(msg))
					continue
				}
				websocketmanager.SendToIDExpectConn(id, delta, conn)

			case "close":
				websocketmanager.CloseConnection(id, conn)
				repo.BufferLog(reportID, "User closed a WebSocket connection", userEmail)
			}
		}
	}
}


