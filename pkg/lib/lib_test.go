package lib

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestCase[Given any, Expected any] struct {
	given    Given
	expected Expected
}

type TfIdfInput struct {
	termFrequency          float64
	documentsCount         int
	matchingDocumentsCount int
}

func TestTokenize(t *testing.T) {
	cases := []TestCase[string, []string]{
		{
			given:    "",
			expected: []string{},
		},
		{
			given:    "hello, world!",
			expected: []string{"hello", "world"},
		},
		{
			given:    "Lorem ipsum. Dolor? Sit amet!",
			expected: []string{"lorem", "ipsum", "dolor", "sit", "amet"},
		},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("'%v'", c.given), func(t *testing.T) {
			actual := Tokenize(c.given)
			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestCountTokens(t *testing.T) {
	cases := []TestCase[[]string, map[string]int]{
		{
			given:    []string{},
			expected: map[string]int{},
		},
		{
			given:    []string{"hello", "world"},
			expected: map[string]int{"world": 1, "hello": 1},
		},
		{
			given:    []string{"this", "is", "duplicated", "duplicated", "is"},
			expected: map[string]int{"duplicated": 2, "is": 2, "this": 1},
		},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%v", c.given), func(t *testing.T) {
			actual := Count(c.given)
			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestTfIdf(t *testing.T) {
	cases := []TestCase[TfIdfInput, float64]{
		{
			given: TfIdfInput{
				termFrequency:          1,
				documentsCount:         1,
				matchingDocumentsCount: 1,
			},
			expected: 0.2876820724517809,
		},
		{
			given: TfIdfInput{
				termFrequency:          0.5,
				documentsCount:         1,
				matchingDocumentsCount: 1,
			},
			expected: 0.14384103622589045,
		},
		{
			given: TfIdfInput{
				termFrequency:          1,
				documentsCount:         3,
				matchingDocumentsCount: 1,
			},
			expected: 0.9808292530117264,
		},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%v", c.given), func(t *testing.T) {
			actual := TfIdf(c.given.termFrequency, c.given.matchingDocumentsCount, c.given.documentsCount)
			assert.Equal(t, c.expected, actual)
		})
	}
}
