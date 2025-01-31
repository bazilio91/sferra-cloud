package admin

import (
	"github.com/bazilio91/sferra-cloud/pkg/db"
	"github.com/bazilio91/sferra-cloud/pkg/proto"
	"github.com/gin-gonic/gin"
	csrf "github.com/utrack/gin-csrf"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

type UserFormInput struct {
	Email    string `form:"email" binding:"required,email"`
	Username string `form:"username" binding:"required"`
	Password string `form:"password" binding:"omitempty,min=6"`
	ClientID uint64 `form:"client_id" binding:"required"`
	Role     string `form:"role" binding:"required"`
}

func ListUsers(c *gin.Context) {
	var users []proto.ClientUserORM

	// GetTaskImage query parameters
	emailFilter := c.Query("email")
	clientNameFilter := c.Query("client_name")

	// Build the query
	query := db.DB.Preload("Client").Model(&proto.ClientUserORM{})
	if emailFilter != "" {
		query = query.Where("client_users.email ILIKE ?", "%"+emailFilter+"%")
	}
	if clientNameFilter != "" {
		query = query.Joins("JOIN clients ON clients.id = client_users.client_id").Where("clients.name ILIKE ?", "%"+clientNameFilter+"%")
	}

	// Execute the query
	if err := query.Find(&users).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "users.html", gin.H{
			"Error": "Failed to retrieve users",
		})
		return
	}

	c.HTML(http.StatusOK, "user/users_list.html", gin.H{
		"Users":            users,
		"EmailFilter":      emailFilter,
		"ClientNameFilter": clientNameFilter,
	})
}

func NewUser(c *gin.Context) {
	// Fetch clients for the client dropdown
	var clients []proto.ClientORM
	db.DB.Find(&clients)
	c.HTML(http.StatusOK, "user/user_new.html", gin.H{
		"Clients":   clients,
		"CsrfToken": csrf.GetToken(c),
	})
}

func CreateUser(c *gin.Context) {
	var input UserFormInput
	if err := c.ShouldBind(&input); err != nil {
		// Fetch clients to repopulate the form
		var clients []proto.ClientORM
		db.DB.Find(&clients)
		c.HTML(http.StatusBadRequest, "user/user_new.html", gin.H{
			"Error":     "Validation error: " + err.Error(),
			"Clients":   clients,
			"CsrfToken": csrf.GetToken(c),
		})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.HTML(http.StatusBadRequest, "user/user_new.html", gin.H{
			"Error":     "Failed to hash password",
			"CsrfToken": csrf.GetToken(c),
		})
		return
	}

	user := proto.ClientUserORM{
		Email:    input.Email,
		Username: input.Username,
		Password: string(hashedPassword),
		ClientId: &input.ClientID,
		Role:     input.Role,
	}

	if err := db.DB.Create(&user).Error; err != nil {
		c.HTML(http.StatusBadRequest, "user/user_new.html", gin.H{
			"Error":     "Failed to create user",
			"CsrfToken": csrf.GetToken(c),
		})
		return
	}
	c.Redirect(http.StatusFound, "/users")
}

func ViewUser(c *gin.Context) {
	id := c.Param("id")
	var user proto.ClientUserORM
	if err := db.DB.Preload("Client").First(&user, id).Error; err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.HTML(http.StatusOK, "user/user_view.html", gin.H{
		"User": user,
	})
}

func EditUser(c *gin.Context) {
	id := c.Param("id")
	var user proto.ClientUserORM
	if err := db.DB.First(&user, id).Error; err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	// Fetch clients for the client dropdown
	var clients []proto.ClientORM
	db.DB.Find(&clients)

	c.HTML(http.StatusOK, "user/user_edit.html", gin.H{
		"User":      user,
		"Clients":   clients,
		"CsrfToken": csrf.GetToken(c),
	})
}

func UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var user proto.ClientUserORM
	if err := db.DB.First(&user, id).Error; err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	var input UserFormInput
	if err := c.ShouldBind(&input); err != nil {
		// Fetch clients to repopulate the form
		var clients []proto.ClientORM
		db.DB.Find(&clients)
		c.HTML(http.StatusBadRequest, "user/user_edit.html", gin.H{
			"Error":     "Validation error: " + err.Error(),
			"User":      user,
			"Clients":   clients,
			"CsrfToken": csrf.GetToken(c),
		})
		return
	}

	user.Email = input.Email
	user.Username = input.Username
	user.ClientId = &input.ClientID
	user.Role = input.Role

	if input.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			c.HTML(http.StatusBadRequest, "user/user_edit.html", gin.H{
				"Error":     "Failed to hash password",
				"User":      user,
				"CsrfToken": csrf.GetToken(c),
			})
			return
		}
		user.Password = string(hashedPassword)
	}

	if err := db.DB.Save(&user).Error; err != nil {
		c.HTML(http.StatusBadRequest, "user/user_edit.html", gin.H{
			"Error":     "Failed to update user",
			"User":      user,
			"CsrfToken": csrf.GetToken(c),
		})
		return
	}
	c.Redirect(http.StatusFound, "/users")
}

func DeleteUser(c *gin.Context) {
	id := c.Param("id")
	var user proto.ClientUserORM
	if err := db.DB.First(&user, id).Error; err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	if err := db.DB.Delete(&user).Error; err != nil {
		c.HTML(http.StatusBadRequest, "user/users_list.html", gin.H{
			"Error": "Failed to delete user",
		})
		return
	}
	c.Redirect(http.StatusFound, "/users")
}
