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
	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		panic(err)
	}
}

func hello(res http.ResponseWriter, req *http.Request) {
	bundlerVersion, err := exec.Command("bundle", "--version").Output()
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		log.Print("ERROR:", err)
		fmt.Fprintf(res, "ERROR: %v\n", err)
	} else {
		fmt.Fprintf(res, "The bundler version is: %s\n", bundlerVersion)
	}
}
