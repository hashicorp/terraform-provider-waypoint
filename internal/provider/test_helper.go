package provider

import (
	"log"
	"os"
)

// helperTestAccTFExampleConfig is intended for us to more easily pass in just the
// corresponding resource folder and file
// Example: helperTestAccTFExampleConfig("resources/waypoint_project/project.tf")
// Example: helperTestAccTFExampleConfig("data-sources/waypoint_project/project.tf")
func helperTestAccTFExampleConfig(filename string) (string, error) {
	fileContent, err := os.ReadFile("../../examples/" + filename)
	if err != nil {
		log.Fatalf("err reading file: %s, %s", filename, err)
		return "", err
	}

	return string(fileContent), nil
}
