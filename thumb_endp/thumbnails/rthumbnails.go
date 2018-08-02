package thumbnails

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"pineapple.no-ip.biz/dao"
)

//search partition function for items in ascending order
func ascendingCompare(listday time.Time, reqtime time.Time) bool {
	return listday.Before(reqtime)
}

func RThumbnails(cluster *gocql.ClusterConfig, c *gin.Context) {
	fromDate, count, paramErrors := getQueryParameters(c, time.Time{})
	if paramErrors != nil {
		c.JSON(500, paramErrors.Error())
	}

	session, _ := cluster.CreateSession()
	defer session.Close()

	jpegs := []dao.ImageDatePair{}

	//get a list of the days
	days, queryerr := dao.GetImageDaysReverse(session)
	if queryerr != nil {
		c.JSON(500, gin.H{"error": "An error occured making a days query"})
		fmt.Printf("%v\n", queryerr)
	}

	day := findStartPartitionFromList(days, fromDate, ascendingCompare)
	if day == nil {
		c.JSON(404, gin.H{"error": "Out of range"})
		return
	}

	jpegs, queryerr = findNImagesStartingAtDayAsc(session, day, fromDate, count)
	if queryerr != nil {
		c.JSON(500, gin.H{"error": "An error occured getting shas"})
		return
	}

	c.JSON(200, gin.H{
		"jpegs": jpegs,
	})

}
