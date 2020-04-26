package main
import (
	"fmt"
	"../src/pregol"
)

func main() {

	// graphName := "rand20.json"
	// graph := fmt.Sprintf("../examples/data/unweighted/prob/%s", graphName)
	// guiport := args[2]

	graphName := "rand20.json"
	graph := fmt.Sprintf("../examples/data/weighted/prob/%s", graphName)

	pregol.RunGUI("9000", graph, "../run_master/ip_add.txt", graphName)
	// go pregol.RunGUI(guiport, graph, "ip_add.txt", graphName)
}	