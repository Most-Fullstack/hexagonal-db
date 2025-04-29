package http

import (
	"hexgonaldb/internal/app/service"
	"hexgonaldb/internal/domain"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RunServer(svc *service.Service) {
	r := gin.Default()

	r.POST("/register", func(c *gin.Context) {
		var report domain.Report
		if err := c.ShouldBindJSON(&report); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
	})

	r.Run(":8080")
}
