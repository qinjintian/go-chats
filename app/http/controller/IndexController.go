package controller

import (
	"github.com/gin-gonic/gin"
)

type IndexController struct {}

func (i *IndexController) Index(c *gin.Context) {
	c.HTML(200, "index.html", gin.H{
		"title": "Index site",
	})
}
