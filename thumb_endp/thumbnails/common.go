package thumbnails

import (
	"container/list"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"pineapple.no-ip.biz/dao"
)

//get the t query parameter as a date - default now
//get the c query parameter as an int - default 32
func getQueryParameters(c *gin.Context, defaultTime time.Time) (time.Time, int, error) {
	fromDateString := c.DefaultQuery("t", defaultTime.Format(time.RFC3339))
	countString := c.DefaultQuery("c", "32")
	fromDate, dateErr := time.Parse(time.RFC3339, fromDateString)
	if dateErr != nil {
		return time.Time{}, 0, dateErr
	}
	parsedCount, countError := strconv.ParseInt(countString, 10, 32)
	count := int(parsedCount)
	if countError != nil {
		return time.Time{}, 0, countError
	}

	return fromDate, count, nil
}

type TimeCompare func(listday time.Time, reqtime time.Time) bool

//locate the starting date.  Use the specified function to find the right partition
//to start in
// days - the list of "days" partitions
func findStartPartitionFromList(days *list.List, fromDate time.Time, compare TimeCompare) *list.Element {
	var day *list.Element
	for day = days.Front(); day != nil && compare(day.Value.(time.Time), fromDate); day = day.Next() {
	}
	return day
}

//daylist - list of day partitions forwarded to the starting partition (use findStartPartitionfromList)
//fromDate - the mark date to get images from before this time
//count - amount of images to get
func findNImagesStartingAtDay(session *gocql.Session, dayList *list.Element,
	fromDate time.Time, count int,
	queryfunc dao.ImageQueryFunc) ([]dao.ImageDatePair, error) {
	jpegs := []dao.ImageDatePair{}
	for ; dayList != nil && len(jpegs) < count; dayList = dayList.Next() {
		var shas []dao.ImageDatePair
		var queryerr error

		shas, queryerr = queryfunc(session, dayList.Value.(time.Time), fromDate, int(count))
		if queryerr != nil {
			return []dao.ImageDatePair{}, queryerr
		}

		jpegs = append(jpegs, shas...)
	}
	return jpegs, nil
}

func findNImagesStartingAtDayDesc(session *gocql.Session,
	dayList *list.Element,
	fromDate time.Time,
	count int) ([]dao.ImageDatePair, error) {
	return findNImagesStartingAtDay(session, dayList, fromDate, count, dao.GetImagesBeforeTime)
}

func findNImagesStartingAtDayAsc(session *gocql.Session,
	dayList *list.Element,
	fromDate time.Time,
	count int) ([]dao.ImageDatePair, error) {
	return findNImagesStartingAtDay(session, dayList, fromDate, count, dao.GetImagesAfterTime)
}
