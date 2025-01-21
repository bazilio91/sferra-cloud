package admin

import (
	"github.com/bazilio91/sferra-cloud/pkg/proto"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"log"
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
	if err := initAdminUser(); err != nil {
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

	// Initialize default admin user
	if err := initAdminUser(); err != nil {
		return err
	}

	// Set up routes
	r.GET("/login", ShowLoginPage)
	r.POST("/login", PerformLogin)
	r.GET("/logout", PerformLogout)

	// Protected routes
	authorized := r.Group("/")
	authorized.Use(AuthRequired)
	{
		authorized.GET("/", Dashboard)
		authorized.GET("/clients", ListClients)
		authorized.GET("/clients/new", NewClient)
		authorized.POST("/clients", CreateClient)
		authorized.GET("/clients/:id", ViewClient)
		authorized.GET("/clients/:id/edit", EditClient)
		authorized.POST("/clients/:id", UpdateClient)
		authorized.POST("/clients/:id/delete", DeleteClient)

		authorized.GET("/users", ListUsers)
		authorized.GET("/users/new", NewUser)
		authorized.POST("/users", CreateUser)
		authorized.GET("/users/:id", ViewUser)
		authorized.GET("/users/:id/edit", EditUser)
		authorized.POST("/users/:id", UpdateUser)
		authorized.POST("/users/:id/delete", DeleteUser)
	}

	// Start the server
	return r.Run(":" + cfg.AdminServerPort)
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

	for _, page := range pages {
		// Create template name from file name
		tmplName := filepath.Base(page)
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

func initAdminUser() error {
	var count int64
	db.DB.Model(&proto.Admin{}).Count(&count)
	if count == 0 {
		// Create a default admin user
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		admin := proto.Admin{
			Username: "admin",
			Password: string(hashedPassword),
		}
		if err := db.DB.Create(&admin).Error; err != nil {
			return err
		}
		log.Println("Default admin user created with username 'admin' and password 'admin123'")
	}
	return nil
}
