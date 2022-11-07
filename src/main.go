package main

import (
	"flag"
	"log"

	"github.com/micpst/full-text-search-engine/src/app"
)

func main() {
	port := flag.Uint("p", 3000, "Port for the server to listen on")
	flag.Parse()

	a := app.New()
	err := a.Run(port)
	log.Fatal(err)
}
