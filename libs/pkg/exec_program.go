package pkg

import (
	"os"
	"os/exec"
)

// ExecProgram is struct for exec program
type ExecProgram struct{}

// NewExecProgram return instance of exec program
func NewExecProgram() *ExecProgram {
	return &ExecProgram{}
}

// prepareEnvironments execute prepare environments for command
func (e *ExecProgram) prepareEnvironments(cmd *exec.Cmd, envs []string) error {
	cmd.Env = os.Environ()
	for _, env := range envs {
		cmd.Env = append(cmd.Env, env)
	}
	return nil
}

// ExecuteWithOutput running program in process terminal
// and return output
func (e *ExecProgram) ExecuteWithOutput(command string, environments []string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	if err := e.prepareEnvironments(cmd, environments); err != nil {
		return "", err
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// Execute running program in process terminal
func (e *ExecProgram) Execute(command string, environments []string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := e.prepareEnvironments(cmd, environments); err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
