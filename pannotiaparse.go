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
	RowArray, ColArray, DataArray, ColCnt []int32
}

func (a *CsrArraysT) freeArrays() {
	if !reflect.ValueOf(a.RowArray).IsNil() {
		a.RowArray = a.RowArray[:0]
	}
	if !reflect.ValueOf(a.ColArray).IsNil() {
		a.ColArray = a.ColArray[:0]
	}
	if !reflect.ValueOf(a.DataArray).IsNil() {
		a.DataArray = a.DataArray[:0]
	}
	if !reflect.ValueOf(a.ColCnt).IsNil() {
		a.ColCnt = a.ColCnt[:0]
	}
}

type ellArraysT struct {
	maxHeight, numNodes int
	ColArray, dataArray []int
}

type cooedgetuple struct {
	row, col, val int32
}

type Cooedgetuples []cooedgetuple

func (s Cooedgetuples) Len() int { return len(s) }
func (s Cooedgetuples) Less(i, j int) bool {
	return s[i].row < s[j].row
}
func (s Cooedgetuples) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func ParseMetis(tmpchar string, pNumNodes, pNumEdges *int, directed bool) *CsrArraysT {
	cnt := 0
	numEdges := 0
	numNodes := 0
	var colCnt []int32

	var tupleArray []cooedgetuple
	file, err := os.Open(tmpchar)
	// file, err := os.Open("/path/to/file.txt")
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

			print("Read from file: num_nodes = %d, num_edges = %d\n", numNodes, numEdges)
			tupleArray = make([]cooedgetuple, numEdges)
		} else { //from the second line
			var punctuation = []rune{'.', '-', ',', ' '}

			words := Create(line, punctuation)
			for _, pch := range words {
				// fmt.Println(pch)
				head := lineno
				tail, _ := strconv.Atoi(pch)
				if tail <= 0 {
					break
				}

				if tail == head {
					print("reporting self loop: %d, %d\n", lineno+1, lineno)
				}

				temp.row = int32(head - 1)
				temp.col = int32(tail - 1)
				temp.val = int32(weight)

				colCnt[head-1]++
				cnt++
				tupleArray[cnt] = temp

			}
		}

		lineno++

	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	RowArray := make([]int32, numNodes+1)
	ColArray := make([]int32, numEdges)
	dataArray := make([]int32, numEdges)

	rowCnt := 0
	prev := int32(-1)
	var idx int
	for idx = 0; idx < numEdges; idx++ {
		curr := tupleArray[idx].row
		if curr != prev {
			rowCnt++
			RowArray[rowCnt] = int32(idx)
			prev = curr
		}
		ColArray[idx] = tupleArray[idx].col
		dataArray[idx] = tupleArray[idx].val

	}
	RowArray[rowCnt] = int32(idx)

	csr := new(CsrArraysT)
	csr.RowArray = make([]int32, *pNumNodes+1)
	csr.ColArray = make([]int32, *pNumEdges)
	csr.DataArray = make([]int32, *pNumEdges)
	csr.ColCnt = make([]int32, *pNumEdges)

	csr.RowArray = RowArray
	csr.ColArray = ColArray
	csr.DataArray = dataArray
	csr.ColCnt = colCnt
	return csr
}

func parseCOO(tmpchar string, pNumNodes, pNumEdges *int, directed bool) *CsrArraysT {
	cnt := 0
	numNodes := 0
	numEdges := 0
	var sp [2]byte
	var a, p byte

	var tupleArray []cooedgetuple

	file, err := os.Open(tmpchar)
	// file, err := os.Open("/path/to/file.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineno := 0
	for scanner.Scan() {
		var head, tail int
		weight := 0
		var temp cooedgetuple

		line := scanner.Text()
		switch line[0] {
		case 'c':
			break
		case 'p':
			fmt.Sscanf(line, "%c %s %d %d", &p, sp, pNumNodes, pNumEdges)
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
			break
		case 'a':
			fmt.Sscanf(line, "%c %d %d %d", &a, &head, &tail, &weight)
			if tail == head {
				fmt.Printf("reporting self loop\n")
			}
			temp.row = int32(head - 1)
			temp.col = int32(tail - 1)
			temp.val = int32(weight)
			tupleArray[cnt] = temp
			cnt += 1
			if !directed {
				temp.row = int32(tail - 1)
				temp.col = int32(head - 1)
				temp.val = int32(weight)
				tupleArray[cnt] = temp
				cnt += 1
			}

			break
		default:
			fmt.Printf("Error! existing loop!\n")
			break
		}
		lineno++
	}
	sort.Stable(Cooedgetuples(tupleArray))

	row_array := make([]int32, numNodes+1)
	col_array := make([]int32, numEdges)
	data_array := make([]int32, numEdges)

	row_cnt := 0
	prev := int32(-1)
	var idx int
	for idx = 0; idx < numEdges; idx++ {
		curr := tupleArray[idx].row
		if curr != prev {
			row_array[row_cnt] = int32(idx)
			row_cnt += 1
			prev = curr
		}
		col_array[idx] = tupleArray[idx].col
		data_array[idx] = tupleArray[idx].val
	}
	row_array[row_cnt] = int32(idx)

	tupleArray = nil //free

	csr := new(CsrArraysT)
	csr.RowArray = make([]int32, *pNumNodes+1)
	csr.ColArray = make([]int32, *pNumEdges)
	csr.DataArray = make([]int32, *pNumEdges)
	csr.ColCnt = make([]int32, *pNumEdges)

	csr.RowArray = row_array
	csr.ColArray = col_array
	csr.DataArray = data_array
	//csr.ColCnt = col
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
