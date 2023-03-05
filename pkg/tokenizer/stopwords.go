package tokenizer

import (
	"github.com/micpst/minisearch/pkg/tokenizer/stopwords"
)

type StopWords map[string]struct{}

var stopWords = map[Language]StopWords{
	ENGLISH:   stopwords.English,
	FRENCH:    stopwords.French,
	HUNGARIAN: stopwords.Hungarian,
	NORWEGIAN: stopwords.Norwegian,
	RUSSIAN:   stopwords.Russian,
	SPANISH:   stopwords.Spanish,
	SWEDISH:   stopwords.Swedish,
}
