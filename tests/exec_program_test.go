package tests

import (
	"testing"

	"github.com/DelfiaProducts/docp-agent-os-instance/libs/bdd"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/pkg"
)

func TestNewExecProgram(t *testing.T) {
	bdd.Feature(t, "TestNewExecProgram", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Deve criar ExecProgram sem erro", func(s *bdd.Scenario) {
			var execProgram *pkg.ExecProgram
			s.When("NewExecProgram é chamado", func() {
				execProgram = pkg.NewExecProgram()
			})
			s.Then("ExecProgram não deve ser nil", func(t *testing.T) {
				bdd.AssertIsNotNil(t, execProgram, "ExecProgram deve ser diferente de nil")
			})
		})
	})
}

func TestExecProgramExecuteWithOutput(t *testing.T) {
	bdd.Feature(t, "TestExecProgramExecuteWithOutput", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Deve executar comando com saída sem erro", func(s *bdd.Scenario) {
			var execProgram *pkg.ExecProgram
			var output string
			var err error
			s.When("NewExecProgram é chamado", func() {
				execProgram = pkg.NewExecProgram()
			})
			s.Given("ExecProgram válido", func() {
				if execProgram == nil {
					t.Errorf("TestExecProgramExecuteWithOutput: expect(!nil) - got(%v)\n", execProgram)
				}
			})
			s.When("ExecuteWithOutput é chamado", func() {
				if execProgram != nil {
					output, err = execProgram.ExecuteWithOutput("uname", []string{"DOCP_DISTRO=ubuntu"}, "-o")
				}
			})
			s.Then("não deve retornar erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "ExecuteWithOutput não deve retornar erro")
			})
			s.Then("output deve ser diferente de vazio", func(t *testing.T) {
				if output == "" {
					t.Errorf("output esperado não deve ser vazio")
				}
			})
		})
	})
}

func TestExecProgramExecute(t *testing.T) {
	bdd.Feature(t, "TestExecProgramExecute", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Deve executar comando sem erro", func(s *bdd.Scenario) {
			var execProgram *pkg.ExecProgram
			var err error
			s.When("NewExecProgram é chamado", func() {
				execProgram = pkg.NewExecProgram()
			})
			s.Given("ExecProgram válido", func() {
				if execProgram == nil {
					t.Errorf("TestExecProgramExecute: expect(!nil) - got(%v)\n", execProgram)
				}
			})
			s.When("Execute é chamado", func() {
				if execProgram != nil {
					err = execProgram.Execute("ls", []string{"DOCP_DISTRO=ubuntu"}, "-la")
				}
			})
			s.Then("não deve retornar erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "Execute não deve retornar erro")
			})
		})
	})
}
