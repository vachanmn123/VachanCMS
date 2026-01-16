package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vachanmn123/vachancms/services"
)

type PagesConfigResponse struct {
	Initialized bool   `json:"initialized"`
	BaseURL     string `json:"baseUrl,omitempty"`
}

func GetPagesConfig(c *gin.Context) {
	owner := c.Param("owner")
	repo := c.Param("repo")
	access_token := c.GetString("user_access_token")

	pagesConfig, err := services.GetPagesConfig(access_token, owner, repo)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "An error occured"})
	}

	if pagesConfig.Initialized {
		c.JSON(http.StatusOK, PagesConfigResponse{
			Initialized: pagesConfig.Initialized,
			BaseURL:     pagesConfig.URL,
		})
		return
	} else {
		c.JSON(http.StatusOK, PagesConfigResponse{
			Initialized: pagesConfig.Initialized,
		})
	}
}
