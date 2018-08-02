package main

import "github.com/gin-gonic/gin"

func main() {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.Data(200, "text/plain", []byte("<html><head><title>Foo</title></head><body>blah</body></html>"))
	})

	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "test",
			"submsg":  gin.H{"a": "b"},
		})
	})
	r.Run("0.0.0.0:5555")
}
