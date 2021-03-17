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
	"strings"
)

func doCompare(elem1, elem2 cooedgetuple) bool {
	return elem1.row < elem2.row
}

type CsrArraysT struct {
	ColArray, DataArray, ColCnt                []int32
	RowArray, RowArrayT, ColArrayT, DataArrayT []int32
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
	ColArray, DataArray []int
}

type cooedgetuple struct {
	row, col, val int32
}

func transform(tupleArray *[]cooedgetuple, numEdges int, rowArray, colArray, dataArray *[]int32) {
	rowCnt := 0
	prev := int32(-1)
	var idx int
	for idx = 0; idx < numEdges; idx++ {
		curr := (*tupleArray)[idx].row
		if curr != prev {
			(*rowArray)[rowCnt] = int32(idx)
			rowCnt++
			prev = curr
		}

		(*colArray)[idx] = (*tupleArray)[idx].col
		(*dataArray)[idx] = (*tupleArray)[idx].val

	}
	(*rowArray)[rowCnt] = int32(idx)

}

func ParseMetis(tmpchar string, pNumNodes, pNumEdges *int, directed bool) *CsrArraysT {
	cnt := 0
	numEdges := 0
	numNodes := 0
	var ColCnt []int32

	var tupleArray []cooedgetuple
	file, err := os.Open(tmpchar)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	fmt.Printf("Opening file: %s\n", tmpchar)

	scanner := bufio.NewScanner(file)
	lineno := uint(0)
	for scanner.Scan() {
		weight := 0
		var temp cooedgetuple

		line := scanner.Text()
		if line[0] == '%' {
			continue
		}
		if lineno == 0 {
			fmt.Sscanf(line, "%d %d", pNumNodes, pNumEdges)
			ColCnt = make([]int32, *pNumNodes)

			for i := 0; i < *pNumNodes; i++ {
				ColCnt[i] = 0
			}

			if !directed {
				*pNumEdges *= 2
				fmt.Printf("This is an undirected graph\n")
			} else {
				fmt.Printf("This is a directed graph\n")
			}
			numNodes = *pNumNodes
			numEdges = *pNumEdges

			fmt.Printf("Read from file: num_nodes = %d, num_edges = %d\n", numNodes, numEdges)
			tupleArray = make([]cooedgetuple, numEdges)
		} else if lineno > 0 { //from the second line

			// Although pannotia's parse.cpp used " ,.-" for strtok(),
			// it seems only " " is used in given graph files
			words := strings.Split(line, " ")
			for _, pch := range words {
				tail, _ := strconv.Atoi(pch)
				head := int(lineno)
				if tail <= 0 {
					break
				}

				if tail == head {
					print("reporting self loop: %d, %d\n", lineno+1, lineno)
				}

				temp.row = int32(head) - 1
				temp.col = int32(tail) - 1
				temp.val = int32(weight)

				ColCnt[head-1]++

				tupleArray[cnt] = temp
				cnt++
			}
		}

		lineno++

	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	// sort.Slice(tupleArray[:numEdges], func(i, j int) bool {
	// 	return tupleArray[i].row < tupleArray[j].row
	// })

	RowArray := make([]int32, numNodes+1)
	ColArray := make([]int32, numEdges)
	DataArray := make([]int32, numEdges)

	rowCnt := 0
	prev := int32(-1)
	var idx int
	for idx = 0; idx < numEdges; idx++ {
		curr := tupleArray[idx].row
		if curr != prev {
			RowArray[rowCnt] = int32(idx)
			rowCnt++
			prev = curr
		}
		ColArray[idx] = tupleArray[idx].col
		DataArray[idx] = tupleArray[idx].val
	}
	RowArray[rowCnt] = int32(idx)

	csr := new(CsrArraysT)
	csr.RowArray = make([]int32, numNodes+1)
	csr.ColArray = make([]int32, numEdges)
	csr.DataArray = make([]int32, numEdges)
	csr.ColCnt = make([]int32, *pNumNodes)
	csr.RowArray = RowArray
	csr.ColArray = ColArray
	csr.DataArray = DataArray
	csr.ColCnt = ColCnt
	return csr
}

func ParseCOO(tmpchar string, pNumNodes, pNumEdges *int, directed bool) *CsrArraysT {
	// ParseCOO() in BC.h
	cnt := 0
	cnt1 := 0
	numEdges := 0
	numNodes := 0
	a := 'x'
	p := 'x'
	sp := "xx"

	var tupleArray []cooedgetuple
	var tupleArrayT []cooedgetuple
	file, err := os.Open(tmpchar)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	fmt.Printf("Opening file: %s\n", tmpchar)

	scanner := bufio.NewScanner(file)
	lineno := uint(0)
	for scanner.Scan() {
		var head, tail, weight int

		line := scanner.Text()
		switch line[0] {
		case 'c':
			{
				break
			}
		case 'p':
			{
				fmt.Sscanf(line, "%c %s %d %d", &p, &sp, pNumNodes, pNumEdges)

				if !directed {
					*pNumEdges *= 2
					fmt.Printf("This is an undirected graph\n")
				} else {
					fmt.Printf("This is a directed graph\n")
				}
				numNodes = *pNumNodes
				numEdges = *pNumEdges

				fmt.Printf("Read from file: num_nodes = %d, num_edges = %d\n", numNodes, numEdges)
				tupleArray = make([]cooedgetuple, numEdges)
				tupleArrayT = make([]cooedgetuple, numEdges)
				break
			}
		case 'a':
			{
				fmt.Sscanf(line, "%c %d %d %d", &a, &head, &tail, &weight)
				if tail == head {
					fmt.Printf("reporting self loop")
				}
				var temp, temp1 cooedgetuple
				temp.row = int32(head) - 1
				temp.col = int32(tail) - 1
				temp.val = int32(weight)

				temp1.row = int32(tail) - 1
				temp1.col = int32(head) - 1
				temp1.val = int32(weight)

				tupleArray[cnt] = temp
				cnt++
				tupleArrayT[cnt1] = temp1
				cnt1++

				if !directed {
					temp.row = int32(tail) - 1
					temp.col = int32(head) - 1
					temp.val = int32(weight)

					temp1.row = int32(head) - 1
					temp1.col = int32(tail) - 1
					temp1.val = int32(weight)

					tupleArray[cnt] = temp
					cnt++
					tupleArrayT[cnt1] = temp1
					cnt1++
				}
				break
			}
		default:
			{
				fmt.Printf("exiting loop\n")
				break
			}
		}
		lineno++

		sort.Slice(tupleArray[:numEdges], func(i, j int) bool {
			return tupleArray[i].row < tupleArray[j].row
		})
		sort.Slice(tupleArrayT[:numEdges], func(i, j int) bool {
			return tupleArrayT[i].row < tupleArrayT[j].row
		})
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	RowArray := make([]int32, numNodes+1)
	ColArray := make([]int32, numEdges)
	DataArray := make([]int32, numEdges)

	RowArrayT := make([]int32, numNodes+1)
	ColArrayT := make([]int32, numEdges)
	DataArrayT := make([]int32, numEdges)

	transform(&tupleArray, numEdges, &RowArray, &ColArray, &DataArray)
	transform(&tupleArrayT, numEdges, &RowArrayT, &ColArrayT, &DataArrayT)

	// Implemented in transform()
	// rowCnt := 0
	// prev := int32(-1)
	// var idx int
	// for idx = 0; idx < numEdges; idx++ {
	// 	curr := tupleArray[idx].row
	// 	if curr != prev {
	// 		RowArray[rowCnt] = int32(idx)
	// 		rowCnt++
	// 		prev = curr
	// 	}
	// 	ColArray[idx] = tupleArray[idx].col
	// 	DataArray[idx] = tupleArray[idx].val
	// }

	// RowArray[rowCnt] = int32(idx)

	csr := new(CsrArraysT)
	csr.RowArray = make([]int32, numNodes+1)
	csr.ColArray = make([]int32, numEdges)
	csr.DataArray = make([]int32, numEdges)
	csr.RowArray = RowArray
	csr.ColArray = ColArray
	csr.DataArray = DataArray
	csr.RowArrayT = make([]int32, numNodes+1)
	csr.ColArrayT = make([]int32, numEdges)
	csr.DataArrayT = make([]int32, numEdges)
	csr.RowArrayT = RowArrayT
	csr.ColArrayT = ColArrayT
	csr.DataArrayT = DataArrayT
	return csr
}

func ParseCOOTranspose(tmpchar string, pNumNodes, pNumEdges *int, directed bool) *CsrArraysT {
	cnt := 0
	lineno := uint(0)
	numEdges := 0
	numNodes := 0
	a := 'x'
	p := 'x'
	sp := "xx"

	var tupleArray []cooedgetuple
	file, err := os.Open(tmpchar)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	fmt.Printf("Opening file: %s\n", tmpchar)

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		var head, tail, weight int

		line := scanner.Text()
		switch line[0] {
		case 'c':
			{
				break
			}
		case 'p':
			{
				fmt.Sscanf(line, "%c %s %d %d", &p, &sp, pNumNodes, pNumEdges)

				if !directed {
					*pNumEdges *= 2
					fmt.Printf("This is an undirected graph\n")
				} else {
					fmt.Printf("This is a directed graph\n")
				}
				numNodes = *pNumNodes
				numEdges = *pNumEdges

				fmt.Printf("Read from file: num_nodes = %d, num_edges = %d\n", numNodes, numEdges)
				tupleArray = make([]cooedgetuple, numEdges)
				break
			}
		case 'a':
			{
				fmt.Sscanf(line, "%c %d %d %d", &a, &head, &tail, &weight)
				if tail == head {
					fmt.Printf("reporting self loop")
				}
				var temp cooedgetuple
				temp.val = int32(weight)
				temp.row = int32(tail) - 1
				temp.col = int32(head) - 1

				tupleArray[cnt] = temp
				cnt++

				if !directed {
					temp.val = int32(weight)
					temp.row = int32(tail) - 1
					temp.col = int32(head) - 1

					tupleArray[cnt] = temp
					cnt++
				}
				break
			}
		default:
			{
				fmt.Printf("exiting loop\n")
				break
			}
		}
		lineno++
	}
	sort.SliceStable(tupleArray[:numEdges], func(i, j int) bool {
		return tupleArray[i].row < tupleArray[j].row
	})

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	RowArray := make([]int32, numNodes+1)
	ColArray := make([]int32, numEdges)
	DataArray := make([]int32, numEdges)

	// Implemented in transform()
	rowCnt := 0
	prev := int32(-1)
	var idx int
	for idx = 0; idx < numEdges; idx++ {
		curr := tupleArray[idx].row
		if curr != prev {
			RowArray[rowCnt] = int32(idx)
			rowCnt++
			prev = curr
		}
		ColArray[idx] = tupleArray[idx].col
		DataArray[idx] = tupleArray[idx].val
	}

	// RowArray[rowCnt] = int32(idx)
	for rowCnt <= numNodes {
		RowArray[rowCnt] = int32(idx)
		rowCnt++
	}

	csr := new(CsrArraysT)
	csr.RowArray = make([]int32, numNodes+1)
	csr.ColArray = make([]int32, numEdges)
	csr.DataArray = make([]int32, numEdges)
	csr.RowArray = RowArray
	csr.ColArray = ColArray
	csr.DataArray = DataArray
	return csr
}

func main() {
	fmt.Println(errors.New("testing new error"))
}
