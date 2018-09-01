package thumbnails

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"pineapple.no-ip.biz/dao"
)

func LookupFullImagePath(cluster *gocql.ClusterConfig, c *gin.Context) {
	hash := c.Param("hash")

	session, _ := cluster.CreateSession()
	defer session.Close()

	path, err := dao.GetImagePathFromHash(session, hash)

	if err != nil {
		c.JSON(500, gin.H{"error": "An error occured looking up the hash"})
		fmt.Printf("%v\n", err)
	} else {
		c.JSON(200, gin.H{"path": path})
	}

}
