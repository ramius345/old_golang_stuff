package commoncql

import (
	"container/list"
	"time"

	"github.com/gocql/gocql"
)

func GetRelpaths(session *gocql.Session) ([]string, error) {
	var relpath string
	relpaths := []string{}

	iter := session.Query("select relpath from relpaths").Iter()
	for iter.Scan(&relpath) {
		relpaths = append(relpaths, relpath)
	}

	return relpaths, iter.Close()
}

type RelpathDateEntry struct {
	Relpath string
	Olddate time.Time
	Newdate time.Time
}

func GetRelpathsByDate(session *gocql.Session) (*list.List, error) {
	var entry RelpathDateEntry
	querystring := "select relpath,olddate,newdate from relpaths_by_olddate"
	entries := list.New()

	iter := session.Query(querystring).Iter()
	for iter.Scan(&entry.Relpath, &entry.Olddate, &entry.Newdate) {
		entries.PushBack(entry)
	}

	return entries, iter.Close()
}
