package tests

import (
	"testing"

	"github.com/DelfiaProducts/docp-agent-os-instance/libs/bdd"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/pkg"
)

func TestNewFileSystem(t *testing.T) {
	bdd.Feature(t, "FileSystem", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("criar novo FileSystem", func(s *bdd.Scenario) {
			var fileSystem *pkg.FileSystem
			s.When("instancio o FileSystem", func() {
				fileSystem = pkg.NewFileSystem()
			})
			s.Then("deve ser diferente de nil", func(t *testing.T) {
				bdd.AssertIsNotNil(t, fileSystem, "FileSystem deve ser diferente de nil")
			})
		})
	})
}

func TestFileSystemVerifyDirExist(t *testing.T) {
	bdd.Feature(t, "FileSystem", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("verificar existência de diretório", func(s *bdd.Scenario) {
			var fileSystem *pkg.FileSystem
			var err error
			s.Given("um FileSystem válido", func() {
				fileSystem = pkg.NewFileSystem()
			})
			s.When("verifico se o diretório existe", func() {
				_, err = fileSystem.VerifyDirExist("/etc/datadog-agent/")
			})
			s.Then("não deve retornar erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "VerifyDirExist não deve retornar erro")
			})
		})
	})

}

func TestFileSystemGetFileContent(t *testing.T) {
	bdd.Feature(t, "FileSystem", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("obter conteúdo de arquivo", func(s *bdd.Scenario) {
			var fileSystem *pkg.FileSystem
			var content []byte
			var err error
			s.Given("um FileSystem válido", func() {
				fileSystem = pkg.NewFileSystem()
			})
			s.When("leio o conteúdo do arquivo", func() {
				content, err = fileSystem.GetFileContent("./config.json")
			})
			s.Then("não deve retornar erro e conteúdo não pode ser nil", func(t *testing.T) {
				bdd.AssertNoError(t, err, "GetFileContent não deve retornar erro")
				bdd.AssertIsNotNil(t, content, "Conteúdo não pode ser nil")
			})
		})
	})
}

func TestFileSystemWriteFileContent(t *testing.T) {
	bdd.Feature(t, "FileSystem", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("escrever conteúdo em arquivo", func(s *bdd.Scenario) {
			var fileSystem *pkg.FileSystem
			var err error
			s.Given("um FileSystem válido", func() {
				fileSystem = pkg.NewFileSystem()
			})
			s.When("escrevo conteúdo no arquivo", func() {
				err = fileSystem.WriteFileContent("/opt/docp-agent/state/hash", []byte("xpto-123"))
			})
			s.Then("não deve retornar erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "WriteFileContent não deve retornar erro")
			})
		})
	})
}

func TestFileSystemVerifyDirExistAndCreate(t *testing.T) {
	bdd.Feature(t, "FileSystem", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("verificar existência e criar diretório", func(s *bdd.Scenario) {
			var fileSystem *pkg.FileSystem
			var err error
			s.Given("um FileSystem válido", func() {
				fileSystem = pkg.NewFileSystem()
			})
			s.When("verifico/crio diretório", func() {
				err = fileSystem.VerifyDirExistAndCreate("/opt/docp-agent/state")
			})
			s.Then("não deve retornar erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "VerifyDirExistAndCreate não deve retornar erro")
			})
		})
	})
}

func TestFileSystemCreateFile(t *testing.T) {
	bdd.Feature(t, "FileSystem", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("criar arquivo", func(s *bdd.Scenario) {
			var fileSystem *pkg.FileSystem
			var err error
			s.Given("um FileSystem válido", func() {
				fileSystem = pkg.NewFileSystem()
			})
			s.When("crio um arquivo", func() {
				err = fileSystem.CreateFile("/opt/docp-agent/state/hash")
			})
			s.Then("não deve retornar erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "CreateFile não deve retornar erro")
			})
		})
	})
}

func TestFileSystemCreateOrUpdateSymlink(t *testing.T) {
	bdd.Feature(t, "FileSystem", func(t *testing.T, scenario func(description string, steps func(s *bdd.Scenario))) {
		scenario("criar ou atualizar symlink", func(s *bdd.Scenario) {
			var fileSystem *pkg.FileSystem
			var err error
			s.Given("um FileSystem válido", func() {
				fileSystem = pkg.NewFileSystem()
			})
			s.When("crio um symlink", func() {
				err = fileSystem.CreateOrUpdateSymlink("/opt/docp-agent/bin/releases/0.1.1/manager", "/opt/docp-agent/bin/current/manager")
			})
			s.Then("não deve retornar erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "CreateOrUpdateSymlink não deve retornar erro")
			})
		})
	})
}
