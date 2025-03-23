package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sema/models/delta"
	"sema/models/joinMessage"
	"sema/models/syncMessage"
	"sema/models/updateRepoMessage"
	"sema/repository"
	"sema/services/reportGeneration"
	"sema/services/websockets"

	"github.com/gin-gonic/gin"
)

var websocketmanager = websockethelper.SpawnWebSocketManager()

/* Put in models later */
type ReportRequest struct { 
	Type string `json:"type" binding:"required"`
	Name string `json:"name" binding:"required"`
}

type ReportGenerationRequest struct {
	Type string `json:"type" binding:"required"`
	ReportID string `json:"reportID" binding:"required"`

}

/* find somewhere else for this */
func generateID() string {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

func HomeHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}


func CreateReportHandler(repo *repository.FirestoreRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ReportRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Success": false, "error": "Invalid Request"})
			return
		}


		reportID := generateID() 

		repo.CreateReport(req.Name, reportID, req.Type) 

		c.JSON(http.StatusOK, gin.H{"Success": true, "reportID": reportID})
	}
}

func ReportHandler(repo *repository.FirestoreRepository) gin.HandlerFunc {

	/* Get Sections + Subsections from report or Get template report is based off of */
	return func(c *gin.Context) {	
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

func GenerateReportHandler(repo *repository.FirestoreRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ReportGenerationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Success": false, "error": "Invalid Request"})
			return
		}


		log.Println("Generating Report: ", req.ReportID);
		log.Println("\n\nReport Contents\n")
		reportName, reportContent, _ := repo.FetchReportContent(req.ReportID)

		reportGeneration.GeneratePDF(reportName, reportContent) 
		log.Println()


	}
}

func WebSocketHandler(repo *repository.FirestoreRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := websockethelper.Upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println("Error upgrading websocket: ", err)
			return 
		}
		defer conn.Close()
		id := c.Request.URL.String() 
		fmt.Println(id)


		websocketmanager.OpenConnection(id, conn)
		defer websocketmanager.CloseConnection(id, conn)

		// Handler here
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println("Error reading message: ", err)
				websocketmanager.CloseConnection(id, conn)
				break
			}

			/* TODO: map to json instead of interface */
			var message map[string]interface{} 
			if err := json.Unmarshal(msg, &message); err !=  nil {
				log.Println("Error unmarshalling message:", err)
				log.Println("Raw JSON:", string(msg)) 
				continue
			}

			if message["type"] =="join" {

				log.Println("A client joined: ", id)

				/*
				Call other connection to update, and update the database
				*/

				/* New webscoket needs to fetch contents from other conn */
				if websocketmanager.GetNumofConns(id) == 1 {
					var joinMsg joinMessage.JoinMessage
					if err := json.Unmarshal(msg, &joinMsg); err != nil {
						log.Println("Error unmarshalling sync message:", err)
						log.Println("Raw JSON:", string(msg)) 
						continue
					}
					// Use fetch from database  
					contents, err := repo.FetchReportSectionContents(joinMsg.ReportID, joinMsg.Section)	
					if err != nil {
						log.Println("Error fetching contents from databse: ", err)
						return 
					}

					fmt.Println("Fetched database contents: ", contents)

					for _, content := range contents {
						if content == "" {
							continue
						}
						bContent := []byte(content)
						var delta delta.Delta
						if err := json.Unmarshal(bContent, &delta); err !=  nil {
							log.Println("Error unmarshalling delta:", err)
							log.Println("Raw JSON:", string(bContent)) 
							continue

						}
						fmt.Println(delta)
						conn.WriteJSON(delta)

					}


				} else {		
					// Fetch random conn and request delta contents foreach editor
					websocketmanager.RequestSectionContents(id, conn)	
				}

			} else if message["type"] == "sync" {
				log.Println("sync")
				log.Println("Received JSON sync message:", string(msg))
				var syncMsg syncMessage.SyncMessage
				if err := json.Unmarshal(msg, &syncMsg); err != nil {
					log.Println("Error unmarshalling sync message:", err)
					log.Println("Raw JSON:", string(msg)) 
					continue
				}

				log.Println("Parsed SyncMessage:", syncMsg)

				for editorID, content := range syncMsg.Contents {

					websocketmanager.SendToIDExpectConn(id, content, conn)	

					// benefit from: goroutine?
					deltaJSON, err := json.Marshal(content)
					if err != nil {
						log.Fatal("Error marshalling Delta:", err)
						log.Println("Raw JSON:", string(msg)) 
					}

					err = repo.UpdateReportSectionContents(syncMsg.ReportID, syncMsg.Section, editorID, string(deltaJSON))
					if err != nil {
						log.Println("Failed to update DB subsection:", err)
					}

					log.Printf("Editor %s received sync data: %+v", editorID, content)	
				}	


				// Update repo as well

			} else if message["type"] == "updateRepo" {
				log.Println("repo")
				log.Println("Received JSON sync message:", string(msg))
				var updateRepoMsg updateRepoMessage.UpdateRepoMessage 
				if err := json.Unmarshal(msg, &updateRepoMsg); err != nil {
					log.Println("Error unmarshalling sync message:", err)
					log.Println("Raw JSON:", string(msg)) 
					continue
				}

				log.Println("Parsed SyncMessage:", updateRepoMsg)

				for editorID, content := range updateRepoMsg.Contents {


					deltaJSON, err := json.Marshal(content)
					if err != nil {
						log.Fatal("Error marshalling Delta:", err)
						log.Println("Raw JSON:", string(msg)) 
					}

					err = repo.UpdateReportSectionContents(updateRepoMsg.ReportID, updateRepoMsg.Section, editorID, string(deltaJSON))
					if err != nil {
						log.Println("Failed to update DB subsection:", err)
					}

					log.Printf("Editor %s received sync data: %+v", editorID, content)	
				}	

			} else if message["type"] == "delta" {
				log.Println("Received JSON sync message:", string(msg))
				var delta delta.Delta
				if err := json.Unmarshal(msg, &delta); err !=  nil {
					log.Println("Error unmarshalling delta:", err)
					log.Println("Raw JSON:", string(msg)) 
					continue
				}
				log.Println("Recevied message: ", message)
				log.Println("Recevied delta: ",  delta)

				websocketmanager.SendToIDExpectConn(id, delta, conn)

			} else if message["type"] == "close" {
				websocketmanager.CloseConnection(id ,conn)
				log.Println("Client has closed Socket")
				break
			}

		}
	}
}
