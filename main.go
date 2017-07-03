package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/otiai10/fastpot-api/app"
)

// Get API Server Up!!
func main() {

	port := func() string {
		if p := os.Getenv("PORT"); p != "" {
			return fmt.Sprintf(":%s", p)
		}
		return ":8080"
	}()

	log.Printf("Server listening %s", port)
	http.ListenAndServe(port, nil)
}
