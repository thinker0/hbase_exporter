package collector

type hbaseSystemResponse struct {
	Name                       string  `json:"name"`
	ModelerType                string  `json:"modelerType"`
	OpenFileDescriptorCount    int64   `json:"OpenFileDescriptorCount"`
	MaxFileDescriptorCount     int64   `json:"MaxFileDescriptorCount"`
	CommittedVirtualMemorySize int64   `json:"CommittedVirtualMemorySize"`
	TotalSwapSpaceSize         int64   `json:"TotalSwapSpaceSize"`
	FreeSwapSpaceSize          int64   `json:"FreeSwapSpaceSize"`
	ProcessCpuTime             int64   `json:"ProcessCpuTime"`
	TotalPhysicalMemorySize    int64   `json:"TotalPhysicalMemorySize"`
	SystemCpuLoad              float64 `json:"SystemCpuLoad"`
	ProcessCpuLoad             float64 `json:"ProcessCpuLoad"`
	FreePhysicalMemorySize     int64   `json:"FreePhysicalMemorySize"`
	AvailableProcessors        int64   `json:"AvailableProcessors"`
	SystemLoadAverage          float64 `json:"SystemLoadAverage"`
}
