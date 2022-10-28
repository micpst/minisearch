package lib

import (
	"fmt"
	"reflect"
	"testing"
)

type TokenizeTestCase struct {
	input    string
	expected []string
}

type CountTestCase struct {
	input    []string
	expected map[string]uint32
}

func TestTokenize(t *testing.T) {
	cases := []TokenizeTestCase{
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
			if !reflect.DeepEqual(actual, c.expected) {
				t.Fatalf("Expected %v, got %v", c.expected, actual)
			}
		})
	}
}

func TestCountTokens(t *testing.T) {
	cases := []CountTestCase{
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
			if !reflect.DeepEqual(actual, c.expected) {
				t.Fatalf("Expected %v, got %v", c.expected, actual)
			}
		})
	}
}
