package main

import (
	"flag"
	"log"

	"github.com/micpst/fts-engine/api"
)

func main() {
	port := flag.Uint("p", 3000, "Port for the server to listen on")
	uploadLimit := flag.Int64("m", 8<<27, "Memory limit for file uploads (in bytes)")
	flag.Parse()

	s := api.New(&api.Config{
		Port:               *port,
		MaxMultipartMemory: *uploadLimit,
	})
	err := s.Run()
	log.Fatal(err)
}
