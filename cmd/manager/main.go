package main

import "github.com/DelfiaProducts/docp-agent-os-instance/builders"

func main() {
	manager := builders.ManagerBuilder()
	if err := manager.Start(); err != nil {
		panic(err)
	}
}
