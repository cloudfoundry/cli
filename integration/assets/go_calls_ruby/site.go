package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
)

func main() {
	http.HandleFunc("/", hello)
	fmt.Println("listening...")

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", os.Getenv("PORT")),
		Handler: nil,
	}
	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

func hello(res http.ResponseWriter, req *http.Request) {
	bundlerVersion, err := exec.Command("bundle", "--version").Output()
	if err != nil {
		log.Print("ERROR:", err)
		fmt.Fprintf(res, "ERROR: %v\n", err)
	} else {
		fmt.Fprintf(res, "The bundler version is: %s\n", bundlerVersion)
	}
}
