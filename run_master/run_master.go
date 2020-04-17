package main

import (
	"os"
	"pregol"
)

func main() {
	args := os.Args[1:]
	port := args[0]
	primaryAddress := ""
	if len(args) > 1 {
		primaryAddress = args[1]
	}
	m := pregol.NewMaster(3, 1, "ip_add.txt", "../examples/data/unweighted/prob/rand20.json", port, primaryAddress)
	m.Run()
}
