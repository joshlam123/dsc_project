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
	graphName := "rand100.json"
	graph := fmt.Sprintf("../examples/data/weighted/prob/%s", graphName)

	// guiport := args[2]

	m := pregol.NewMaster(3, 1, "ip_add.txt", graph, port, primaryAddress)

	// go pregol.RunGUI(guiport, graph, "ip_add.txt", graphName)
	m.Run()
}
