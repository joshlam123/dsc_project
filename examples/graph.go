
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
	Value         int
	OutgoingEdges map[int]int
}



func check(e error) {
    if e != nil {
        panic(e)
    }
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
	var maxValSoFar int = 0
	var minValSoFar int = 0
	aggregatorName = os.Args[1]

	fmt.Println(os.Args[2])

	maxNoNodes, _ := strconv.Atoi(os.Args[2])

	nodeVals := make(map[int]int)
    distinctPoints := make(map[int]Node)

    for node := 1; node <= maxNoNodes; node++ {
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

    	nodeVals[node] = nodeVal

    	}

    for node := 1; node <= maxNoNodes; node++ {

	    childMap := make(map[int]int)
	// for each node, generate a random number of values and take the value from nodeVals
		numNodes := rand.Intn(maxNoNodes)
		for numNodes == 0 {
			numNodes = rand.Intn(maxNoNodes)
		}

		fmt.Println("Number of nodes %d", numNodes)
		for node2 := 1; node2 <= numNodes; node2++ {
			rndNode := rand.Intn(maxNoNodes)
			for rndNode == 0 {
				rndNode = rand.Intn(maxNoNodes)
			}

			childMap[rndNode] = nodeVals[rndNode]
		}
	    distinctPoints[node] = Node{Value:nodeVals[node], OutgoingEdges:childMap}
 	}
    
 	d2 := map[string]int{"max_value":maxValSoFar, "min_value":minValSoFar}

    fmt.Println(distinctPoints)
    fmt.Println("Max Value", maxValSoFar)
    fmt.Println("min Value", minValSoFar)

    writeToJson(distinctPoints, os.Args[1], maxNoNodes)

    writeToJson(d2, "solutions", maxNoNodes)


}