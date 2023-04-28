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

type BM25Input struct {
	termFrequency          float64
	matchingDocumentsCount int
	fieldLength            int
	averageFieldLength     float64
	documentsCount         int
	k                      float64
	b                      float64
	d                      float64
}

type BoundedLevenshteinInput struct {
	a         []rune
	b         []rune
	tolerance int
}

type BoundedLevenshteinOutput struct {
	distance  int
	isBounded bool
}

type CommonPrefixInput struct {
	a []rune
	b []rune
}

type CommonPrefixOutput struct {
	commonPrefix []rune
	equal        bool
}

type PaginateInput struct {
	offset int
	limit  int
	count  int
}

type PaginateOutput struct {
	start int
	stop  int
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

func TestBM25(t *testing.T) {
	cases := []TestCase[BM25Input, float64]{
		{
			given: BM25Input{
				termFrequency:          1.0,
				documentsCount:         1,
				matchingDocumentsCount: 1,
				fieldLength:            10,
				averageFieldLength:     10.0,
				k:                      1.0,
				b:                      1.0,
				d:                      1.0,
			},
			expected: 0.43152310867767135,
		},
		{
			given: BM25Input{
				termFrequency:          0.5,
				documentsCount:         2,
				matchingDocumentsCount: 1,
				fieldLength:            10,
				averageFieldLength:     12.43,
				k:                      1.2,
				b:                      0.75,
				d:                      0.5,
			},
			expected: 0.7276874539155506,
		},
		{
			given: BM25Input{
				termFrequency:          0.75,
				documentsCount:         3,
				matchingDocumentsCount: 1,
				fieldLength:            10,
				averageFieldLength:     5.32,
				k:                      1.2,
				b:                      0.75,
				d:                      0.5,
			},
			expected: 0.7691433563655649,
		},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%v", c.given), func(t *testing.T) {
			actual := BM25(
				c.given.termFrequency,
				c.given.matchingDocumentsCount,
				c.given.fieldLength,
				c.given.averageFieldLength,
				c.given.documentsCount,
				c.given.k,
				c.given.b,
				c.given.d,
			)
			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestPaginate(t *testing.T) {
	cases := []TestCase[PaginateInput, PaginateOutput]{
		{
			given: PaginateInput{
				offset: 5,
				limit:  5,
				count:  10,
			},
			expected: PaginateOutput{
				start: 5,
				stop:  10,
			},
		},
		{
			given: PaginateInput{
				offset: 0,
				limit:  10,
				count:  11,
			},
			expected: PaginateOutput{
				start: 0,
				stop:  10,
			},
		},
		{
			given: PaginateInput{
				offset: 11,
				limit:  10,
				count:  20,
			},
			expected: PaginateOutput{
				start: 11,
				stop:  20,
			},
		},
		{
			given: PaginateInput{
				offset: 0,
				limit:  10,
				count:  5,
			},
			expected: PaginateOutput{
				start: 0,
				stop:  5,
			},
		},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%v", c.given), func(t *testing.T) {
			start, stop := Paginate(c.given.offset, c.given.limit, c.given.count)

			assert.Equal(t, c.expected.start, start)
			assert.Equal(t, c.expected.stop, stop)
		})
	}
}

func TestCommonPrefix(t *testing.T) {
	cases := []TestCase[CommonPrefixInput, CommonPrefixOutput]{
		{
			given: CommonPrefixInput{
				a: []rune("hello"),
				b: []rune("world"),
			},
			expected: CommonPrefixOutput{
				commonPrefix: []rune(""),
				equal:        false,
			},
		},
		{
			given: CommonPrefixInput{
				a: []rune("hello"),
				b: []rune("hello"),
			},
			expected: CommonPrefixOutput{
				commonPrefix: []rune("hello"),
				equal:        true,
			},
		},
		{
			given: CommonPrefixInput{
				a: []rune("hello"),
				b: []rune("hello world"),
			},
			expected: CommonPrefixOutput{
				commonPrefix: []rune("hello"),
				equal:        false,
			},
		},
		{
			given: CommonPrefixInput{
				a: []rune("读写"),
				b: []rune("读写汉字"),
			},
			expected: CommonPrefixOutput{
				commonPrefix: []rune("读写"),
				equal:        false,
			},
		},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%v", c.given), func(t *testing.T) {
			commonPrefix, eq := CommonPrefix(c.given.a, c.given.b)

			assert.Equal(t, c.expected.commonPrefix, commonPrefix)
			assert.Equal(t, c.expected.equal, eq)
		})
	}
}

func TestBoundedLevenshtein(t *testing.T) {
	cases := []TestCase[BoundedLevenshteinInput, BoundedLevenshteinOutput]{
		{
			given: BoundedLevenshteinInput{
				a:         []rune(""),
				b:         []rune(""),
				tolerance: 1,
			},
			expected: BoundedLevenshteinOutput{
				distance:  0,
				isBounded: true,
			},
		},
		{
			given: BoundedLevenshteinInput{
				a:         []rune("kitten"),
				b:         []rune("sitting"),
				tolerance: 1,
			},
			expected: BoundedLevenshteinOutput{
				distance:  -1,
				isBounded: false,
			},
		},
		{
			given: BoundedLevenshteinInput{
				a:         []rune("Saturday"),
				b:         []rune("Sunday"),
				tolerance: 3,
			},
			expected: BoundedLevenshteinOutput{
				distance:  3,
				isBounded: true,
			},
		},
		{
			given: BoundedLevenshteinInput{
				a:         []rune("foo"),
				b:         []rune("bar"),
				tolerance: 2,
			},
			expected: BoundedLevenshteinOutput{
				distance:  -1,
				isBounded: false,
			},
		},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%v", c.given), func(t *testing.T) {
			distance, isBounded := BoundedLevenshtein(c.given.a, c.given.b, c.given.tolerance)

			assert.Equal(t, c.expected.distance, distance)
			assert.Equal(t, c.expected.isBounded, isBounded)
		})
	}
}
