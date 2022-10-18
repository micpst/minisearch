package main

import (
	"github.com/micpst/full-text-search-engine/api/v1"
)

func main() {
	r := v1.InitRoutes()
	_ = r.Run()
}
