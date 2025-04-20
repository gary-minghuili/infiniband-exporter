package main

import (
	"infiniband_exporter/cmd"
	"os"
)

func main() {
	command := cmd.NewInfinibandExporterCommand()
	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
