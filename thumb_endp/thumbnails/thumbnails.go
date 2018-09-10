package thumbnails

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"pineapple.no-ip.biz/dao"
)

//search partition function for items in descending order
func descendingTimeCompare(listday time.Time, reqtime time.Time) bool {
	return listday.After(reqtime)
}

func Thumbnails(cluster *gocql.ClusterConfig, c *gin.Context) {
	fromDate, count, paramErrors := getQueryParameters(c, time.Now())
	if paramErrors != nil {
		c.JSON(500, gin.H{"error": paramErrors.Error()})
		return
	}
	session, err := cluster.CreateSession()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer session.Close()

	jpegs := []dao.ImageDatePair{}

	//get a list of the days
	days, queryerr := dao.GetImageDays(session)
	if queryerr != nil {
		c.JSON(500, gin.H{"error": "An error occured making a days query"})
		fmt.Printf("error: %v\n", queryerr)
		return
	}

	day := findStartPartitionFromList(days, fromDate, descendingTimeCompare)
	if day == nil {
		c.JSON(404, gin.H{"error": "Out of range"})
		return
	}

	jpegs, queryerr = findNImagesStartingAtDayDesc(session, day, fromDate, count)
	if queryerr != nil {
		c.JSON(500, gin.H{"error": "An error occured getting shas"})
		return
	}

	c.JSON(200, gin.H{"jpegs": jpegs})
}
