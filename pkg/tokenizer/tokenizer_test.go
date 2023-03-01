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

type TokenizeOutput struct {
	tokens []string
	err    error
}

func TestTokenize(t *testing.T) {
	tr := New(&Config{
		EnableStemming:  true,
		EnableStopWords: true,
	})
	cases := []TestCase[TokenizeInput, TokenizeOutput]{
		{
			given: TokenizeInput{
				Text:            "",
				Language:        ENGLISH,
				AllowDuplicates: false,
			},
			expected: TokenizeOutput{
				tokens: []string{},
				err:    nil,
			},
		},
		{
			given: TokenizeInput{
				Text:            "it's alive! it's alive!",
				Language:        ENGLISH,
				AllowDuplicates: false,
			},
			expected: TokenizeOutput{
				tokens: []string{"it'", "aliv", "it'"},
				err:    nil,
			},
		},
		{
			given: TokenizeInput{
				Text:            "it's alive! it's alive!",
				Language:        ENGLISH,
				AllowDuplicates: true,
			},
			expected: TokenizeOutput{
				tokens: []string{"it'", "aliv", "it'", "aliv"},
				err:    nil,
			},
		},
		{
			given: TokenizeInput{
				Text:            "Lorem ipsum. Dolor? Sit amet!",
				Language:        "pl",
				AllowDuplicates: false,
			},
			expected: TokenizeOutput{
				tokens: nil,
				err:    LanguageNotSupported,
			},
		},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("'%v'", c.given), func(t *testing.T) {
			actual, err := tr.Tokenize(&c.given)

			assert.Equal(t, c.expected.tokens, actual)
			assert.Equal(t, c.expected.err, err)
		})
	}
}
