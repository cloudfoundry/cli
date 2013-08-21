package main

import (
	"os"
	"cf/app"
)

func main() {
	app := app.New()
	app.Run(os.Args)
}
