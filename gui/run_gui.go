package main
import (
	"fmt"
	"../src/pregol"
)

func main() {

	// graphName := "rand20.json"
	// graph := fmt.Sprintf("../examples/data/unweighted/prob/%s", graphName)

	graphName := "checkpoint.json"
	graph := fmt.Sprintf("../run_master/%s", graphName)

	pregol.RunGUI("3000", graph, "../run_master/ip_add.txt", graphName)
}	