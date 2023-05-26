package provider

import (
	"os"
	"testing"
)

func TestAccConfigSource(*testing.T) {
	host := os.Getenv("WAYPOINT_HOST")
	token := os.Getenv("WAYPOINT_TOKEN")

}
