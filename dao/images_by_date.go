package dao

import (
	"container/list"
	"time"

	"github.com/gocql/gocql"
)

func getImageDays(session *gocql.Session, querystring string) (*list.List, error) {
	var day time.Time
	entries := list.New()
	iter := session.Query(querystring).Iter()
	for iter.Scan(&day) {
		entries.PushBack(day)
	}
	return entries, iter.Close()
}

func GetImageDays(session *gocql.Session) (*list.List, error) {
	querystring := "select day from image_days where force=1 order by day desc"
	return getImageDays(session, querystring)
}

func GetImageDaysReverse(session *gocql.Session) (*list.List, error) {
	querystring := "select day from image_days where force=1 order by day asc"
	return getImageDays(session, querystring)
}

type ImageDatePair struct {
	Sha  string    `json:"sha"`
	Date time.Time `json:"date"`
}

func getImages(session *gocql.Session, day time.Time, marktime time.Time, max int, querystring string) ([]ImageDatePair, error) {
	iter := session.Query(querystring, day, marktime, max).Iter()
	var sha string
	var date time.Time
	shas := []ImageDatePair{}
	for iter.Scan(&sha, &date) {
		datepair := ImageDatePair{sha, date}
		shas = append(shas, datepair)
	}
	return shas, iter.Close()

}

type ImageQueryFunc func(session *gocql.Session, day time.Time, marktime time.Time, max int) ([]ImageDatePair, error)

func GetImagesBeforeTime(session *gocql.Session, day time.Time, before time.Time, max int) ([]ImageDatePair, error) {
	querystring := "select sha,imagedate from images_by_date where day=? and imagedate<? order by imagedate desc limit ? "
	return getImages(session, day, before, max, querystring)
}

func GetImagesAfterTime(session *gocql.Session, day time.Time, after time.Time, max int) ([]ImageDatePair, error) {
	querystring := "select sha,imagedate from images_by_date where day=? and imagedate>? order by imagedate asc limit ?"
	return getImages(session, day, after, max, querystring)
}
