package admin

import (
	"github.com/bazilio91/sferra-cloud/pkg/proto"
	csrf "github.com/utrack/gin-csrf"
	"net/http"

	"github.com/bazilio91/sferra-cloud/pkg/db"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func ShowLoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{
		"CsrfToken": csrf.GetToken(c),
	})
}

func PerformLogin(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	var admin proto.Admin
	if err := db.DB.Where("username = ?", username).First(&admin).Error; err != nil {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{
			"Error": "Invalid username or password",
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password)); err != nil {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{
			"Error": "Invalid username or password",
		})
		return
	}

	// Save admin ID in session
	session := sessions.Default(c)
	session.Set("admin_id", admin.Id)
	session.Save()

	c.Redirect(http.StatusFound, "/")
}

func PerformLogout(c *gin.Context) {
	session := sessions.Default(c)
	session.Delete("admin_id")
	session.Save()
	c.Redirect(http.StatusFound, "/login")
}
