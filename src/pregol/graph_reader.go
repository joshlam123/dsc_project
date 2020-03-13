package pregol

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type ActiveNode struct {
	ip            string
	partitionList []int
}

type infoReader struct {
	Name          string
	NumVertices   int
	NumPartitions int
}

type vertexReader struct {
	Name  string
	Value float64
}

type edgeReader struct {
	VerticeID int
	Value     float64
}

type graphReader struct {
	Info        infoReader
	Vertices    map[int]vertexReader
	Edges       map[int][]edgeReader
	ActiveNodes []ActiveNode
}

func getGraphFromFile(graphFile string) *graphReader {
	jsonFile, err := os.Open(graphFile)

	if err != nil {
		panic(err)
	}

	defer jsonFile.Close()

	byteVal, _ := ioutil.ReadAll(jsonFile)
	var g graphReader
	json.Unmarshal(byteVal, &g)
	fmt.Println(g)
	return &g
}

func getGraphFromJSONByte(jsonBytes []byte) *graphReader {
	var g graphReader
	json.Unmarshal(jsonBytes, &g)
	fmt.Println(g)
	return &g
}
