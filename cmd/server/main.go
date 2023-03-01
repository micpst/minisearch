package main

import (
	"flag"
	"log"

	"github.com/micpst/minisearch/api"
	"github.com/micpst/minisearch/pkg/tokenizer"
)

func main() {
	lang := flag.String("l", string(tokenizer.ENGLISH), "Default language for the search engine")
	port := flag.Uint("p", 3000, "Port for the server to listen on")
	uploadLimit := flag.Int64("m", 8<<27, "Memory limit for file uploads (in bytes)")
	flag.Parse()

	s := api.New(&api.Config{
		DefaultLanguage: tokenizer.Language(*lang),
		Port:            *port,
		UploadLimit:     *uploadLimit,
	})
	err := s.Run()
	log.Fatal(err)
}
