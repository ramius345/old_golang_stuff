package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/gocql/gocql"
)

type Entry struct {
	Filepath  string
	Sha       string
	Imagedate time.Time
	Terminate bool
}

func doesFileExist(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func removeByShaEntry(entry Entry, batch *gocql.Batch) {
	batch.Query("delete from images_by_sha where sha=?", entry.Sha)
}

func determineImageDay(imagedate time.Time) time.Time {
	day := imagedate.Day()
	month := imagedate.Month()
	year := imagedate.Year()
	location := imagedate.Location()

	dayonly := time.Date(year, month, day, 0, 0, 0, 0, location)
	return dayonly
}

func removeByDateEntry(entry Entry, batch *gocql.Batch) {
	day := determineImageDay(entry.Imagedate)
	fmt.Printf("Doing lookup on %v,%v,%v\n", day, entry.Imagedate, entry.Sha)
	batch.Query("delete from images_by_date where day=? and imagedate=? and sha=?",
		day,
		entry.Imagedate,
		entry.Sha)
}

func removeThumbnailForEntry(entry Entry, thumbnailPath string) error {
	fullPath := thumbnailPath + "/" + entry.Sha + ".jpg"
	fmt.Printf("Checking if thumbnail path exists for %s\n", fullPath)
	if doesFileExist(fullPath) {
		fmt.Printf("Removing thumbnail %s\n", fullPath)
		return os.Remove(fullPath)
	} else {
		fmt.Printf("Thumbnail already gone %s\n", fullPath)
		return nil
	}
}

func scheduleCleanupThread(entryCleanChannel chan Entry, session *gocql.Session, thumbnailPath string) {
	finished := false
	for !finished {
		entry := <-entryCleanChannel
		if !entry.Terminate {
			fmt.Printf("Cleaning up entry for %s %s %v\n", entry.Filepath, entry.Sha, entry.Imagedate)

			batch := session.NewBatch(gocql.LoggedBatch)
			removeByShaEntry(entry, batch)
			removeByDateEntry(entry, batch)
			err := session.ExecuteBatch(batch)
			if err != nil {
				fmt.Printf("Executing remove batch failed! %v\n", err)
			}

			err = removeThumbnailForEntry(entry, thumbnailPath)
			if err != nil {
				fmt.Printf("Executing remove thumbnail failed: %v\n", err)
			}

		}
		finished = entry.Terminate
	}
}

func verifyImagesThread(filePathChannel chan Entry, cleanupChannel chan Entry) {
	finished := false
	i := 0
	for !finished {
		entry := <-filePathChannel
		if (!entry.Terminate && !doesFileExist(entry.Filepath)) || entry.Terminate {
			cleanupChannel <- entry
		}

		i += 1
		finished = entry.Terminate
	}
}

func checkFilesystemMounted(path string) {
	fmt.Println("Ensuring that there are at least 10 images in the mount")
	files, _ := ioutil.ReadDir(path)
	filecount := len(files)
	if filecount < 10 {
		fmt.Println("Refusing to continue because there are less than 10 files")
		os.Exit(1)
	}
}

func main() {

	db_url := "winredgrape.pineapple.no-ip.biz"
	db_port := 30000
	path_to_images := "/mnt/san/Pictures/Carolyn"
	path_to_thumbnails := "/mnt/san/thumbnails"

	checkFilesystemMounted(path_to_images)

	fmt.Println("Connecting to database")
	cluster := gocql.NewCluster(db_url)
	cluster.Port = db_port
	cluster.Keyspace = "imageapp"
	cluster.Consistency = gocql.Quorum
	session, _ := cluster.CreateSession()
	defer session.Close()

	iter := session.Query("select sha,relpath,filename,imagedate from images_by_sha").Iter()

	filePathChannel := make(chan Entry, 128)
	cleanUpChannel := make(chan Entry, 128)
	go verifyImagesThread(filePathChannel, cleanUpChannel)
	cleanupThreadSession, _ := cluster.CreateSession()
	go scheduleCleanupThread(cleanUpChannel, cleanupThreadSession, path_to_thumbnails)

	var sha, relpath, filename string
	var imagedate time.Time
	i := 0
	for iter.Scan(&sha, &relpath, &filename, &imagedate) {
		filePathChannel <- Entry{path_to_images + "/" + relpath + "/" + filename, sha, imagedate, false}
		i += 1
	}
	fmt.Println("Finished scanning database for image entries, signaling")
	fmt.Printf("Processed %d db entries\n", i)
	filePathChannel <- Entry{"", "", time.Time{}, false}

	err := iter.Close()
	if err != nil {
		fmt.Printf("An error occured querying images by sha %v\n", err)
	}
}
