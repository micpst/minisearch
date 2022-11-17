package lib

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestCase[Input any, Expected any] struct {
	input    Input
	expected Expected
}

func TestTokenize(t *testing.T) {
	cases := []TestCase[string, []string]{
		{
			input:    "",
			expected: []string{},
		},
		{
			input:    "hello, world!",
			expected: []string{"hello", "world"},
		},
		{
			input:    "Lorem ipsum. Dolor? Sit amet!",
			expected: []string{"lorem", "ipsum", "dolor", "sit", "amet"},
		},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("'%v'", c.input), func(t *testing.T) {
			actual := Tokenize(c.input)
			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestCountTokens(t *testing.T) {
	cases := []TestCase[[]string, map[string]uint32]{
		{
			input:    []string{},
			expected: map[string]uint32{},
		},
		{
			input:    []string{"hello", "world"},
			expected: map[string]uint32{"world": 1, "hello": 1},
		},
		{
			input:    []string{"this", "is", "duplicated", "duplicated", "is"},
			expected: map[string]uint32{"duplicated": 2, "is": 2, "this": 1},
		},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%v", c.input), func(t *testing.T) {
			actual := Count(c.input)
			assert.Equal(t, c.expected, actual)
		})
	}
}
