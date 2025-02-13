package admin

import (
	"github.com/bazilio91/sferra-cloud/pkg/db"
	"github.com/bazilio91/sferra-cloud/pkg/proto"
	"github.com/gin-gonic/gin"
	csrf "github.com/utrack/gin-csrf"
	"net/http"
)

type ClientFormInput struct {
	Name       string `form:"name" binding:"required,min=3,max=100"`
	Quota      int64  `form:"quota" binding:"required,gte=0"`
	TotalQuota int64  `form:"total_quota" binding:"required,gte=0"`
	OwnerFio   string `form:"owner_fio" binding:"required"`
	Inn        string `form:"inn" binding:"required"`
	Ogrn       string `form:"ogrn" binding:"required"`
}

func ListClients(c *gin.Context) {
	var clients []proto.ClientORM

	// GetTaskImage query parameters
	nameFilter := c.Query("name")
	idFilter := c.Query("id")

	// Build the query
	query := db.DB.Model(&proto.ClientORM{})
	if nameFilter != "" {
		query = query.Where("name ILIKE ?", "%"+nameFilter+"%")
	}
	if idFilter != "" {
		query = query.Where("id = ?", idFilter)
	}

	// Execute the query
	if err := query.Find(&clients).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "clients.html", gin.H{
			"Error": "Failed to retrieve clients",
		})
		return
	}

	c.HTML(http.StatusOK, "client/clients_list.html", gin.H{
		"Clients":    clients,
		"NameFilter": nameFilter,
		"IDFilter":   idFilter,
	})
}

func NewClient(c *gin.Context) {
	c.HTML(http.StatusOK, "client/client_new.html", gin.H{
		"CsrfToken": csrf.GetToken(c),
	})
}

func CreateClient(c *gin.Context) {
	var input ClientFormInput
	if err := c.ShouldBind(&input); err != nil {
		c.HTML(http.StatusBadRequest, "client/client_new.html", gin.H{
			"Error":     "Validation error: " + err.Error(),
			"CsrfToken": csrf.GetToken(c),
		})
		return
	}

	client := proto.ClientORM{
		Name:       input.Name,
		Quota:      input.Quota,
		TotalQuota: input.TotalQuota,
		OwnerFio:   input.OwnerFio,
		Inn:        input.Inn,
		Ogrn:       input.Ogrn,
	}
	if err := db.DB.Create(&client).Error; err != nil {
		c.HTML(http.StatusBadRequest, "client/client_new.html", gin.H{
			"Error":     "Failed to create client",
			"CsrfToken": csrf.GetToken(c),
		})
		return
	}
	c.Redirect(http.StatusFound, "/clients")
}

func ViewClient(c *gin.Context) {
	id := c.Param("id")
	var client proto.ClientORM
	if err := db.DB.First(&client, id).Error; err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.HTML(http.StatusOK, "client/client_view.html", gin.H{
		"Client": client,
	})
}

func EditClient(c *gin.Context) {
	id := c.Param("id")
	var client proto.ClientORM
	if err := db.DB.First(&client, id).Error; err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.HTML(http.StatusOK, "client/client_edit.html", gin.H{
		"Client":    client,
		"CsrfToken": csrf.GetToken(c),
	})
}

func UpdateClient(c *gin.Context) {
	id := c.Param("id")
	var client proto.ClientORM
	if err := db.DB.First(&client, id).Error; err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	var input ClientFormInput
	if err := c.ShouldBind(&input); err != nil {
		c.HTML(http.StatusBadRequest, "client/client_edit.html", gin.H{
			"Error":     "Validation error: " + err.Error(),
			"Client":    client,
			"CsrfToken": csrf.GetToken(c),
		})
		return
	}

	client.Name = input.Name
	client.Quota = input.Quota
	client.TotalQuota = input.TotalQuota
	client.OwnerFio = input.OwnerFio
	client.Inn = input.Inn
	client.Ogrn = input.Ogrn

	if err := db.DB.Save(&client).Error; err != nil {
		c.HTML(http.StatusBadRequest, "client/client_edit.html", gin.H{
			"Error":     "Failed to update client",
			"Client":    client,
			"CsrfToken": csrf.GetToken(c),
		})
		return
	}
	c.Redirect(http.StatusFound, "/clients")
}

func DeleteClient(c *gin.Context) {
	id := c.Param("id")
	var client proto.ClientORM
	if err := db.DB.First(&client, id).Error; err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	if err := db.DB.Delete(&client).Error; err != nil {
		c.HTML(http.StatusBadRequest, "client/clients.html", gin.H{
			"Error": "Failed to delete client",
		})
		return
	}
	c.Redirect(http.StatusFound, "/clients")
}
