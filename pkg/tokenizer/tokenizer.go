package tokenizer

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

const (
	ENGLISH   Language = "en"
	FRENCH    Language = "fr"
	HUNGARIAN Language = "hu"
	NORWEGIAN Language = "no"
	RUSSIAN   Language = "ru"
	SPANISH   Language = "es"
	SWEDISH   Language = "sv"
)

var splitRules = map[Language]*regexp.Regexp{
	ENGLISH:   regexp.MustCompile(`[^A-Za-zàèéìòóù0-9_'-]`),
	FRENCH:    regexp.MustCompile(`[^a-z0-9äâàéèëêïîöôùüûœç-]`),
	HUNGARIAN: regexp.MustCompile(`[^a-z0-9áéíóöőúüűÁÉÍÓÖŐÚÜŰ]`),
	NORWEGIAN: regexp.MustCompile(`[^a-z0-9_æøåÆØÅäÄöÖüÜ]`),
	RUSSIAN:   regexp.MustCompile(`[^a-z0-9а-яА-ЯёЁ]`),
	SPANISH:   regexp.MustCompile(`[^a-z0-9A-Zá-úÁ-ÚñÑüÜ]`),
	SWEDISH:   regexp.MustCompile(`[^a-z0-9_åÅäÄöÖüÜ-]`),
}

var (
	LanguageNotSupported = errors.New("language not supported")
)

type Language string

type TokenizeInput struct {
	Text            string
	Language        Language
	AllowDuplicates bool
}

type Config struct {
	EnableStemming  bool
	EnableStopWords bool
}

type Tokenizer struct {
	config *Config
	cache  map[string]string
}

func New(c *Config) *Tokenizer {
	return &Tokenizer{
		config: c,
		cache:  make(map[string]string),
	}
}

func (t *Tokenizer) IsSupportedLanguage(language Language) bool {
	_, ok := splitRules[language]
	return ok
}

func (t *Tokenizer) Tokenize(input *TokenizeInput) ([]string, error) {
	splitRule, ok := splitRules[input.Language]
	if !ok {
		return nil, LanguageNotSupported
	}

	input.Text = strings.ToLower(input.Text)
	splitText := splitRule.Split(input.Text, -1)

	tokens := make([]string, 0)
	for _, token := range splitText {

		if normToken := t.normalizeToken(token, input.Language); normToken != "" {
			tokens = append(tokens, normToken)
		}
	}

	return tokens, nil
}

func (t *Tokenizer) normalizeToken(token string, language Language) string {
	if token == "" {
		return ""
	}

	key := fmt.Sprintf("%s:%s", language, token)
	if cached, ok := t.cache[key]; ok {
		return cached
	}

	if _, ok := stopWords[language][token]; t.config.EnableStopWords && ok {
		t.cache[key] = ""
		return ""
	}

	if stem, ok := stems[language]; t.config.EnableStemming && ok {
		token = stem(token, false)
	}

	t.cache[key] = token

	return token
}
