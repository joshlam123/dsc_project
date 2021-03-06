package main

import (
	"dsc_project/src/pregol"
	"fmt"
	"os"
)

func main() {
	args := os.Args[1:]
	port := args[0]
	primaryAddress := ""
	if len(args) > 1 {
		primaryAddress = args[1]
	}
	graphName := "rand20.json"
	graph := fmt.Sprintf("../examples/data/unweighted/prob/%s", graphName)

	m := pregol.NewMaster(3, 1, "ip_add.txt", graph, port, primaryAddress)

	m.Run()
}
