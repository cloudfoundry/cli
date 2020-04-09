package main

import (
	"fmt"
	"net/http"
	"os"

	"code.cloudfoundry.org/cli/integration/assets/hydrabroker/app"
)

func main() {
	port := port()
	fmt.Printf("Listening on port: %s\n", port)
	http.Handle("/", app.App())
	http.ListenAndServe(port, nil)
}

func port() string {
	if port := os.Getenv("PORT"); port != "" {
		return fmt.Sprintf(":%s", port)
	}
	return ":8080"
}
