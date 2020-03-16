package main

import (
	"log"
	"pregol"
)



func main() {
	m := pregol.NewMaster(1, 10, "ip_add.txt")
	m.InitConnections()
	m.AssignPartitions("example.json")
	m.DisseminateGraph()

	// runGUI()
}
