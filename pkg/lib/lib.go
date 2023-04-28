package lib

import (
	"math"
)

func Count(tokens []string) map[string]int {
	tokensCounter := make(map[string]int, len(tokens))
	for _, token := range tokens {
		tokensCounter[token]++
	}
	return tokensCounter
}

func BM25(tf float64, matchingDocsCount int, fieldLength int, avgFieldLength float64, docsCount int, k float64, b float64, d float64) float64 {
	idf := math.Log(1 + (float64(docsCount-matchingDocsCount)+0.5)/(float64(matchingDocsCount)+0.5))
	return idf * (d + tf*(k+1)) / (tf + k*(1-b+(b*float64(fieldLength))/avgFieldLength))
}

func Paginate(offset int, limit int, sliceLength int) (int, int) {
	if offset > sliceLength {
		offset = sliceLength
	}

	end := offset + limit
	if end > sliceLength {
		end = sliceLength
	}

	return offset, end
}

func CommonPrefix(a []rune, b []rune) ([]rune, bool) {
	minLength := int(math.Min(float64(len(a)), float64(len(b))))
	commonPrefix := make([]rune, 0, minLength)
	equal := len(a) == len(b)

	for i := 0; i < minLength; i++ {
		if a[i] != b[i] {
			equal = false
			break
		}
		commonPrefix = append(commonPrefix, a[i])
	}

	return commonPrefix, equal
}

func BoundedLevenshtein(a []rune, b []rune, tolerance int) (int, bool) {
	distance := boundedLevenshtein(a, b, tolerance)
	return distance, distance >= 0
}

/**
 * Inspired by:
 * https://github.com/Yomguithereal/talisman/blob/86ae55cbd040ff021d05e282e0e6c71f2dde21f8/src/metrics/levenshtein.js#L218-L340
 */
func boundedLevenshtein(a []rune, b []rune, tolerance int) int {
	// the strings are the same
	if string(a) == string(b) {
		return 0
	}

	// a should be the shortest string
	if len(a) > len(b) {
		a, b = b, a
	}

	// ignore common suffix
	lenA, lenB := len(a), len(b)
	for lenA > 0 && a[lenA-1] == b[lenB-1] {
		lenA--
		lenB--
	}

	// early return when the smallest string is empty
	if lenA == 0 {
		if lenB > tolerance {
			return -1
		}
		return lenB
	}

	// ignore common prefix
	startIdx := 0
	for startIdx < lenA && a[startIdx] == b[startIdx] {
		startIdx++
	}
	lenA -= startIdx
	lenB -= startIdx

	// early return when the smallest string is empty
	if lenA == 0 {
		if lenB > tolerance {
			return -1
		}
		return lenB
	}

	delta := lenB - lenA

	if tolerance > lenB {
		tolerance = lenB
	} else if delta > tolerance {
		return -1
	}

	i := 0
	row := make([]int, lenB)
	characterCodeCache := make([]int, lenB)

	for i < tolerance {
		characterCodeCache[i] = int(b[startIdx+i])
		row[i] = i + 1
		i++
	}

	for i < lenB {
		characterCodeCache[i] = int(b[startIdx+i])
		row[i] = tolerance + 1
		i++
	}

	offset := tolerance - delta
	haveMax := tolerance < lenB

	jStart := 0
	jEnd := tolerance

	var current, left, above, charA, j int

	// Starting the nested loops
	for i := 0; i < lenA; i++ {
		left = i
		current = i + 1

		charA = int(a[startIdx+i])
		if i > offset {
			jStart = 1
		}
		if jEnd < lenB {
			jEnd++
		}

		for j = jStart; j < jEnd; j++ {
			above = current

			current = left
			left = row[j]

			if charA != characterCodeCache[j] {
				// insert current
				if left < current {
					current = left
				}

				// delete current
				if above < current {
					current = above
				}

				current++
			}

			row[j] = current
		}

		if haveMax && row[i+delta] > tolerance {
			return -1
		}
	}

	if current <= tolerance {
		return current
	}

	return -1
}
