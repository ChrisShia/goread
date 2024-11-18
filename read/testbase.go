package read

import "testing"

func equalMap[T int | string | float64](actualMap, expectedMap map[int][]T, t *testing.T) bool {
	if len(actualMap) != len(expectedMap) {
		t.Errorf("got map length %v, want %v", len(actualMap), len(expectedMap))
		return false
	}
	for k, actualList := range actualMap {
		expectedList, ok := expectedMap[k]
		if !ok {
			t.Errorf("got list %v from actual result, want %v", actualList, expectedList)
			return false
		}
		if !equalList(actualList, expectedList) {
			t.Errorf("got %v with key %v, want %v", actualList, k, expectedList)
			return false
		}
	}
	return true
}

func equalList[T int | string | float64 | byte](l1, l2 []T) bool {
	if len(l1) != len(l2) {
		return false
	}
	if len(l1) == 0 {
		return false
	}
	if len(l2) == 0 {
		return false
	}
	for i := 0; i < len(l1); i++ {
		if l1[i] != l2[i] {
			return false
		}
	}
	return true
}

func bSeq(sequences [][]byte) []BSeq {
	var ws = make([]BSeq, 0)
	for _, s := range sequences {
		ws = append(ws, s)
	}
	return ws
}

func search(sequences ...[]byte) *Search {
	return newSearch(bSeq(sequences)...)
}

func equalPoolResults(actualPool, expectedPool *SearchResultPool, t *testing.T) {
	if len(actualPool.Results) != len(expectedPool.Results) {
		t.Errorf("got %v, want %v pool results", len(actualPool.Results), expectedPool.Results)
	}
	for _, sr := range actualPool.Results {
		expectedSr, found := resultFromPool(expectedPool, sr.B)
		if !found {
			t.Errorf("got result %v from actual result, was not found in expected results", sr.B)
		}
		if !equalMap(sr.Result, expectedSr.Result, t) {
			t.Errorf("got %v, want %v", sr.Result, expectedSr.Result)
		}
	}
}

func resultFromPool(result *SearchResultPool, b []byte) (SearchResult, bool) {
	notFound := false
	found := true
	for _, sr := range result.Results {
		if equalList(sr.B, b) {
			return *sr, found
		}
	}
	return SearchResult{}, notFound
}

func createTestResultPool(search *Search, sr ...*SearchResult) *SearchResultPool {
	results := make([]*SearchResult, 0)
	for _, r := range sr {
		results = append(results, r)
	}
	return &SearchResultPool{Results: results, SearchFor: search}
}

func createTestPoolForReadInputTestFile() *SearchResultPool {
	searchFor := search([]byte("Time"), []byte("User"), []byte("Distance"))
	res1 := &SearchResult{B: []byte("Time: 123465, Distance: 8093745, User: 092374, Distance: 8093745, Distance: 8093745"), Result: map[int][]int{1: {0}, 2: {33}, 0: {14, 47, 66}}}
	res2 := &SearchResult{B: []byte("Distance: 8093745, Distance: 8093745, User: 092374"), Result: map[int][]int{2: {38}, 0: {0, 19}}}
	res3 := &SearchResult{B: []byte("Time: 123465, Distance: 8093745, User: 092374, Time: 123465, Distance: 8093745, User: 092374"), Result: map[int][]int{1: {0, 47}, 2: {33, 80}, 0: {14, 61}}}
	res4 := &SearchResult{B: []byte("234576Time: 123465"), Result: map[int][]int{1: {6}}}
	res5 := &SearchResult{B: []byte("User: 092374, Distance: 8093745, User: 092374, Distance: 8093745, Time: 123465, Time: 123465"), Result: map[int][]int{1: {66, 80}, 2: {0, 33}, 0: {14, 47}}}
	res6 := &SearchResult{B: []byte("   random words 3456 Time: Distance: 8093745,"), Result: map[int][]int{1: {21}, 0: {27}}}
	return createTestResultPool(searchFor, res1, res2, res3, res4, res5, res6)
}

func createTestPoolForExtractLinesContainingSearchSequences() *SearchResultPool {
	searchFor := search([]byte("Time"), []byte("User"), []byte("Distance"))
	res1 := &SearchResult{B: []byte("Time: 123465, Distance: 8093745, User: 092374, Distance: 8093745, Distance: 8093745"), Result: map[int][]int{1: {0}, 2: {33}, 0: {14}}}
	res2 := &SearchResult{B: []byte("Distance: 8093745, Distance: 8093745, User: 092374"), Result: map[int][]int{2: {38}, 0: {0}}}
	res3 := &SearchResult{B: []byte("Time: 123465, Distance: 8093745, User: 092374, Time: 123465, Distance: 8093745, User: 092374"), Result: map[int][]int{1: {0}, 2: {33}, 0: {14}}}
	res4 := &SearchResult{B: []byte("234576Time: 123465"), Result: map[int][]int{1: {6}}}
	res5 := &SearchResult{B: []byte("User: 092374, Distance: 8093745, User: 092374, Distance: 8093745, Time: 123465, Time: 123465"), Result: map[int][]int{1: {66}, 2: {0}, 0: {14}}}
	res6 := &SearchResult{B: []byte("   random words 3456 Time: Distance: 8093745,"), Result: map[int][]int{1: {21}, 0: {27}}}
	return createTestResultPool(searchFor, res1, res2, res3, res4, res5, res6)
}
