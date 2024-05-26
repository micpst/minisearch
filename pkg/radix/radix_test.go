package radix

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestCase[Given any, Expected any] struct {
	given    Given
	expected Expected
}

type RecordInfo struct {
	termFrequency float64
}

func TestInsert(t *testing.T) {
	cases := []TestCase[[]InsertParams[string, RecordInfo], Trie[string, RecordInfo]]{
		{
			given: []InsertParams[string, RecordInfo]{
				{
					Id:   "2e48c6df-bafa-4981-b61a-16879dcdde2a",
					Word: "territory",
					Data: RecordInfo{termFrequency: 3.64961844222847},
				},
				{
					Id:   "998c8de6-3c50-4e9e-9835-10f8d1215327",
					Word: "territory",
					Data: RecordInfo{termFrequency: 1.29513358272291},
				},
			},
			expected: Trie[string, RecordInfo]{
				length: 1,
				root: &node[string, RecordInfo]{
					subword: nil,
					data:    map[string]RecordInfo{},
					children: map[rune]*node[string, RecordInfo]{
						't': {
							subword:  []rune("territory"),
							children: map[rune]*node[string, RecordInfo]{},
							data: map[string]RecordInfo{
								"2e48c6df-bafa-4981-b61a-16879dcdde2a": {
									termFrequency: 3.64961844222847,
								},
								"998c8de6-3c50-4e9e-9835-10f8d1215327": {
									termFrequency: 1.29513358272291,
								},
							},
						},
					},
				},
			},
		},
		{
			given: []InsertParams[string, RecordInfo]{
				{
					Id:   "2e48c6df-bafa-4981-b61a-16879dcdde2a",
					Word: "australian",
					Data: RecordInfo{termFrequency: 3.64961844222847},
				},
				{
					Id:   "998c8de6-3c50-4e9e-9835-10f8d1215327",
					Word: "territory",
					Data: RecordInfo{termFrequency: 1.29513358272291},
				},
			},
			expected: Trie[string, RecordInfo]{
				length: 2,
				root: &node[string, RecordInfo]{
					subword: nil,
					data:    map[string]RecordInfo{},
					children: map[rune]*node[string, RecordInfo]{
						'a': {
							subword: []rune("australian"),
							data: map[string]RecordInfo{
								"2e48c6df-bafa-4981-b61a-16879dcdde2a": {
									termFrequency: 3.64961844222847,
								},
							},
							children: map[rune]*node[string, RecordInfo]{},
						},
						't': {
							subword: []rune("territory"),
							data: map[string]RecordInfo{
								"998c8de6-3c50-4e9e-9835-10f8d1215327": {
									termFrequency: 1.29513358272291,
								},
							},
							children: map[rune]*node[string, RecordInfo]{},
						},
					},
				},
			},
		},
		{
			given: []InsertParams[string, RecordInfo]{
				{
					Id:   "2e48c6df-bafa-4981-b61a-16879dcdde2a",
					Word: "terrorist",
					Data: RecordInfo{termFrequency: 3.64961844222847},
				},
				{
					Id:   "998c8de6-3c50-4e9e-9835-10f8d1215327",
					Word: "territory",
					Data: RecordInfo{termFrequency: 1.29513358272291},
				},
			},
			expected: Trie[string, RecordInfo]{
				length: 2,
				root: &node[string, RecordInfo]{
					subword: nil,
					data:    map[string]RecordInfo{},
					children: map[rune]*node[string, RecordInfo]{
						't': {
							subword: []rune("terr"),
							data:    map[string]RecordInfo{},
							children: map[rune]*node[string, RecordInfo]{
								'i': {
									subword: []rune("itory"),
									data: map[string]RecordInfo{
										"998c8de6-3c50-4e9e-9835-10f8d1215327": {termFrequency: 1.29513358272291},
									},
									children: map[rune]*node[string, RecordInfo]{},
								},
								'o': {
									subword: []rune("orist"),
									data: map[string]RecordInfo{
										"2e48c6df-bafa-4981-b61a-16879dcdde2a": {termFrequency: 3.64961844222847},
									},
									children: map[rune]*node[string, RecordInfo]{},
								},
							},
						},
					},
				},
			},
		},
		{
			given: []InsertParams[string, RecordInfo]{
				{
					Id:   "2e48c6df-bafa-4981-b61a-16879dcdde2a",
					Word: "autobiography",
					Data: RecordInfo{termFrequency: 3.64961844222847},
				},
				{
					Id:   "998c8de6-3c50-4e9e-9835-10f8d1215327",
					Word: "auto",
					Data: RecordInfo{termFrequency: 1.29513358272291},
				},
			},
			expected: Trie[string, RecordInfo]{
				length: 2,
				root: &node[string, RecordInfo]{
					subword: nil,
					data:    map[string]RecordInfo{},
					children: map[rune]*node[string, RecordInfo]{
						'a': {
							subword: []rune("auto"),
							data: map[string]RecordInfo{
								"998c8de6-3c50-4e9e-9835-10f8d1215327": {termFrequency: 1.29513358272291},
							},
							children: map[rune]*node[string, RecordInfo]{
								'b': {
									subword: []rune("biography"),
									data: map[string]RecordInfo{
										"2e48c6df-bafa-4981-b61a-16879dcdde2a": {termFrequency: 3.64961844222847},
									},
									children: map[rune]*node[string, RecordInfo]{},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%v", c.given), func(t *testing.T) {
			index := New[string, RecordInfo]()

			for _, p := range c.given {
				index.Insert(&p)
			}

			assert.Equal(t, &c.expected, index)
		})
	}
}

func TestDelete(t *testing.T) {
	cases := []TestCase[[]DeleteParams[string], Trie[string, RecordInfo]]{
		{
			given: []DeleteParams[string]{
				{
					Id:   "2e48c6df-bafa-4981-b61a-16879dcdde2a",
					Word: "australian",
				},
			},
			expected: Trie[string, RecordInfo]{
				length: 2,
				root: &node[string, RecordInfo]{
					subword: nil,
					data:    map[string]RecordInfo{},
					children: map[rune]*node[string, RecordInfo]{
						'a': {
							subword: []rune("australia"),
							data: map[string]RecordInfo{
								"998c8de6-3c50-4e9e-9835-10f8d1215327": {termFrequency: 1.29513358272291},
							},
							children: map[rune]*node[string, RecordInfo]{},
						},
						't': {
							subword: []rune("territory"),
							data: map[string]RecordInfo{
								"1e44c6df-bafa-4981-b61a-16879d2dddghf": {termFrequency: 2.27923284424328},
							},
							children: map[rune]*node[string, RecordInfo]{},
						},
					},
				},
			},
		},
		{
			given: []DeleteParams[string]{
				{
					Id:   "998c8de6-3c50-4e9e-9835-10f8d1215327",
					Word: "australia",
				},
			},
			expected: Trie[string, RecordInfo]{
				length: 2,
				root: &node[string, RecordInfo]{
					subword: nil,
					data:    map[string]RecordInfo{},
					children: map[rune]*node[string, RecordInfo]{
						'a': {
							subword: []rune("australian"),
							data: map[string]RecordInfo{
								"2e48c6df-bafa-4981-b61a-16879dcdde2a": {termFrequency: 3.64961844222847},
							},
							children: map[rune]*node[string, RecordInfo]{},
						},
						't': {
							subword: []rune("territory"),
							data: map[string]RecordInfo{
								"1e44c6df-bafa-4981-b61a-16879d2dddghf": {termFrequency: 2.27923284424328},
							},
							children: map[rune]*node[string, RecordInfo]{},
						},
					},
				},
			},
		},
		{
			given: []DeleteParams[string]{
				{
					Id:   "11111111-3c50-4e9e-9835-10f8d1215327",
					Word: "gibberish",
				},
			},
			expected: Trie[string, RecordInfo]{
				length: 3,
				root: &node[string, RecordInfo]{
					subword: nil,
					data:    map[string]RecordInfo{},
					children: map[rune]*node[string, RecordInfo]{
						'a': {
							subword: []rune("australia"),
							data: map[string]RecordInfo{
								"998c8de6-3c50-4e9e-9835-10f8d1215327": {termFrequency: 1.29513358272291},
							},
							children: map[rune]*node[string, RecordInfo]{
								'n': {
									subword: []rune("n"),
									data: map[string]RecordInfo{
										"2e48c6df-bafa-4981-b61a-16879dcdde2a": {termFrequency: 3.64961844222847},
									},
									children: map[rune]*node[string, RecordInfo]{},
								},
							},
						},
						't': {
							subword: []rune("territory"),
							data: map[string]RecordInfo{
								"1e44c6df-bafa-4981-b61a-16879d2dddghf": {termFrequency: 2.27923284424328},
							},
							children: map[rune]*node[string, RecordInfo]{},
						},
					},
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%v", c.given), func(t *testing.T) {
			index := New[string, RecordInfo]()
			index.Insert(&InsertParams[string, RecordInfo]{
				Id:   "2e48c6df-bafa-4981-b61a-16879dcdde2a",
				Word: "australian",
				Data: RecordInfo{termFrequency: 3.64961844222847},
			})
			index.Insert(&InsertParams[string, RecordInfo]{
				Id:   "998c8de6-3c50-4e9e-9835-10f8d1215327",
				Word: "australia",
				Data: RecordInfo{termFrequency: 1.29513358272291},
			})
			index.Insert(&InsertParams[string, RecordInfo]{
				Id:   "1e44c6df-bafa-4981-b61a-16879d2dddghf",
				Word: "territory",
				Data: RecordInfo{termFrequency: 2.27923284424328},
			})

			for _, p := range c.given {
				index.Delete(&p)
			}

			assert.Equal(t, &c.expected, index)
		})
	}
}

func TestFind(t *testing.T) {
	cases := []TestCase[FindParams, map[string]RecordInfo]{
		{
			given: FindParams{
				Term:      "what",
				Tolerance: 0,
				Exact:     false,
			},
			expected: map[string]RecordInfo{},
		},
		{
			given: FindParams{
				Term:  "australia",
				Exact: true,
			},
			expected: map[string]RecordInfo{
				"998c8de6-3c50-4e9e-9835-10f8d1215327": {termFrequency: 1.29513358272291},
			},
		},
		{
			given: FindParams{
				Term:      "australia",
				Tolerance: 0,
				Exact:     false,
			},
			expected: map[string]RecordInfo{
				"998c8de6-3c50-4e9e-9835-10f8d1215327": {termFrequency: 1.29513358272291},
				"2e48c6df-bafa-4981-b61a-16879dcdde2a": {termFrequency: 3.64961844222847},
			},
		},
		{
			given: FindParams{
				Term:      "australian",
				Tolerance: 2,
				Exact:     false,
			},
			expected: map[string]RecordInfo{
				"2e48c6df-bafa-4981-b61a-16879dcdde2a": {termFrequency: 3.64961844222847},
			},
		},
		{
			given: FindParams{
				Term:      "austra",
				Tolerance: 3,
				Exact:     false,
			},
			expected: map[string]RecordInfo{
				"998c8de6-3c50-4e9e-9835-10f8d1215327": {termFrequency: 1.29513358272291},
			},
		},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%v", c.given), func(t *testing.T) {
			index := New[string, RecordInfo]()
			index.Insert(&InsertParams[string, RecordInfo]{
				Id:   "2e48c6df-bafa-4981-b61a-16879dcdde2a",
				Word: "australian",
				Data: RecordInfo{termFrequency: 3.64961844222847},
			})
			index.Insert(&InsertParams[string, RecordInfo]{
				Id:   "998c8de6-3c50-4e9e-9835-10f8d1215327",
				Word: "australia",
				Data: RecordInfo{termFrequency: 1.29513358272291},
			})
			index.Insert(&InsertParams[string, RecordInfo]{
				Id:   "1e44c6df-bafa-4981-b61a-16879d2dddghf",
				Word: "austrian",
				Data: RecordInfo{termFrequency: 2.27923284424328},
			})

			results := index.Find(&c.given)

			assert.Equal(t, c.expected, results)
		})
	}
}
