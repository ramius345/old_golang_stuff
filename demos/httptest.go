package main

import (
	"fmt"
	"net/http"
)

func main() {
	client := http.Client{}
	req, _ := http.NewRequest("GET", "http://localhost:9000", nil)
	req.Header.Set("X-Vault-Check", "blah")
	req.Header.Set("Accept", "*/*")
	_, err := client.Do(req)
	fmt.Println(err)
}
