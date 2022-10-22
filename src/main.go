package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/micpst/full-text-search-engine/src/api/v1"
)

func main() {
	port := flag.Int("p", 3000, "Port for the server to listen on")
	flag.Parse()

	r := v1.InitRouter()
	err := r.Run(fmt.Sprintf(":%d", *port))
	log.Fatal(err)
}
