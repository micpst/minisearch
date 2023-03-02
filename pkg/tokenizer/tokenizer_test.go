package tokenizer

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestCase[Given any, Expected any] struct {
	given    Given
	expected Expected
}

type TokenizeInput struct {
	params TokenizeParams
	config Config
}

type TokenizeOutput struct {
	tokens []string
	err    error
}

func TestTokenize(t *testing.T) {
	cases := []TestCase[TokenizeInput, TokenizeOutput]{
		{
			given: TokenizeInput{
				params: TokenizeParams{
					Text:            "",
					Language:        ENGLISH,
					AllowDuplicates: false,
				},
				config: Config{
					EnableStemming:  true,
					EnableStopWords: true,
				},
			},
			expected: TokenizeOutput{
				tokens: []string{},
				err:    nil,
			},
		},
		{
			given: TokenizeInput{
				params: TokenizeParams{
					Text:            "it's alive! it's alive!",
					Language:        ENGLISH,
					AllowDuplicates: false,
				},
				config: Config{
					EnableStemming:  true,
					EnableStopWords: true,
				},
			},
			expected: TokenizeOutput{
				tokens: []string{"it", "aliv"},
				err:    nil,
			},
		},
		{
			given: TokenizeInput{
				params: TokenizeParams{
					Text:            "it's alive! it's alive!",
					Language:        ENGLISH,
					AllowDuplicates: true,
				},
				config: Config{
					EnableStemming:  true,
					EnableStopWords: true,
				},
			},
			expected: TokenizeOutput{
				tokens: []string{"it", "aliv", "it", "aliv"},
				err:    nil,
			},
		},
		{
			given: TokenizeInput{
				params: TokenizeParams{
					Text:            "Lorem ipsum. Dolor? Sit amet!",
					Language:        "pl",
					AllowDuplicates: false,
				},
				config: Config{
					EnableStemming:  true,
					EnableStopWords: true,
				},
			},
			expected: TokenizeOutput{
				tokens: nil,
				err:    LanguageNotSupported,
			},
		},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("'%v'", c.given), func(t *testing.T) {
			actual, err := Tokenize(&c.given.params, &c.given.config)

			assert.Equal(t, c.expected.tokens, actual)
			assert.Equal(t, c.expected.err, err)
		})
	}
}
