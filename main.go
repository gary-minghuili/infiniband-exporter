package main

import (
	cmd "infiniband/cmd"
	"os"
)

func main() {
	cmd := cmd.NewInfinibandExporterCommand()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
