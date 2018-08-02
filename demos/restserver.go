package main

import "github.com/gin-gonic/gin"

func main() {
	r := gin.Default()
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "test",
			"submsg":  gin.H{"a": "b"},
		})
	})
	r.Run("0.0.0.0:5555")
}
