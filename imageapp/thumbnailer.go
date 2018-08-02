package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"os"

	"github.com/gocql/gocql"
	"github.com/nfnt/resize"
)

func getDbIterator(session *gocql.Session) *gocql.Iter {
	return session.Query("SELECT sha,relpath,filename from images_by_sha").PageSize(128).Iter()
}

func targetPathDoesntExist(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return true
	}
	return false
}

func readSourceImage(from_path string) (image.Image, error) {
	file, err := os.Open(from_path)
	defer file.Close()
	if err != nil {
		return nil, err
	}

	img, err := jpeg.Decode(file)
	if err != nil {
		return nil, err
	}

	return img, nil
}

func scaleImage(img image.Image, widthInPixels uint) image.Image {
	return resize.Resize(widthInPixels, 0, img, resize.Lanczos3)
}

func writeThumbnail(thumbPath string, img image.Image) error {
	out, err := os.Create(thumbPath)
	if err != nil {
		return err
	}
	defer out.Close()

	options := &jpeg.Options{Quality: 100}

	err = jpeg.Encode(out, img, options)
	return err
}

func main() {
	path_to_images := "/mnt/san/Pictures/Carolyn"
	path_to_thumbnails := "/mnt/san/thumbnails"
	var widthInPixels uint = 300

	fmt.Println("Connecting to database")
	cluster := gocql.NewCluster("greengrape")
	cluster.Port = 30000
	cluster.Keyspace = "imageapp"
	cluster.Consistency = gocql.Quorum
	session, _ := cluster.CreateSession()
	defer session.Close()
	fmt.Println("Connected.")

	//get an iterator
	iter := getDbIterator(session)

	var sha, relpath, filename string

	for i := 0; iter.Scan(&sha, &relpath, &filename); i++ {
		fmt.Printf("%d: Got relpath: %s, filename: %s, sha: %s\n",
			i, relpath, filename, sha)

		full_thumb_path := path_to_thumbnails + "/" + sha + ".jpg"
		fmt.Printf("Would have genrated: %s\n", full_thumb_path)
		from_path := path_to_images + "/" + relpath + "/" + filename
		fmt.Printf("From: %s\n", from_path)

		//check target path exists
		if targetPathDoesntExist(full_thumb_path) {
			//if it does, then load the image
			img, err := readSourceImage(from_path)
			if err != nil {
				fmt.Println("Failed reading image from source path!")
				continue
			}

			fmt.Printf("Decoded %s\n", from_path)
			//scale to a thumbnail
			scaled := scaleImage(img, widthInPixels)

			fmt.Printf("Image scaled to %d px", widthInPixels)

			//write the image
			if err := writeThumbnail(full_thumb_path, scaled); err != nil {
				fmt.Printf("Failed to write thumb image %s\n", full_thumb_path)
				continue
			}

		}
	}

}
