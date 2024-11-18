// pkg/api/router/router.go
package router

import (
	_ "github.com/bazilio91/sferra-cloud/pkg/api/docs" // Swagger docs
	"github.com/bazilio91/sferra-cloud/pkg/api/handlers"
	"github.com/bazilio91/sferra-cloud/pkg/api/middleware"
	"github.com/bazilio91/sferra-cloud/pkg/auth"
	"github.com/bazilio91/sferra-cloud/pkg/config"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(jwtManager *auth.JWTManager, cfg *config.Config) *gin.Engine {
	router := gin.Default()

	// Initialize middleware with JWT manager
	middleware.SetJWTManager(jwtManager)

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
			// Other protected routes
		}
	}

	return router
}
