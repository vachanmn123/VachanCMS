package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/vachanmn123/vachancms/handlers"
	"github.com/vachanmn123/vachancms/middleware"
)

func SetupRoutes(router gin.IRouter) {
	handlers.InitAuthHandler()

	// Define routes here
	router.GET("/auth/login", handlers.LoginHandler)
	router.GET("/auth/callback", handlers.CallbackHandler)

	protected := router.Group("", middleware.AuthMiddleware)
	protected.GET("/me", handlers.GetMeHandler)
	protected.GET("/repos", handlers.ListRepositoriesHandler)

	repoGroup := protected.Group("/:owner/:repo")
	repoGroup.GET("/config", handlers.GetRepoConfig)
	repoGroup.POST("/init", handlers.InitializeRepo)

	repoGroup.GET("/content-types", handlers.ListContentTypes)
	repoGroup.POST("/content-types", handlers.CreateContentType)
	// Delete and Update will come later, not needed for MVP.

	repoGroup.GET("/:ctSlug", handlers.ListValuesByType)
	repoGroup.POST("/:ctSlug", handlers.CreateValueOfType)
	repoGroup.GET("/:ctSlug/:id", handlers.GetValueById)
	repoGroup.PUT("/:ctSlug/:id", handlers.UpdateValueById)

	repoGroup.GET("/media", handlers.ListMedia)
	repoGroup.POST("/media", handlers.UploadMedia)
	repoGroup.GET("/media/:id", handlers.GetMediaById)

	repoGroup.GET("/pages", handlers.GetPagesConfig)
}
