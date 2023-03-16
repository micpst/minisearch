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
