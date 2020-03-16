package main

import (
	"log"
	"os/exec"
	"pregol"
)

// include this later for GUI
// "os/exec")

// function added by josh to run the server backend
func runGUI() {
	cmd := exec.Command("npm run dev")
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	m := pregol.NewMaster(1, 10, "ip_add.txt")
	m.InitConnections()
	m.AssignPartitions("example.json")
	m.DisseminateGraph()

	// runGUI()
}
