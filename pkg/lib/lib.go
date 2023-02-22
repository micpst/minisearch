package lib

import (
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
