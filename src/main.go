package main

import (
	"flag"
	"log"

	"github.com/micpst/full-text-search-engine/src/api"
)

func main() {
	port := flag.Uint("p", 3000, "Port for the server to listen on")
	flag.Parse()

	app := api.New()
	err := app.Run(port)
	log.Fatal(err)
}
