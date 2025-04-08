package middleware

import (
	"fmt"
	"net/http"

	"sema/services/authentication"
	"sema/repository"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware verifies Firebase JWT tokens from cookies
func AuthMiddleware(authService *authentication.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("firebaseToken")
		if err != nil || token == "" {
			fmt.Println("Missing token, redirecting to register")
			c.Redirect(http.StatusFound, "/register")
			c.Abort()
			return
		}

		decodedToken, err := authService.VerifyToken(token)
		if err != nil {
			fmt.Println("Token verification error, redirecting to register")
			c.Redirect(http.StatusFound, "/register")
			c.Abort()
			return
		}

		user, err := authService.GetUserByUID(decodedToken.UID)
		if err != nil {
			fmt.Println("Error fetching user details: ", err)
			c.Redirect(http.StatusFound, "/register")
			c.Abort()
			return
		}

		c.Set("email", user.Email)
		c.Set("uid", user.UID)
		c.Next()
	}
}

func AuthUserinReport(authService *authentication.AuthService, repo repository.ReportRepository) gin.HandlerFunc {
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "uid is not a string"})
			return
		}

		reportID := c.Param("reportID")
		if yes, err := repo.IsUserInReport(uidStr, reportID); !yes {
			fmt.Println("Error user not in report: ", err)
			c.Redirect(http.StatusFound, "/")
			c.Abort()
			return
		}

		c.Next()
	}
}

func AuthAdmininReport(authService *authentication.AuthService, repo repository.ReportRepository) gin.HandlerFunc {
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "uid is not a string"})
			return
		}

		reportID := c.Param("reportID")
		if yes, err := repo.IsAdminInReport(uidStr, reportID); !yes {
			fmt.Println("Error not admin in report: ", err)
			c.Redirect(http.StatusFound, "/")
			c.Abort()
			return
		}

		c.Next()
	}
}

