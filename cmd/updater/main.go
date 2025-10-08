package main

import "github.com/OryaHub/agent-os-instance/builders"

func main() {
	updater := builders.UpdaterBuilder()
	if err := updater.Start(); err != nil {
		panic(err)
	}
}
