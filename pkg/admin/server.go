package admin

import (
	"github.com/bazilio91/sferra-cloud/pkg/services/storage"
	"html/template"
	"path/filepath"
	"time"

	"github.com/bazilio91/sferra-cloud/pkg/config"
	"github.com/bazilio91/sferra-cloud/pkg/db"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/utrack/gin-csrf"
)

func RunAdminServer() error {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	// Initialize the database
	if err := db.InitDB(cfg); err != nil {
		return err
	}

	// Initialize default admin user
	if err := seed(storage.NewS3Client(cfg)); err != nil {
		return err
	}

	// Initialize the router
	r := gin.Default()

	// Set up sessions
	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("admin_session", store))

	// Use CSRF middleware
	r.Use(csrf.Middleware(csrf.Options{
		Secret: cfg.JWTSecret, // Replace with a secure random string
		ErrorFunc: func(c *gin.Context) {
			c.String(400, "CSRF token mismatch")
			c.Abort()
		},
	}))

	// Set up the template renderer
	r.HTMLRender = createRenderer()

	// Static files
	r.Static("/static", "./static")

	// Set up routes
	setupRoutes(r)

	// Start the server
	return r.Run(":" + cfg.AdminServerPort)
}

func setupRoutes(router *gin.Engine) {
	router.GET("/login", ShowLoginPage)
	router.POST("/login", PerformLogin)

	// Protected routes
	authorized := router.Group("/")
	authorized.Use(AuthRequired)
	{
		authorized.GET("/", Dashboard)
		authorized.GET("/logout", PerformLogout)

		// Client routes
		authorized.GET("/clients", ListClients)
		authorized.GET("/clients/new", NewClient)
		authorized.POST("/clients", CreateClient)
		authorized.GET("/clients/:id", ViewClient)
		authorized.GET("/clients/:id/edit", EditClient)
		authorized.POST("/clients/:id", UpdateClient)
		authorized.POST("/clients/:id/delete", DeleteClient)

		// User routes
		authorized.GET("/users", ListUsers)
		authorized.GET("/users/new", NewUser)
		authorized.POST("/users", CreateUser)
		authorized.GET("/users/:id/edit", EditUser)
		authorized.POST("/users/:id", UpdateUser)
		authorized.POST("/users/:id/delete", DeleteUser)

		// Recognition Task routes
		authorized.GET("/recognition-tasks", ListRecognitionTasks)
		authorized.GET("/recognition-tasks/:id/edit", EditRecognitionTask)
		authorized.POST("/recognition-tasks/:id", UpdateRecognitionTask)
	}
}

// Provided createRenderer function
func createRenderer() multitemplate.Renderer {
	r := multitemplate.NewRenderer()

	// Base templates
	layout := "templates/layouts/layout.html"
	includes, err := filepath.Glob("templates/includes/*.html")
	if err != nil {
		panic(err.Error())
	}

	// Parse templates for each page
	pages, err := filepath.Glob("templates/pages/*.html")
	if err != nil {
		panic(err.Error())
	}
	// Parse templates for subfolders
	subfolders, err := filepath.Glob("templates/pages/**/*.html")
	if err != nil {
		panic(err.Error())
	}

	for _, page := range append(pages, subfolders...) {
		// Create template name from file name
		tmplName, err := filepath.Rel("templates/pages/", page)
		if err != nil {
			panic(err.Error())
		}
		// Combine base layout, includes, and page template
		templates := append([]string{page}, includes...)
		templates = append(templates, layout)

		// Add the template to the renderer
		r.AddFromFilesFuncs(tmplName, template.FuncMap{
			"year": func() int {
				return time.Now().Year()
			},
		}, templates...)
	}

	return r
}
