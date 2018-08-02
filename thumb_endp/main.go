package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"pineapple.no-ip.biz/thumb_endp/thumbnails"
)

func fromEnv(variable string, defaultValue string) string {
	if value := os.Getenv(variable); value != "" {
		return value
	} else {
		return defaultValue
	}
}

func main() {
	listen_port := fromEnv("LISTEN_ADDR", "0.0.0.0:5555")
	cluster_name := fromEnv("CLUSTER_NAME", "greengrape")
	cluster_port := fromEnv("CLUSTER_PORT", "30000")
	keyspace := fromEnv("KEYSPACE", "imageapp")

	cluster_port_int, err := strconv.ParseInt(cluster_port, 10, 32)
	if err != nil {
		fmt.Println("Error reading cluster port!")
		os.Exit(1)
	}

	fmt.Println("Connecting to cluster " + cluster_name + " on port " + cluster_port)
	fmt.Println("Using keyspace " + keyspace)

	cluster := gocql.NewCluster(cluster_name)
	cluster.Port = int(cluster_port_int)
	cluster.Keyspace = keyspace
	cluster.Consistency = gocql.Quorum

	r := gin.Default()
	//get thumbnails in newest first order
	r.GET("/thumbnails", func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		thumbnails.Thumbnails(cluster, c)
	})

	r.GET("/rthumbnails", func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		thumbnails.RThumbnails(cluster, c)
	})

	r.Run(listen_port)
}
