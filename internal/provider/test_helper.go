package provider

import (
	"os"
	"testing"
)

// helperTestAccTFExampleConfig is intended for us to more easily pass in just the
// corresponding resource folder and file
// Example: helperTestAccTFExampleConfig("resources/waypoint_project/project.tf")
// Example: helperTestAccTFExampleConfig("data-sources/waypoint_project/project.tf")
func helperTestAccTFExampleConfig(t *testing.T, filename string) string {
	t.Helper()
	fileContent, err := os.ReadFile("../../examples/" + filename)
	if err != nil {
		t.Fatalf("err reading file: %s, %s", filename, err)
	}

	return string(fileContent)
}
