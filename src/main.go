package main

import (
	"fluxid-loop/src/app"
	"os"
)

func main() {
	// Ensure logs go to stdout and no timestamp
	app.SetupLogger(os.Stdout)
	app.PrintHello()
}
