package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	
	// Normally we would need a token, but let's just see if the server is up
	resp, err := http.Get("http://localhost:8000/")
	if err != nil {
		fmt.Println("Server is down:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("Server response:", string(body))
}
