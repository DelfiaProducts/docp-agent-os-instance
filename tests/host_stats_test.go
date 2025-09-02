package tests

import (
	"testing"

	"github.com/DelfiaProducts/docp-agent-os-instance/libs/bdd"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/pkg"
)

func TestNewHostStats(t *testing.T) {
	bdd.Feature(t, "TestNewHostStats", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Deve criar HostStats sem erro", func(s *bdd.Scenario) {
			var hostStats *pkg.HostStats
			s.When("NewHostStats é chamado", func() {
				hostStats = pkg.NewHostStats()
			})
			s.Then("HostStats não deve ser nil", func(t *testing.T) {
				if hostStats == nil {
					t.Errorf("TestNewHostStats: expect(!nil) - got(%v)\n", hostStats)
				}
			})
		})
	})
}

func TestHostStatsHostInfo(t *testing.T) {
	bdd.Feature(t, "TestHostStatsHostInfo", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Deve obter informações do host sem erro", func(s *bdd.Scenario) {
			var hostStats *pkg.HostStats
			var hostInfo interface{}
			var err error
			s.When("NewHostStats é chamado", func() {
				hostStats = pkg.NewHostStats()
			})
			s.Given("HostStats válido", func() {
				if hostStats == nil {
					t.Errorf("TestHostStatsHostInfo: expect(!nil) - got(%v)\n", hostStats)
				}
			})
			s.When("ComputeInfo é chamado", func() {
				if hostStats != nil {
					hostInfo, err = hostStats.ComputeInfo()
				}
			})
			s.Then("não deve retornar erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "ComputeInfo não deve retornar erro")
			})
			s.Then("hostInfo deve ser diferente de nil", func(t *testing.T) {
				if hostInfo == nil {
					t.Errorf("TestHostStatsHostInfo: expect(!nil) - got(%v)\n", hostInfo)
				}
			})
		})
	})
}

func TestHostStatsCPUInfo(t *testing.T) {
	bdd.Feature(t, "TestHostStatsCPUInfo", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Deve obter informações da CPU sem erro", func(s *bdd.Scenario) {
			var hostStats *pkg.HostStats
			var cpuInfo interface{}
			var err error
			s.When("NewHostStats é chamado", func() {
				hostStats = pkg.NewHostStats()
			})
			s.Given("HostStats válido", func() {
				if hostStats == nil {
					t.Errorf("TestHostStatsCPUInfo: expect(!nil) - got(%v)\n", hostStats)
				}
			})
			s.When("CPUInfo é chamado", func() {
				if hostStats != nil {
					cpuInfo, err = hostStats.CPUInfo()
				}
			})
			s.Then("não deve retornar erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "CPUInfo não deve retornar erro")
			})
			s.Then("cpuInfo deve ser diferente de nil", func(t *testing.T) {
				if cpuInfo == nil {
					t.Errorf("TestHostStatsCPUInfo: expect(!nil) - got(%v)\n", cpuInfo)
				}
			})
		})
	})
}

func TestHostStatsMemoryInfo(t *testing.T) {
	bdd.Feature(t, "TestHostStatsMemoryInfo", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Deve obter informações da memória sem erro", func(s *bdd.Scenario) {
			var hostStats *pkg.HostStats
			var memoryInfo interface{}
			var err error
			s.When("NewHostStats é chamado", func() {
				hostStats = pkg.NewHostStats()
			})
			s.Given("HostStats válido", func() {
				if hostStats == nil {
					t.Errorf("TestHostStatsMemoryInfo: expect(!nil) - got(%v)\n", hostStats)
				}
			})
			s.When("MemoryInfo é chamado", func() {
				if hostStats != nil {
					memoryInfo, err = hostStats.MemoryInfo()
				}
			})
			s.Then("não deve retornar erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "MemoryInfo não deve retornar erro")
			})
			s.Then("memoryInfo deve ser diferente de nil", func(t *testing.T) {
				if memoryInfo == nil {
					t.Errorf("TestHostStatsMemoryInfo: expect(!nil) - got(%v)\n", memoryInfo)
				}
			})
		})
	})
}

func TestHostStatsDiskInfo(t *testing.T) {
	bdd.Feature(t, "TestHostStatsDiskInfo", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Deve obter informações do disco sem erro", func(s *bdd.Scenario) {
			var hostStats *pkg.HostStats
			var diskInfo interface{}
			var err error
			s.When("NewHostStats é chamado", func() {
				hostStats = pkg.NewHostStats()
			})
			s.Given("HostStats válido", func() {
				if hostStats == nil {
					t.Errorf("TestHostStatsDiskInfo: expect(!nil) - got(%v)\n", hostStats)
				}
			})
			s.When("DiskInfo é chamado", func() {
				if hostStats != nil {
					diskInfo, err = hostStats.DiskInfo()
				}
			})
			s.Then("não deve retornar erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "DiskInfo não deve retornar erro")
			})
			s.Then("diskInfo deve ser diferente de nil", func(t *testing.T) {
				if diskInfo == nil {
					t.Errorf("TestHostStatsDiskInfo: expect(!nil) - got(%v)\n", diskInfo)
				}
			})
		})
	})
}

func TestHostStatsProcessInfo(t *testing.T) {
	bdd.Feature(t, "TestHostStatsProcessInfo", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Deve obter informações dos processos sem erro", func(s *bdd.Scenario) {
			var hostStats *pkg.HostStats
			var processInfo interface{}
			var err error
			s.When("NewHostStats é chamado", func() {
				hostStats = pkg.NewHostStats()
			})
			s.Given("HostStats válido", func() {
				if hostStats == nil {
					t.Errorf("TestHostStatsProcessInfo: expect(!nil) - got(%v)\n", hostStats)
				}
			})
			s.When("ProcessInfo é chamado", func() {
				if hostStats != nil {
					processInfo, err = hostStats.ProcessInfo()
				}
			})
			s.Then("não deve retornar erro", func(t *testing.T) {
				bdd.AssertNoError(t, err, "ProcessInfo não deve retornar erro")
			})
			s.Then("processInfo deve ser diferente de nil", func(t *testing.T) {
				if processInfo == nil {
					t.Errorf("TestHostStatsProcessInfo: expect(!nil) - got(%v)\n", processInfo)
				}
			})
		})
	})
}
