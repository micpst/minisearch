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

func TestInsert(t *testing.T) {
	cases := []TestCase[[]InsertParams, Trie]{
		{
			given: []InsertParams{
				{
					Id:            "2e48c6df-bafa-4981-b61a-16879dcdde2a",
					Word:          "territory",
					TermFrequency: 3.64961844222847,
				},
				{
					Id:            "998c8de6-3c50-4e9e-9835-10f8d1215327",
					Word:          "territory",
					TermFrequency: 1.29513358272291,
				},
			},
			expected: Trie{
				length: 1,
				root: &node{
					subword: nil,
					infos:   []RecordInfo{},
					children: map[rune]*node{
						't': {
							subword: []rune("territory"),
							infos: []RecordInfo{
								{
									Id:            "2e48c6df-bafa-4981-b61a-16879dcdde2a",
									TermFrequency: 3.64961844222847,
								},
								{
									Id:            "998c8de6-3c50-4e9e-9835-10f8d1215327",
									TermFrequency: 1.29513358272291,
								},
							},
							children: map[rune]*node{},
						},
					},
				},
			},
		},
		{
			given: []InsertParams{
				{
					Id:            "2e48c6df-bafa-4981-b61a-16879dcdde2a",
					Word:          "australian",
					TermFrequency: 3.64961844222847,
				},
				{
					Id:            "998c8de6-3c50-4e9e-9835-10f8d1215327",
					Word:          "territory",
					TermFrequency: 1.29513358272291,
				},
			},
			expected: Trie{
				length: 2,
				root: &node{
					subword: nil,
					infos:   []RecordInfo{},
					children: map[rune]*node{
						'a': {
							subword: []rune("australian"),
							infos: []RecordInfo{
								{
									Id:            "2e48c6df-bafa-4981-b61a-16879dcdde2a",
									TermFrequency: 3.64961844222847,
								},
							},
							children: map[rune]*node{},
						},
						't': {
							subword: []rune("territory"),
							infos: []RecordInfo{
								{
									Id:            "998c8de6-3c50-4e9e-9835-10f8d1215327",
									TermFrequency: 1.29513358272291,
								},
							},
							children: map[rune]*node{},
						},
					},
				},
			},
		},
		{
			given: []InsertParams{
				{
					Id:            "2e48c6df-bafa-4981-b61a-16879dcdde2a",
					Word:          "terrorist",
					TermFrequency: 3.64961844222847,
				},
				{
					Id:            "998c8de6-3c50-4e9e-9835-10f8d1215327",
					Word:          "territory",
					TermFrequency: 1.29513358272291,
				},
			},
			expected: Trie{
				length: 2,
				root: &node{
					subword: nil,
					infos:   []RecordInfo{},
					children: map[rune]*node{
						't': {
							subword: []rune("terr"),
							infos:   []RecordInfo{},
							children: map[rune]*node{
								'i': {
									subword: []rune("itory"),
									infos: []RecordInfo{
										{
											Id:            "998c8de6-3c50-4e9e-9835-10f8d1215327",
											TermFrequency: 1.29513358272291,
										},
									},
									children: map[rune]*node{},
								},
								'o': {
									subword: []rune("orist"),
									infos: []RecordInfo{
										{
											Id:            "2e48c6df-bafa-4981-b61a-16879dcdde2a",
											TermFrequency: 3.64961844222847,
										},
									},
									children: map[rune]*node{},
								},
							},
						},
					},
				},
			},
		},
		{
			given: []InsertParams{
				{
					Id:            "2e48c6df-bafa-4981-b61a-16879dcdde2a",
					Word:          "autobiography",
					TermFrequency: 3.64961844222847,
				},
				{
					Id:            "998c8de6-3c50-4e9e-9835-10f8d1215327",
					Word:          "auto",
					TermFrequency: 1.29513358272291,
				},
			},
			expected: Trie{
				length: 2,
				root: &node{
					subword: nil,
					infos:   []RecordInfo{},
					children: map[rune]*node{
						'a': {
							subword: []rune("auto"),
							infos: []RecordInfo{
								{
									Id:            "998c8de6-3c50-4e9e-9835-10f8d1215327",
									TermFrequency: 1.29513358272291,
								},
							},
							children: map[rune]*node{
								'b': {
									subword: []rune("biography"),
									infos: []RecordInfo{
										{
											Id:            "2e48c6df-bafa-4981-b61a-16879dcdde2a",
											TermFrequency: 3.64961844222847,
										},
									},
									children: map[rune]*node{},
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
			index := New()

			for _, p := range c.given {
				index.Insert(&p)
			}

			assert.Equal(t, &c.expected, index)
		})
	}
}

func TestDelete(t *testing.T) {
	cases := []TestCase[[]DeleteParams, Trie]{
		{
			given: []DeleteParams{
				{
					Id:   "2e48c6df-bafa-4981-b61a-16879dcdde2a",
					Word: "australian",
				},
			},
			expected: Trie{
				length: 2,
				root: &node{
					subword: nil,
					infos:   []RecordInfo{},
					children: map[rune]*node{
						'a': {
							subword: []rune("australia"),
							infos: []RecordInfo{
								{
									Id:            "998c8de6-3c50-4e9e-9835-10f8d1215327",
									TermFrequency: 1.29513358272291,
								},
							},
							children: map[rune]*node{},
						},
						't': {
							subword: []rune("territory"),
							infos: []RecordInfo{
								{
									Id:            "1e44c6df-bafa-4981-b61a-16879d2dddghf",
									TermFrequency: 2.27923284424328,
								},
							},
							children: map[rune]*node{},
						},
					},
				},
			},
		},
		{
			given: []DeleteParams{
				{
					Id:   "998c8de6-3c50-4e9e-9835-10f8d1215327",
					Word: "australia",
				},
			},
			expected: Trie{
				length: 2,
				root: &node{
					subword: nil,
					infos:   []RecordInfo{},
					children: map[rune]*node{
						'a': {
							subword: []rune("australian"),
							infos: []RecordInfo{
								{
									Id:            "2e48c6df-bafa-4981-b61a-16879dcdde2a",
									TermFrequency: 3.64961844222847,
								},
							},
							children: map[rune]*node{},
						},
						't': {
							subword: []rune("territory"),
							infos: []RecordInfo{
								{
									Id:            "1e44c6df-bafa-4981-b61a-16879d2dddghf",
									TermFrequency: 2.27923284424328,
								},
							},
							children: map[rune]*node{},
						},
					},
				},
			},
		},
		{
			given: []DeleteParams{
				{
					Id:   "11111111-3c50-4e9e-9835-10f8d1215327",
					Word: "gibberish",
				},
			},
			expected: Trie{
				length: 3,
				root: &node{
					subword: nil,
					infos:   []RecordInfo{},
					children: map[rune]*node{
						'a': {
							subword: []rune("australia"),
							infos: []RecordInfo{
								{
									Id:            "998c8de6-3c50-4e9e-9835-10f8d1215327",
									TermFrequency: 1.29513358272291,
								},
							},
							children: map[rune]*node{
								'n': {
									subword: []rune("n"),
									infos: []RecordInfo{
										{
											Id:            "2e48c6df-bafa-4981-b61a-16879dcdde2a",
											TermFrequency: 3.64961844222847,
										},
									},
									children: map[rune]*node{},
								},
							},
						},
						't': {
							subword: []rune("territory"),
							infos: []RecordInfo{
								{
									Id:            "1e44c6df-bafa-4981-b61a-16879d2dddghf",
									TermFrequency: 2.27923284424328,
								},
							},
							children: map[rune]*node{},
						},
					},
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%v", c.given), func(t *testing.T) {
			index := New()
			index.Insert(&InsertParams{
				Id:            "2e48c6df-bafa-4981-b61a-16879dcdde2a",
				Word:          "australian",
				TermFrequency: 3.64961844222847,
			})
			index.Insert(&InsertParams{
				Id:            "998c8de6-3c50-4e9e-9835-10f8d1215327",
				Word:          "australia",
				TermFrequency: 1.29513358272291,
			})
			index.Insert(&InsertParams{
				Id:            "1e44c6df-bafa-4981-b61a-16879d2dddghf",
				Word:          "territory",
				TermFrequency: 2.27923284424328,
			})

			for _, p := range c.given {
				index.Delete(&p)
			}

			assert.Equal(t, &c.expected, index)
		})
	}
}

func TestFind(t *testing.T) {
	cases := []TestCase[FindParams, RecordInfos]{
		{
			given: FindParams{
				Term: "what",
			},
			expected: RecordInfos{},
		},
		{
			given: FindParams{
				Term: "australian",
			},
			expected: RecordInfos{
				{
					Id:            "2e48c6df-bafa-4981-b61a-16879dcdde2a",
					TermFrequency: 3.64961844222847,
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%v", c.given), func(t *testing.T) {
			index := New()
			index.Insert(&InsertParams{
				Id:            "2e48c6df-bafa-4981-b61a-16879dcdde2a",
				Word:          "australian",
				TermFrequency: 3.64961844222847,
			})
			index.Insert(&InsertParams{
				Id:            "998c8de6-3c50-4e9e-9835-10f8d1215327",
				Word:          "australia",
				TermFrequency: 1.29513358272291,
			})
			index.Insert(&InsertParams{
				Id:            "1e44c6df-bafa-4981-b61a-16879d2dddghf",
				Word:          "territory",
				TermFrequency: 2.27923284424328,
			})

			results := index.Find(&c.given)

			assert.Equal(t, c.expected, results)
		})
	}
}
