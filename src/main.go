package main

import (
	"github.com/micpst/full-text-search-engine/src/api/v1"
)

func main() {
	r := v1.InitRouter()
	_ = r.Run()
}
