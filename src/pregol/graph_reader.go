package pregol

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type activeNode struct {
	IP            string
	PartitionList []int
}

type infoReader struct {
	Name          string
	NumVertices   int
	NumPartitions int
	NodeID        int
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
	Info            infoReader
	Vertices        map[int]vertexReader
	Edges           map[int][]edgeReader
	PartitionToNode map[int]int
	ActiveNodes     []activeNode
}

func printGraphReader(gr graphReader) {
	fmt.Println("Graph ID: ", gr.Info.NodeID)
	fmt.Println("# Vertices:", gr.Info.NumVertices)
	fmt.Println("Graph contains vertices ---")
	for vID, vert := range gr.Vertices {
		fmt.Println("Vertice", vID, ": ", vert.Value)
	}
}

func newGraphReader() graphReader {
	gR := graphReader{}
	gR.Vertices = make(map[int]vertexReader)
	gR.Edges = make(map[int][]edgeReader)
	return gR
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
	return &g
}

func getGraphFromJSONByte(jsonBytes []byte) *graphReader {
	var g graphReader
	json.Unmarshal(jsonBytes, &g)
	return &g
}

func getJSONByteFromGraph(gR graphReader) []byte {
	b, err := json.Marshal(gR)
	if err != nil {
		panic(err)
	}
	return b
}
