package read

import (
	"testing"
)

func TestContainsAnyIdentifier(t *testing.T) {
	identifiers := make([][]byte, 0)
	identifiers = append(identifiers, []byte("Time"), []byte("User"), []byte("Distance"))
	var tests = []struct {
		name        string
		input       string
		identifiers [][]byte
		expectedLen int
		expectedMap map[int][]int
	}{
		{"Test_ContainsAnyIdentifier_only_Time_id_present", "Time: 123465", identifiers, 1, map[int][]int{0: {0}}},
		{"test all identifiers present once", "Time: 123465, Distance: 8093745, User: 092374", identifiers, 3, map[int][]int{0: {0}, 1: {33}, 2: {14}}},
	}
	for _, tt := range tests {
		mp, _ := IndexOfFirstInstance([]byte(tt.input), search(tt.identifiers))
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectedLen != len(mp) {
				t.Errorf("got %v, want %v", len(mp), tt.expectedLen)
			}
			for i, m := range mp {
				for j, e := range m {
					if e != tt.expectedMap[i][j] {
						t.Errorf("got %v, want %v for identifier %s", e, tt.expectedMap[i][j], tt.identifiers[i])
					}
				}
			}
		})
	}
}

//func TestFindLinesOfFileContainingSequences(t *testing.T) {
//	identifiers := make([][]byte, 0)
//	identifiers = append(identifiers, []byte("Time"), []byte("User"), []byte("Distance"))
//	tests := []struct {
//		name          string
//		inputFilePath string
//		identifiers   [][]byte
//	}{
//		{"test file", "readinput_test.txt", identifiers},
//		//{"test file", "input/test/scaninput_test.txt", identifiers},
//	}
//	for _, tt := range tests {
//		file := OpenFileLogFatal(tt.inputFilePath)
//		results := FindLinesContainingByteSequences(file, tt.identifiers...)
//		CloseFile(file)
//		t.Run(tt.name, func(t *testing.T) {
//			for _, sr := range results {
//				for i, e := range sr.result {
//					fmt.Println(i, e)
//					//if e != tt.expectedMap[i][j] {
//					//	t.Errorf("got %v, want %v for identifier %s", e, tt.expectedMap[i][j], tt.identifiers[i])
//					//}
//				}
//			}
//		})
//	}
//}

func TestIndexAllInstances(t *testing.T) {
	identifiers := make([][]byte, 0)
	identifiers = append(identifiers, []byte("Time"), []byte("User"), []byte("Distance"))
	tests := []struct {
		name     string
		input    []byte
		sep      []byte
		expected []int
	}{
		{"test TimeOtherWordsTime", []byte("TimeOtherWordsTime"), []byte("Time"), []int{0, 14}},
		//{"test file", "input/test/scaninput_test.txt", identifiers},
	}
	for _, tt := range tests {
		ans := IndexAllInstances(tt.input, tt.sep)
		t.Run(tt.name, func(t *testing.T) {
			if !equal(ans, tt.expected) {
				t.Errorf("got %v, want %v", ans, tt.expected)
			}
		})
	}
}

func equal(l1, l2 []int) bool {
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

func words(sequences [][]byte) []BSeq {
	var ws = make([]BSeq, 0)
	for _, s := range sequences {
		ws = append(ws, s)
	}
	return ws
}

func search(sequences [][]byte) *Search {
	return newSearch(words(sequences)...)
}
