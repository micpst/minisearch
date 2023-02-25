package lib

import (
	"math"
	"regexp"
	"strings"
)

var punctuationRegex = regexp.MustCompile(`[^\w|\s]`)

func Tokenize(data string) []string {
	data = punctuationRegex.ReplaceAllString(data, "")
	data = strings.ToLower(data)
	return strings.Fields(data)
}

func Count(tokens []string) map[string]int {
	dict := make(map[string]int)
	for _, token := range tokens {
		dict[token]++
	}
	return dict
}

func TfIdf(tf float64, matchingDocsCount int, docsCount int) float64 {
	idf := math.Log(1 + (float64(docsCount-matchingDocsCount)+0.5)/(float64(matchingDocsCount)+0.5))
	return tf * idf
}
