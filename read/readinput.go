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

func FindLinesContainingByteSequences(file *os.File, seqs ...BSeq) *SearchResultPool {
	searchFor := newSearch(seqs...)
	fanIn := make(chan *SearchResult)
	searchResults := make([]*SearchResult, 0)
	resultChannels := fanOutFinders(file, searchFor)
	go fanInResults(resultChannels, fanIn)
	for result := range fanIn {
		searchResults = append(searchResults, result)
	}
	return &SearchResultPool{searchResults, searchFor}
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

type data interface {
	int | string | byte | float64
}

type Extractor func(*SearchResultPool)

func (extract *Extractor) From(pool *SearchResultPool) {
	(*extract)(pool)
}

func newExtractor[T data](identifier BSeq, valueGetter func(inLineIndices []int, line []byte, identifier BSeq) []T) (*[]T, Extractor) {
	values := make([]T, 0)
	return &values, func(pool *SearchResultPool) {
		index, _ := pool.SearchFor.Index(identifier)
		for _, sr := range pool.Results {
			if indices, ok := sr.Result[index]; ok {
				values = append(values, valueGetter(indices, sr.B, identifier)...)
			}
		}
	}
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
	if res, ok := IndexFirstInstance(input, byteSequences); ok {
		result <- res
	}
	close(result)
}

func IndexFirstInstance(input []byte, searchFor *Search) (*SearchResult, bool) {
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

func (f Find) AllInstances(searchSequences *Search) (map[int][]int, bool) {
	return IndexOfAllInstances(f, searchSequences)
}

type Indexer func([]byte, []byte) []int

func (f *Indexer) Apply(bs, sep []byte) []int {
	return (*f)(bs, sep)
}

func IndexOfAllInstances(b []byte, searchFor *Search) (map[int][]int, bool) {
	return SequenceIndexerFunc(b, searchFor, IndexAllInstances)
}

func IndexOfFirstInstance(b []byte, searchFor *Search) (map[int][]int, bool) {
	return SequenceIndexerFunc(b, searchFor, ListWithJustFirstInstanceIndex)
}

func SequenceIndexerFunc(b []byte, searchFor *Search, indexer func(b []byte, sep []byte) []int) (map[int][]int, bool) {
	successfulSequences := LocateSequencesIndexerFunc(b, searchFor, indexer)
	return successfulSequences, len(successfulSequences) > 0
}

func LocateSequencesIndexerFunc(b []byte, searchFor *Search, indexer Indexer) map[int][]int {
	presentIdentifiers := make(map[int][]int)
	chanList := make([]chan []int, 0)
	for _, sep := range *searchFor {
		c := make(chan []int)
		chanList = append(chanList, c)
		go func(b, sep []byte, indexer Indexer, c chan []int) {
			indices := indexer.Apply(b, sep)
			if indices != nil && len(indices) > 0 {
				c <- indices
			}
			close(c)
		}(b, sep, indexer, c)
	}
	fanIn := make(chan struct {
		i int
		v []int
	})
	go func(cl []chan []int) {
		for k, ch := range cl {
			for res := range ch {
				fanIn <- struct {
					i int
					v []int
				}{k, res}
			}
		}
		close(fanIn)
	}(chanList)
	for fc := range fanIn {
		appendElementsToKListOfMap(presentIdentifiers, fc.i, fc.v...)
	}
	return presentIdentifiers
}

func ListWithJustFirstInstanceIndex(b, sep []byte) []int {
	index := bytes.Index(b, sep)
	if index == -1 {
		return nil
	}
	return []int{index}
}

func appendElementsToKListOfMap(m map[int][]int, k int, elements ...int) {
	list := m[k]
	if list == nil {
		list = make([]int, 0)
	}
	list = append(list, elements...)
	m[k] = list
}

func IndexAllInstances(b, sep []byte) []int {
	indices := make([]int, 0)
	offset := 0
	for len(b) > len(sep) {
		index := bytes.Index(b, sep)
		if index == -1 {
			break
		}
		indices = append(indices, index+offset)
		b = b[index+len(sep):]
		offset += index + len(sep)
	}
	if len(indices) == 0 {
		return nil
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
