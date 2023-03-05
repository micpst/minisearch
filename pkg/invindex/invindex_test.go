package invindex

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestCase[Given any, Expected any] struct {
	given    Given
	expected Expected
}

type Input struct {
	id            string
	term          string
	termFrequency float64
}

func TestAdd(t *testing.T) {
	cases := []TestCase[Input, InvIndex]{
		{
			given: Input{
				id:            "998c8de6-3c50-4e9e-9835-10f8d1215327",
				term:          "territory",
				termFrequency: 1.29513358272291,
			},
			expected: InvIndex{
				"territory": {
					"998c8de6-3c50-4e9e-9835-10f8d1215327": {
						TermFrequency: 1.29513358272291,
					},
				},
			},
		},
		{
			given: Input{
				id:            "2e48c6df-bafa-4981-b61a-16879dcdde2a",
				term:          "australian",
				termFrequency: 3.64961844222847,
			},
			expected: InvIndex{
				"australian": {
					"2e48c6df-bafa-4981-b61a-16879dcdde2a": {
						TermFrequency: 3.64961844222847,
					},
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%v", c.given), func(t *testing.T) {
			idx := InvIndex{}

			idx.Add(c.given.id, c.given.term, c.given.termFrequency)

			assert.Equal(t, c.expected, idx)
		})
	}
}

func TestRemove(t *testing.T) {
	cases := []TestCase[Input, InvIndex]{
		{
			given: Input{
				id:   "2e48c6df-bafa-4981-b61a-16879dcdde2a",
				term: "territory",
			},
			expected: InvIndex{
				"australian": {
					"2e48c6df-bafa-4981-b61a-16879dcdde2a": {
						TermFrequency: 3.64961844222847,
					},
				},
				"territory": {
					"998c8de6-3c50-4e9e-9835-10f8d1215327": {
						TermFrequency: 1.29513358272291,
					},
				},
			},
		},
		{
			given: Input{
				id:   "998c8de6-3c50-4e9e-9835-10f8d1215327",
				term: "territory",
			},
			expected: InvIndex{
				"australian": {
					"2e48c6df-bafa-4981-b61a-16879dcdde2a": {
						TermFrequency: 3.64961844222847,
					},
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%v", c.given), func(t *testing.T) {
			idx := InvIndex{
				"territory": {
					"998c8de6-3c50-4e9e-9835-10f8d1215327": {
						TermFrequency: 1.29513358272291,
					},
				},
				"australian": {
					"2e48c6df-bafa-4981-b61a-16879dcdde2a": {
						TermFrequency: 3.64961844222847,
					},
				},
			}

			idx.Remove(c.given.id, c.given.term)

			assert.Equal(t, c.expected, idx)
		})
	}
}

func TestFind(t *testing.T) {
	cases := []TestCase[Input, map[string]RecordInfo]{
		{
			given: Input{
				term: "territory",
			},
			expected: map[string]RecordInfo{
				"998c8de6-3c50-4e9e-9835-10f8d1215327": {
					TermFrequency: 1.29513358272291,
				},
			},
		},
		{
			given: Input{
				term: "theory",
			},
			expected: map[string]RecordInfo{},
		},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%v", c.given), func(t *testing.T) {
			idx := InvIndex{
				"territory": {
					"998c8de6-3c50-4e9e-9835-10f8d1215327": {
						TermFrequency: 1.29513358272291,
					},
				},
				"australian": {
					"2e48c6df-bafa-4981-b61a-16879dcdde2a": {
						TermFrequency: 3.64961844222847,
					},
				},
			}

			idInfos := idx.Find(c.given.term)

			assert.Equal(t, c.expected, idInfos)
		})
	}
}
