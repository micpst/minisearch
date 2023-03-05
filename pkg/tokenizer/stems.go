package tokenizer

import (
	"github.com/kljensen/snowball/english"
	"github.com/kljensen/snowball/french"
	"github.com/kljensen/snowball/hungarian"
	"github.com/kljensen/snowball/norwegian"
	"github.com/kljensen/snowball/russian"
	"github.com/kljensen/snowball/spanish"
	"github.com/kljensen/snowball/swedish"
)

type Stem func(string, bool) string

var stems = map[Language]Stem{
	ENGLISH:   english.Stem,
	FRENCH:    french.Stem,
	HUNGARIAN: hungarian.Stem,
	NORWEGIAN: norwegian.Stem,
	RUSSIAN:   russian.Stem,
	SPANISH:   spanish.Stem,
	SWEDISH:   swedish.Stem,
}
