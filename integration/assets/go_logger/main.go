package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	go func() {
		for i := 0; i >= 0; i++ {
			fmt.Printf("this is log %d\n", i)
			time.Sleep(10 * time.Millisecond)
		}
	}()

	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), nil))
}
