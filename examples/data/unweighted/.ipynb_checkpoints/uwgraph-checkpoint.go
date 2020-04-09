
package main

import (
		"fmt"
		"math/rand"
		"strconv"
		"os"
		"time"
		"encoding/json"
		)

// vertex of a graph which has id, value, outgoingEdges
type Node struct {
	Name 		string
	Value       int
}

type Vertice struct{
	VerticeId int
	Value 	  int
}

type Info struct {
	Name string
	NumVertices int
}

func check(e error) {
    if e != nil {
        panic(e)
    }
}

func stringInSlice(a int, list []int) bool {
    for _, b := range list {
        if b == a {
            return false
        }
    }
    return true
}

func writeToJson(jsonFile interface{}, name string, size int){
	jsonString, err := json.Marshal(jsonFile)
    fmt.Println(err)

	file, err := os.Create("./"+name+strconv.Itoa(size)+".json")

	if err != nil {
	panic(err)
	}
	defer file.Close()
	file.Write(jsonString)
	file.Close()
	fmt.Println("JSON data written to ", file.Name())
}

var aggregatorName string

// main takes in two command line arguements: a fileName and an aggregatorName. The file must be stored in a textfile 
// in the examples directory as this script will read the file name from there. aggregatorName will be passed to the master
func main() {
	// make sure graph is strongly connected 

	var maxValSoFar int = 0
	var minValSoFar int = 0
	aggregatorName = os.Args[1]
	fmt.Println(os.Args[2])

	maxNoNodes, _ := strconv.Atoi(os.Args[2])

	nodeVals := make(map[int]Node)
    // vertices 
    for node := 0; node < maxNoNodes; node++ {
    	// generate the nodevalues first
    	rand.Seed(time.Now().UnixNano())

    	// turn this on only if the algorithm can handle negative nodes
    	// nodeVal = rand.Intn(maxNoNodes - (-maxNoNodes)) + (-maxNoNodes)
    	nodeVal := rand.Intn(maxNoNodes)

    	if nodeVal > maxValSoFar {
    		maxValSoFar = nodeVal
    	} else if nodeVal < minValSoFar {
    		minValSoFar = nodeVal
    	}

    	nodeVals[node] = Node{Name:strconv.Itoa(node), Value:nodeVal}
    }


    totalEdgeMap := make(map[int][]Vertice)

    // distinctPoints := make(map[string])
    for node := 0; node < maxNoNodes; node++ {

	// for each node, generate a random number of nodes
		numNodes := rand.Intn(maxNoNodes)
		for numNodes == 0 {
			numNodes = rand.Intn(maxNoNodes)
		}
		fmt.Println("Number of nodes %d", numNodes)

	// for each of the nodes 
	traversedEdges := []int{}
	traversedEdges = append(traversedEdges, node)

		for node2 := 0; node2 < numNodes; node2++ {

			// for each randomly generated node, check whether it has already been generated
			rndNode := rand.Intn(maxNoNodes)
			for stringInSlice(rndNode, traversedEdges) == false {
				rndNode = rand.Intn(maxNoNodes)
			}

			// for  {
			// 	rndNode = rand.Intn(maxNoNodes)
			// }

			totalEdgeMap[node] = append(totalEdgeMap[node], Vertice{VerticeId:rndNode, Value:nodeVals[rndNode].Value})
			traversedEdges = append(traversedEdges, rndNode)
		}
 	}
 	
    infoInt := Info{Name:aggregatorName, NumVertices:maxNoNodes}
 	// infoMap := map[string]interface{}{"info":interface{}{"name":aggregatorName, "numVertices":numNodes}}

 	generatedJson := map[string]interface{}{"info": infoInt, "vertices":nodeVals, "edges":totalEdgeMap}


 	d2 := map[string]int{"max_value":maxValSoFar, "min_value":minValSoFar}

    fmt.Println(generatedJson)
    fmt.Println("Max Value", maxValSoFar)
    fmt.Println("min Value", minValSoFar)

    writeToJson(generatedJson, os.Args[1], maxNoNodes)

    writeToJson(d2, "solutions"+os.Args[1], maxNoNodes)


}