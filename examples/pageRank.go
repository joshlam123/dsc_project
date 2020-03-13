package main

import ("fmt"
		"../src/pregol/graph_reader"
		"os"
		"math")

// type prGraph Struct {

// }

func getMessages(incomingMessages Message) {
	make(map[int]map[int]int)
}

func main() {
	// or 10e-5
	var errorThres float64 = os.Args[1]
	graph := graph_reader.getGraphFromFile("./data/prob/randM20.json")
	fmt.Println(graph)

	var error float64 = math.MaxUint32 - 1
	for error >= errorThres {
		listOfMessages := getMessages() 
			
	}
	



}

