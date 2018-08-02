package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gocql/gocql"
)

func scanDirectoryJpegs(path string) []string {
	var jpegs []string
	pathlen := len(path) + 1
	fmt.Println("Doing walk")
	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() == true &&
			(strings.HasSuffix(strings.ToLower(path), "jpg") ||
				strings.HasSuffix(strings.ToLower(path), "jpeg")) {
			subpath := string(path[pathlen:])
			jpegs = append(jpegs, subpath)
		}
		return nil
	})

	return jpegs
}

type Shadata struct {
	data [sha256.Size]byte
}

type ShaHandler func(prefix string, subpath string, sha Shadata)

func computeShas(prefix string, subpaths []string, handler ShaHandler) {
	subpath_length := float(len(subpaths))

	for i, subpath := range subpaths {
		filename := prefix + "/" + subpath
		fmt.Println("Opening " + filename)
		file, openerr := os.Open(filename)
		if openerr != nil {
			fmt.Println("Error opening file " + filename)
			continue
		}
		fmt.Println("Reading " + filename)
		data, readerr := ioutil.ReadAll(file)
		if readerr != nil {
			fmt.Println("Error reading file " + filename)
			continue
		}

		fmt.Println("Computing sum of file " + filename)
		sumbytes := sha256.Sum256(data)
		shadata := Shadata{sumbytes}
		handler(prefix, subpath, shadata)
		fmt.Println("Completed %f%% of shas\n", float(i)/subpath_length)
	}
}

func getDatabaseWriter(session *gocql.Session) ShaHandler {
	writerfunc := func(prefix string, subpath string, sha Shadata) {
		shastring := getShaString(sha)
		info, staterr := os.Stat(prefix + "/" + subpath)
		if staterr != nil {
			return
		}

		modtime := info.ModTime()
		dirpart := filepath.Dir(subpath)
		filename := filepath.Base(subpath)

		fmt.Println("Inserting relpath entry for " + subpath)
		queryerr := session.Query("INSERT INTO images_by_relpath (relpath,filename,imagedate,sha,insertdate) VALUES (?,?,?,?,?)",
			dirpart, filename, modtime, shastring, time.Now()).Exec()

		if queryerr == nil {
			fmt.Println("Inserted relpath entry for " + subpath)
		} else {
			fmt.Printf("%s\n", queryerr)
		}

		fmt.Println("Inserting sha entry for " + subpath)
		queryerr = session.Query("INSERT INTO images_by_sha (relpath,filename,imagedate,sha,insertdate) VALUES (?,?,?,?,?)",
			dirpart, filename, modtime, shastring, time.Now()).Exec()

		if queryerr == nil {
			fmt.Println("Inserted sha entry for " + subpath)
		} else {
			fmt.Printf("%s\n", queryerr)
		}
	}
	return writerfunc
}

func printShas(prefix string, subpath string, sha Shadata) {
	dirpart := filepath.Dir(subpath)
	filename := filepath.Base(subpath)

	fmt.Println("dirpart: " + dirpart + " filename: " + filename)
	fmt.Println("Sha for subpath " + subpath + " is:\n " + getShaString(sha))
}

func getShaString(sha Shadata) string {
	var buffer bytes.Buffer
	for _, b := range sha.data[:] {
		buffer.WriteString(fmt.Sprintf("%02x", b))
	}
	return buffer.String()
}

func main() {
	fmt.Println("Connecting to database")
	cluster := gocql.NewCluster("greengrape")
	cluster.Port = 30000
	cluster.Keyspace = "imageapp"
	cluster.Consistency = gocql.Quorum
	session, _ := cluster.CreateSession()
	defer session.Close()

	fmt.Println("Scanning jpegs")
	directory := "/mnt/san/Pictures/Carolyn"
	jpegs := scanDirectoryJpegs(directory)

	jpegCount := len(jpegs)
	fmt.Printf("Found %d jpeg images\n", jpegCount)

	computeShas(directory, jpegs, getDatabaseWriter(session))
}
