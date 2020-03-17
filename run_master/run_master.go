package main

import (
	"pregol"
)

func main() {
	m := pregol.NewMaster(1, 10, "ip_add.txt", "example.json")
	m.Run()
}
