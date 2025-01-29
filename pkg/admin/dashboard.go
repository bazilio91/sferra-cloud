package admin

import (
	"github.com/bazilio91/sferra-cloud/pkg/proto"
	"net/http"

	"github.com/bazilio91/sferra-cloud/pkg/db"
	"github.com/gin-gonic/gin"
)

func Dashboard(c *gin.Context) {
	var clientCount int64
	var userCount int64

	db.DB.Model(&proto.ClientORM{}).Count(&clientCount)
	db.DB.Model(&proto.ClientUserORM{}).Count(&userCount)

	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"ClientCount": clientCount,
		"UserCount":   userCount,
	})
}
