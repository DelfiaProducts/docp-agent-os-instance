package main

import "github.com/DelfiaProducts/docp-agent-os-instance/builders"

func main() {
	agent := builders.AgentBuilder()
	if agent == nil {
		panic("agent not found")
	}
	if err := agent.Start(); err != nil {
		panic(err)
	}
}
