// pkg/api/router/router.go
package router

import (
	_ "github.com/bazilio91/sferra-cloud/pkg/api/docs" // Swagger docs
	"github.com/bazilio91/sferra-cloud/pkg/api/handlers"
	"github.com/bazilio91/sferra-cloud/pkg/api/middleware"
	"github.com/bazilio91/sferra-cloud/pkg/auth"
	"github.com/bazilio91/sferra-cloud/pkg/config"
	"github.com/bazilio91/sferra-cloud/pkg/services/storage"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(jwtManager *auth.JWTManager, cfg *config.Config) *gin.Engine {
	router := gin.Default()

	// Initialize middleware with JWT manager
	middleware.SetJWTManager(jwtManager)

	// Initialize S3 client and image handler
	s3Client := storage.NewS3Client(cfg)
	imageHandler := handlers.NewImageHandler(s3Client)

	// Swagger endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := router.Group("/api/v1")
	{
		// Public routes
		api.POST("/auth/login", handlers.Login)
		api.POST("/auth/register", handlers.Register)

		// Protected routes
		apiAuth := api.Group("/")
		apiAuth.Use(middleware.JWTAuthMiddleware())
		{
			apiAuth.GET("/account", handlers.GetAccountInfo)

			// Data Recognition Task routes
			apiAuth.POST("/recognition_tasks", handlers.CreateDataRecognitionTask)
			apiAuth.GET("/recognition_tasks", handlers.ListDataRecognitionTask)
			apiAuth.GET("/recognition_tasks/:id", handlers.GetDataRecognitionTask)
			apiAuth.PUT("/recognition_tasks/:id", handlers.UpdateDataRecognitionTask)
			apiAuth.DELETE("/recognition_tasks/:id", handlers.DeleteDataRecognitionTask)

			// Image routes
			apiAuth.POST("/images/upload", imageHandler.UploadImage)
			apiAuth.GET("/images/:id", imageHandler.GetImage)
		}
	}

	return router
}
