package pannotiaparse

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"sort"
	"strconv"
)

type CsrArraysT struct {
	ColArray, dataArray, colCnt []int32
	RowArray                    []int32
}

func (a *CsrArraysT) freeArrays() {
	if !reflect.ValueOf(a.RowArray).IsNil() {
		a.RowArray = a.RowArray[:0]
	}
	if !reflect.ValueOf(a.ColArray).IsNil() {
		a.ColArray = a.ColArray[:0]
	}
	if !reflect.ValueOf(a.dataArray).IsNil() {
		a.dataArray = a.dataArray[:0]
	}
	if !reflect.ValueOf(a.colCnt).IsNil() {
		a.colCnt = a.colCnt[:0]
	}
}

type ellArraysT struct {
	maxHeight, numNodes int
	ColArray, dataArray []int
}

type cooedgetuple struct {
	row, col, val int32
}

func ParseMetis(tmpchar string, pNumNodes, pNumEdges *int, directed bool) *CsrArraysT {
	cnt := 0
	numEdges := 0
	numNodes := 0
	var colCnt []int32

	var tupleArray []cooedgetuple
	file, err := os.Open(tmpchar)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineno := 0
	for scanner.Scan() {
		weight := 0
		var temp cooedgetuple

		line := scanner.Text()
		if line[0] == '%' {
			continue
		}
		if lineno == 0 {
			fmt.Sscanf(line, "%d %d", pNumNodes, pNumEdges)
			colCnt = make([]int32, *pNumNodes)

			if !directed {
				*pNumEdges = *pNumEdges * 2
				print("This is an undirected graph\n")
			} else {
				print("This is a directed graph\n")
			}
			numNodes = *pNumNodes
			numEdges = *pNumEdges

			fmt.Printf("Read from file: num_nodes = %d, num_edges = %d\n", numNodes, numEdges)
			tupleArray = make([]cooedgetuple, numEdges)
		} else { //from the second line
			var punctuation = []rune{'.', '-', ',', ' '}

			words := Create(line, punctuation)
			for _, pch := range words {
				// fmt.Println(pch)
				head := int32(lineno)
				tail, _ := strconv.Atoi(pch)
				if tail <= 0 {
					break
				}

				if int32(tail) == head {
					print("reporting self loop: %d, %d\n", lineno+1, lineno)
				}

				temp.row = head - 1
				temp.col = int32(tail) - 1
				temp.val = int32(weight)

				colCnt[head-1]++

				tupleArray[cnt] = temp
				cnt++

			}
		}

		lineno++

	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	sort.Slice(tupleArray[:numEdges], func(i, j int) bool {
		return tupleArray[i].row < tupleArray[j].row
	})

	RowArray := make([]int32, numNodes+1)
	ColArray := make([]int32, numEdges)
	dataArray := make([]int32, numEdges)

	rowCnt := 0
	prev := -1
	var idx int
	for idx = 0; idx < numEdges; idx++ {
		curr := int(tupleArray[idx].row)
		if curr != prev {

			RowArray[rowCnt] = int32(idx)
			rowCnt++
			prev = curr
		}
		ColArray[idx] = tupleArray[idx].col
		dataArray[idx] = tupleArray[idx].val

	}
	RowArray[rowCnt] = int32(idx)

	csr := new(CsrArraysT)
	csr.RowArray = make([]int32, numNodes+1)
	csr.ColArray = make([]int32, numEdges)
	csr.dataArray = make([]int32, numEdges)
	csr.colCnt = make([]int32, *pNumNodes)
	csr.RowArray = RowArray
	csr.ColArray = ColArray
	csr.dataArray = dataArray
	csr.colCnt = colCnt
	return csr
}

// https://github.com/dannav/tokenize/blob/master/tokenize.go

//Create takes any text as string, tokenization runes, and returns a slice of string tokens, where each item in the result set are the tokenized words followed by the runes to tokenize on in order.
func Create(text string, tokenizeon []rune) []string {
	resultSet := []string{}
	textAsRune := []rune(text)
	i := 0

	for len(textAsRune) > 0 {
		r := textAsRune[i]

		if RuneIndexOf(tokenizeon, r) > -1 {
			setItem := textAsRune[:i]
			resultSet = append(resultSet, string(removePad(setItem)))
			resultSet = append(resultSet, string(textAsRune[i:i+1]))

			textAsRune = textAsRune[i+1:]
			i = 0
		}

		i++
	}

	return resultSet
}

// RuneIndexOf returns the index of a rune in a slice of runes or -1 if it doesn't exist
func RuneIndexOf(r []rune, el rune) int {
	for i, e := range r {
		if el == e {
			return i
		}
	}

	return -1
}

func removePad(r []rune) []rune {
	if len(r) > 0 {
		if r[0] == ' ' {
			r = r[1:]
		}

		if r[len(r)-1] == ' ' {
			r = r[:len(r)-2]
		}
	}

	return r
}

func main() {
	fmt.Println(errors.New("testing new error"))
}
