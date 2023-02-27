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
	tokensCounter := make(map[string]int, len(tokens))
	for _, token := range tokens {
		tokensCounter[token]++
	}
	return tokensCounter
}

func TfIdf(tf float64, matchingDocsCount int, docsCount int) float64 {
	idf := math.Log(1 + (float64(docsCount-matchingDocsCount)+0.5)/(float64(matchingDocsCount)+0.5))
	return tf * idf
}

func Paginate(offset int, limit int, sliceLength int) (int, int) {
	if offset > sliceLength {
		offset = sliceLength
	}

	end := offset + limit
	if end > sliceLength {
		end = sliceLength
	}

	return offset, end
}
