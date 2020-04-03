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
	Flag  bool
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
	outQueue        map[int][]float64 // worker ID to outQueue map
	superstep       int
	ActiveVerts     []int
}

func printGraphReader(gr graphReader) {
	fmt.Println("Node ID: ", gr.Info.NodeID)
	fmt.Println("-------------")
	fmt.Println("# Vertices:", gr.Info.NumVertices)

	fmt.Println("------Vertex values-----")
	for vID, vert := range gr.Vertices {
		fmt.Println("Value of Vertex: ", vID, ": ", vert.Value)
	}
	fmt.Println("-------------")

	if len(gr.ActiveNodes) != 0 {
		fmt.Println("Partitions:", gr.ActiveNodes[gr.Info.NodeID].PartitionList)
	}
	fmt.Println("-------------")

	//for partID, nodeID := range gr.PartitionToNode{
	//	fmt.Println("Partition ", partID, "is in node: ", nodeID)
	//}

	//for vID, val := range(gr.Vertices){
	//	fmt.Println("Value of Vertex ", vID, ": ", val.Value)
	//}
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
