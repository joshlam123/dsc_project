package main


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

type guiSave struct {
	CurrentIteration int
	GraphsToNodes    []graphReader
}

func printGraphReader(gr graphReader) {
	fmt.Println("Graph ID: ", gr.Info)
	fmt.Println("# Vertices:", gr.Info.NumVertices)
	fmt.Println("ActiveNodes: ", gr.ActiveNodes)
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

func getGraphFromFile(graphFile string) *guiSave {
	jsonFile, err := os.Open(graphFile)

	if err != nil {
		panic(err)
	}

	defer jsonFile.Close()

	byteVal, _ := ioutil.ReadAll(jsonFile)
	var g guiSave
	json.Unmarshal(byteVal, &g)
	return &g
}


func main() {
	// file := 'data/weighted/prob/rand20.json'
	gr := getGraphFromFile("../gui/guiSave.json")
	fmt.Println("GR ", gr.GraphsToNodes)
	// printGraphReader(*gr)
}