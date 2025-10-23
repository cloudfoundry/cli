package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"code.cloudfoundry.org/cli/v8/integration/assets/hydrabroker/app"
)

func main() {
	port := port()
	log.Printf("Listening on port: %s", port)
	http.Handle("/", app.App())
	http.ListenAndServe(port, nil)
}

func port() string {
	if port := os.Getenv("PORT"); port != "" {
		return fmt.Sprintf(":%s", port)
	}
	return ":8080"
}
