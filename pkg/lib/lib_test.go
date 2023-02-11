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
	cases := []TestCase[[]string, map[string]uint32]{
		{
			given:    []string{},
			expected: map[string]uint32{},
		},
		{
			given:    []string{"hello", "world"},
			expected: map[string]uint32{"world": 1, "hello": 1},
		},
		{
			given:    []string{"this", "is", "duplicated", "duplicated", "is"},
			expected: map[string]uint32{"duplicated": 2, "is": 2, "this": 1},
		},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%v", c.given), func(t *testing.T) {
			actual := Count(c.given)
			assert.Equal(t, c.expected, actual)
		})
	}
}
