package client

import "github.com/gin-gonic/gin"

// {filePath}
func downloadFile(c *gin.Context) {
	filePath := c.GetHeader("filePath")
	c.File(filePath)
}
