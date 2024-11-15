package read

import (
	"bufio"
	"bytes"
	"log"
	"os"
	"slices"
)

func Read(inputPath string, lineTrimmer func(line []byte) []byte, fieldFunc func([]byte) [][]byte, fieldTrimmer func([]byte) []byte, lineFieldProcessorFunc func(input [][]byte, last bool)) {
	file := OpenFileLogFatal(inputPath)
	defer CloseFile(file)
	scanner := bufio.NewScanner(file)
	scanWithLastLineAwareness(scanner, lineTrimmer, fieldFunc, fieldTrimmer, lineFieldProcessorFunc)
}

func scan(scanner *bufio.Scanner, lineTrimmer func(line []byte) []byte, fieldFunc func([]byte) [][]byte, fieldTrimmer func([]byte) []byte, lineFieldProcessorFunc func(input [][]byte, last bool)) {
	for scanner.Scan() {
		line := scanner.Bytes()
		trimmedLine := line
		if lineTrimmer != nil {
			trimmedLine = lineTrimmer(line)
		}
		readLine(trimmedLine, fieldFunc, fieldTrimmer, lineFieldProcessorFunc, false)
	}
}

func scanWithLastLineAwareness(scanner *bufio.Scanner, lineTrimmer func(line []byte) []byte, fieldFunc func([]byte) [][]byte, fieldTrimmer func([]byte) []byte, lineFieldProcessorFunc func(input [][]byte, last bool)) {
	for ok := scanner.Scan(); ok; {
		line := scanner.Bytes()
		trimmedLine := line
		if lineTrimmer != nil {
			trimmedLine = lineTrimmer(line)
		}
		var l = make([]byte, len(trimmedLine))
		copy(l, trimmedLine)
		ok = scanner.Scan()
		last := !ok
		readLine(l, fieldFunc, fieldTrimmer, lineFieldProcessorFunc, last)
	}
}

func readLine(line []byte, fieldFunc func([]byte) [][]byte, fieldTrimmer func([]byte) []byte, lineFieldProcessorFunc func(fields [][]byte, last bool), lastLine bool) {
	if len(line) == 0 {
		return
	}
	fields := extractFields(line, fieldFunc)
	trimmedFields := trimUnnecessaryChars(fields, fieldTrimmer)
	lineFieldProcessorFunc(trimmedFields, lastLine)
}

func extractFields(line []byte, fieldFunc func([]byte) [][]byte) [][]byte {
	var fields [][]byte
	if fieldFunc != nil {
		fields = fieldFunc(line)
	} else {
		fields = [][]byte{line}
	}
	return fields
}

func trimUnnecessaryChars(lineFields [][]byte, fieldTrimmer func([]byte) []byte) [][]byte {
	res := make([][]byte, 0)
	for _, f := range lineFields {
		trimmedField := f
		if fieldTrimmer != nil {
			trimmedField = fieldTrimmer(f)
		}
		if len(trimmedField) == 0 {
			continue
		}
		res = append(res, trimmedField)
	}
	return res
}

func IsExcludedCharacter(allExcludedChars []byte) func(r rune) bool {
	return func(r rune) bool {
		return bytes.ContainsRune(allExcludedChars, r)
	}
}

func FindLinesContainingByteSequences(file *os.File, searchFor ...BSeq) *SearchResultPool {
	search := newSearch(searchFor...)
	fanIn := make(chan *SearchResult)
	searchResults := make([]*SearchResult, 0)
	resultChannels := fanOutFinders(file, search)
	go fanInResults(resultChannels, fanIn)
	for result := range fanIn {
		searchResults = append(searchResults, result)
	}
	return &SearchResultPool{searchResults, search}
}

type SearchResultPool struct {
	Results   []*SearchResult
	SearchFor *Search
}

type Search []BSeq

func newSearch(ws ...BSeq) *Search {
	slices.SortFunc(ws, cmp)
	search := Search(ws)
	return &search
}

func (search *Search) Index(target BSeq) (int, bool) {
	return slices.BinarySearchFunc(*search, target, cmp)
}

type Extractor func(*SearchResultPool)

func (extract *Extractor) From(pool *SearchResultPool) {
	(*extract)(pool)
}

func fanOutFinders(file *os.File, bs *Search) []chan *SearchResult {
	resultChannels := make([]chan *SearchResult, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lineResult := make(chan *SearchResult)
		resultChannels = append(resultChannels, lineResult)
		go findFirstInstances(scanner.Bytes(), bs, lineResult)
	}
	return resultChannels
}

func fanInResults(resultChannels []chan *SearchResult, fanIn chan<- *SearchResult) {
	for _, result := range resultChannels {
		for r := range result {
			fanIn <- r
		}
	}
	close(fanIn)
}

type SearchResult struct {
	B      []byte
	Result map[int][]int
}

type BSeq []byte

func (s *BSeq) String() string {
	return string(*s)
}

func (s *BSeq) less(other BSeq) bool {
	if cmp(*s, other) == -1 {
		return true
	}
	return false
}

func cmp(s BSeq, o BSeq) int {
	greater := len(s) > len(o)
	maxChar := minInt(len(s), len(o), greater)
	for i := 0; i < maxChar; i++ {
		if (s)[i] > o[i] {
			return 1
		} else if (s)[i] < o[i] {
			return -1
		}
	}
	if greater {
		return 1
	} else if len(s) < len(o) {
		return -1
	}
	return 0
}

func minInt(a, b int, greater bool) int {
	if greater {
		return b
	}
	return a
}

func (s *BSeq) equals(o BSeq) bool {
	for i, b := range *s {
		if b != o[i] {
			return false
		}
	}
	return true
}

func findFirstInstances(input []byte, byteSequences *Search, result chan<- *SearchResult) {
	if res, ok := IndicesOfFirstInstances(input, byteSequences); ok {
		result <- res
	}
	close(result)
}

func IndicesOfFirstInstances(input []byte, searchFor *Search) (*SearchResult, bool) {
	mp, ok := Find(input).FirstInstances(searchFor)
	return newSearchResult(input, mp), ok
}

func newSearchResult(input []byte, mp map[int][]int) *SearchResult {
	return &SearchResult{input, mp}
}

type Find []byte

func (f Find) FirstInstances(searchSequences *Search) (map[int][]int, bool) {
	return IndexOfFirstInstance(f, searchSequences)
}

func IndexOfFirstInstance(b []byte, searchFor *Search) (map[int][]int, bool) {
	const nothingFound = false
	const found = true
	presentIdentifiers := make(map[int][]int)
	for k, sep := range *searchFor {
		index := bytes.Index(b, sep)
		if index != -1 {
			appendElementToKListOfMap(presentIdentifiers, k, index)
		}
	}
	if len(presentIdentifiers) == 0 {
		return nil, nothingFound
	}
	return presentIdentifiers, found
}

func appendElementToKListOfMap(m map[int][]int, k, element int) {
	list := m[k]
	if list == nil {
		list = make([]int, 0)
	}
	list = append(list, element)
	m[k] = list
}

func IndexAllInstances(b, sep []byte) []int {
	indices := make([]int, 0)
	offset := 0
	for len(b) > len(sep) {
		index := bytes.Index(b, sep)
		indices = append(indices, index+offset)
		b = b[index+len(sep):]
		offset += index + len(sep)
	}
	return indices
}

func OpenFileLogFatal(path string) *os.File {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	return file
}

func CloseFile(file *os.File) {
	func(file *os.File) {
		err := file.Close()
		if err != nil {
		}
	}(file)
}
