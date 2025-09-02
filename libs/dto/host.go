package dto

// DocpData is struct for data docp
type DocpData struct {
	ClientInfo  ClientInfo  `json:"client_info"`
	MetricsInfo MetricsInfo `json:"metrics_info"`
}

// ClientInfo is struct for information the client or user
type ClientInfo struct {
	Identifier       string `json:"identifier"`
	OrganizationId   string `json:"organization_id"`
	OrganizationName string `json:"organization_name"`
}

type ComputeInfo struct {
	Computename     string `json:"compute_name"`
	OperationSystem string `json:"operation_system"`
	Platform        string `json:"platform"`
	PlatformVersion string `json:"platform_version"`
	PlatformArch    string `json:"platform_arch"`
}

// MetricsInfo is struct for metrics the host
type MetricsInfo struct {
	HostInfo     HostInfo      `json:"host_info"`
	CPUInfo      []CPUInfo     `json:"cpus_info"`
	MemoryInfo   MemoryInfo    `json:"memory_info"`
	DiskInfo     DiskInfo      `json:"disk_info"`
	ProcessInfos []ProcessInfo `json:"process_infos"`
}

type HostInfo struct {
	Hostname        string `json:"hostname"`
	OperationSystem string `json:"operation_system"`
	Platform        string `json:"platform"`
	PlatformVersion string `json:"platform_version"`
	PlatformArch    string `json:"platform_arch"`
}
type CPUInfo struct {
	CPU       int32  `json:"cpu"`
	CPUCores  int32  `json:"cpu_cores"`
	CPUFamily string `json:"cpu_family"`
	CPUModel  string `json:"cpu_model"`
}
type MemoryInfo struct {
	MemoryUsed      uint64 `json:"memory_used"`
	MemoryTotal     uint64 `json:"memory_total"`
	MemoryAvailable uint64 `json:"memory_available"`
}
type DiskInfo struct {
	DiskTotal uint64 `json:"disk_total"`
	DistFree  uint64 `json:"disk_free"`
	DistUsed  uint64 `json:"disk_used"`
}
type ProcessInfo struct {
	Pid        int32  `json:"pid"`
	Name       string `json:"name"`
	NumFDs     int32  `json:"num_file_descriptors"`
	NumThreads int32  `json:"num_threads"`
	ExecPath   string `json:"exec_path"`
	Background bool   `json:"background"`
}
