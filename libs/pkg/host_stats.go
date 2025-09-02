package pkg

import (
	"strings"

	"github.com/DelfiaProducts/docp-agent-os-instance/libs/dto"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/process"
)

// HostStats is struct for host stats
type HostStats struct{}

// NewHostStats return instance of host stats
func NewHostStats() *HostStats {
	return &HostStats{}
}

// ComputeInfo return info from host
func (h *HostStats) ComputeInfo() (dto.ComputeInfo, error) {
	hostInfo, err := host.Info()
	if err != nil {
		return dto.ComputeInfo{}, err
	}
	return dto.ComputeInfo{
		Computename:     hostInfo.Hostname,
		OperationSystem: hostInfo.OS,
		Platform:        hostInfo.Platform,
		PlatformVersion: hostInfo.PlatformVersion,
		PlatformArch:    hostInfo.KernelArch,
	}, nil
}

// CPUInfo return slice of cpu info from host
func (h *HostStats) CPUInfo() ([]dto.CPUInfo, error) {
	cpusInfo, err := cpu.Info()
	if err != nil {
		return []dto.CPUInfo{}, err
	}
	var cpus []dto.CPUInfo
	for _, cpu := range cpusInfo {
		cpInfo := dto.CPUInfo{
			CPU:       cpu.CPU,
			CPUCores:  cpu.Cores,
			CPUFamily: cpu.Family,
			CPUModel:  cpu.Model,
		}
		cpus = append(cpus, cpInfo)
	}
	return cpus, nil
}

// ProcessInfo return process info from host
func (h *HostStats) ProcessInfo() ([]dto.ProcessInfo, error) {
	procs, err := process.Processes()
	if err != nil {
		return []dto.ProcessInfo{}, err
	}
	var allProcess []dto.ProcessInfo
	for _, prcs := range procs {
		name, _ := prcs.Name()
		numFds, _ := prcs.NumFDs()
		numThreads, _ := prcs.NumThreads()
		execPath, _ := prcs.Exe()
		background, _ := prcs.Background()

		prs := dto.ProcessInfo{
			Pid:        prcs.Pid,
			Name:       name,
			NumFDs:     numFds,
			NumThreads: numThreads,
			ExecPath:   execPath,
			Background: background,
		}
		allProcess = append(allProcess, prs)
	}
	return allProcess, nil
}

// MemoryInfo return memory info from host
func (h *HostStats) MemoryInfo() (dto.MemoryInfo, error) {
	memoryInfo, err := mem.VirtualMemory()
	if err != nil {
		return dto.MemoryInfo{}, err
	}
	return dto.MemoryInfo{
		MemoryTotal:     memoryInfo.Total,
		MemoryUsed:      memoryInfo.Used,
		MemoryAvailable: memoryInfo.Available,
	}, nil
}

// DiskInfo return disk info from host
func (h *HostStats) DiskInfo() (dto.DiskInfo, error) {
	diskInfo, err := disk.Usage("/")
	if err != nil {
		return dto.DiskInfo{}, err
	}
	return dto.DiskInfo{
		DiskTotal: diskInfo.Total,
		DistFree:  diskInfo.Free,
		DistUsed:  diskInfo.Used,
	}, nil
}

// AptOrDpkgIsRunning return if dpkg or apt is running
func (h *HostStats) AptOrDpkgIsRunning() (bool, error) {
	processes, err := process.Processes()
	if err != nil {
		return false, err
	}
	aptOrDpkgIsRun := false
	for _, ps := range processes {
		name, err := ps.Name()
		if err == nil {
			if strings.Contains(name, "apt-get") || strings.Contains(name, "dpkg") {
				aptOrDpkgIsRun = true
			}
		}
	}
	return aptOrDpkgIsRun, nil
}
