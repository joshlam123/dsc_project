package pregol

import (
	"encoding/json"
	"fmt"
	"os"
)

func writeToJson(jsonFile interface{}, name string) {
	// supposed to support any kind of type - to be tested
	jsonString, err := json.Marshal(jsonFile)
	fmt.Println(err)

	file, err := os.Create("./" + name + ".json")

	if err != nil {
		panic(err)
	}
	defer file.Close()
	file.Write(jsonString)
	file.Close()
	fmt.Println("JSON data written to ", file.Name())
}
