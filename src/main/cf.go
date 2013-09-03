package main

import (
	"os"
	"cf/app"
)

func main() {
	app, err := app.New()
	if err != nil {
		return
	}
	app.Run(os.Args)
}
