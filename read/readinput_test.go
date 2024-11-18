package read

import (
	"testing"
)

func TestContainsAnyIdentifier(t *testing.T) {
	s := search([]byte("Time"), []byte("User"), []byte("Distance"))
	var tests = []struct {
		name        string
		input       string
		s           *Search
		expectedLen int
		expectedMap map[int][]int
	}{
		{"Test_ContainsAnyIdentifier_only_Time_id_present", "Time: 123465", s, 1, map[int][]int{1: {0}}},
		{"test all identifiers present once", "Time: 123465, Distance: 8093745, User: 092374", s, 3, map[int][]int{1: {0}, 2: {33}, 0: {14}}},
	}
	for _, tt := range tests {
		ans, _ := IndexOfFirstInstance([]byte(tt.input), tt.s)
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectedLen != len(ans) {
				t.Errorf("got %v, expected %v", len(ans), tt.expectedLen)
			}
			if !equalMap(ans, tt.expectedMap, t) {
				t.Errorf("got %v, expected %v", ans, tt.expectedMap)
			}

		})
	}
}

func TestFindLinesOfFileContainingSequences(t *testing.T) {
	s := search([]byte("Time"), []byte("User"), []byte("Distance"))
	tests := []struct {
		name          string
		inputFilePath string
		search        *Search
		expectedPool  *SearchResultPool
	}{
		{"test extraction of lines containing search sequences", "readinput_test.txt", s, createTestPoolForExtractLinesContainingSearchSequences()},
	}
	for _, tt := range tests {
		file := OpenFileLogFatal(tt.inputFilePath)
		actualPool := FindLinesContainingByteSequences(file, *tt.search...)
		CloseFile(file)
		t.Run(tt.name, func(t *testing.T) {
			equalPoolResults(actualPool, tt.expectedPool, t)
		})
	}
}

func TestIndexAllInstances(t *testing.T) {
	identifiers := make([][]byte, 0)
	identifiers = append(identifiers, []byte("Time"), []byte("User"), []byte("Distance"))
	tests := []struct {
		name     string
		input    []byte
		sep      []byte
		expected []int
	}{
		{
			"test TimeOtherWordsTime",
			[]byte("TimeOtherWordsTime"), []byte("Time"), []int{0, 14}},
		{
			"test all indices of Distance in: Time: 123465, Distance: 8093745, User: 092374, Distance: 8093745, Distance: 8093745",
			[]byte("Time: 123465, Distance: 8093745, User: 092374, Distance: 8093745, Distance: 8093745"),
			[]byte("Distance"),
			[]int{14, 47, 66},
		},
	}
	for _, tt := range tests {
		ans := IndexAllInstances(tt.input, tt.sep)
		t.Run(tt.name, func(t *testing.T) {
			if !equalList(ans, tt.expected) {
				t.Errorf("got %v, want %v", ans, tt.expected)
			}
		})
	}
}

func TestLocateSequencesAccordingToIndexer(t *testing.T) {
	tests := []struct {
		name          string
		input         []byte
		s             *Search
		expectedFound bool
		expectedMap   map[int][]int
	}{
		{
			"test locate all instances of time user distance in input",
			[]byte("Time: 123465, Distance: 8093745, User: 092374, Distance: 8093745, Distance: 8093745"),
			search([]byte("Time"), []byte("User"), []byte("Distance")),
			true,
			map[int][]int{1: {0}, 2: {33}, 0: {14, 47, 66}},
		},
		{
			"test no instances present",
			[]byte("Time: 123465, Distance: 8093745, User: 092374, Distance: 8093745, Distance: 8093745"),
			search([]byte("Something"), []byte("Else")),
			false,
			map[int][]int{},
		},
	}
	for _, tt := range tests {
		indexMap, found := IndexOfAllInstances(tt.input, tt.s)
		t.Run(tt.name, func(t *testing.T) {
			if found != tt.expectedFound {
				t.Errorf("got %v, expected %v", found, tt.expectedFound)
			}
			if !equalMap(indexMap, tt.expectedMap, t) {
				t.Errorf("got %v, expected %v", indexMap, tt.expectedMap)
			}
		})
	}
}
