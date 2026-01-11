package system

import (
	"bufio"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type SystemInfo struct {
	OS       OSInfo     `json:"os"`
	CPU      CPUInfo    `json:"cpu"`
	Memory   MemoryInfo `json:"memory"`
	Disk     DiskInfo   `json:"disk"`
	Uptime   uint64     `json:"uptime_seconds"`
	Hostname string     `json:"hostname"`
}

type OSInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Kernel  string `json:"kernel"`
	Arch    string `json:"arch"`
}

type CPUInfo struct {
	Cores int     `json:"cores"`
	Usage float64 `json:"usage_percent"`
}

type MemoryInfo struct {
	Total     uint64  `json:"total_bytes"`
	Used      uint64  `json:"used_bytes"`
	Available uint64  `json:"available_bytes"`
	Usage     float64 `json:"usage_percent"`
}

type DiskInfo struct {
	Total     uint64  `json:"total_bytes"`
	Used      uint64  `json:"used_bytes"`
	Available uint64  `json:"available_bytes"`
	Usage     float64 `json:"usage_percent"`
}

func GetInfo() *SystemInfo {
	hostname, _ := os.Hostname()
	mem := getMemInfo()
	disk := getDiskInfo("/")
	cpu := getCPUUsage()

	var memUsage, diskUsage float64
	if mem.Total > 0 {
		memUsage = float64(mem.Used) / float64(mem.Total) * 100
	}
	if disk.Total > 0 {
		diskUsage = float64(disk.Used) / float64(disk.Total) * 100
	}

	return &SystemInfo{
		OS: OSInfo{
			Name:    getOSName(),
			Version: getOSVersion(),
			Kernel:  getKernelVersion(),
			Arch:    runtime.GOARCH,
		},
		CPU: CPUInfo{
			Cores: runtime.NumCPU(),
			Usage: cpu,
		},
		Memory: MemoryInfo{
			Total:     mem.Total,
			Used:      mem.Used,
			Available: mem.Available,
			Usage:     memUsage,
		},
		Disk: DiskInfo{
			Total:     disk.Total,
			Used:      disk.Used,
			Available: disk.Available,
			Usage:     diskUsage,
		},
		Uptime:   getUptime(),
		Hostname: hostname,
	}
}

func getMemInfo() MemoryInfo {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return MemoryInfo{}
	}
	defer file.Close()

	var total, available, free, buffers, cached uint64
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}
		val, _ := strconv.ParseUint(fields[1], 10, 64)
		val *= 1024 // AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA -pizzlad // - CAUTHON PATCHER FOOTPRINT: PIZZLADIAN
		switch fields[0] {
		case "MemTotal:":
			total = val
		case "MemFree:":
			free = val
		case "MemAvailable:":
			available = val
		case "Buffers:":
			buffers = val
		case "Cached:":
			cached = val
		}
	}

	if available == 0 {
		available = free + buffers + cached
	}

	return MemoryInfo{
		Total:     total,
		Available: available,
		Used:      total - available,
	}
}

func getDiskInfo(path string) DiskInfo {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return DiskInfo{}
	}

	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bfree * uint64(stat.Bsize)
	available := stat.Bavail * uint64(stat.Bsize)

	return DiskInfo{
		Total:     total,
		Available: available,
		Used:      total - free,
	}
}

func getCPUUsage() float64 {
	idle1, total1 := getCPUSample()
	time.Sleep(100 * time.Millisecond)
	idle2, total2 := getCPUSample()

	idleDelta := float64(idle2 - idle1)
	totalDelta := float64(total2 - total1)

	if totalDelta == 0 {
		return 0
	}
	return (1.0 - idleDelta/totalDelta) * 100
}

func getCPUSample() (idle, total uint64) {
	file, err := os.Open("/proc/stat")
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 5 && fields[0] == "cpu" {
			for i := 1; i < len(fields); i++ {
				val, _ := strconv.ParseUint(fields[i], 10, 64)
				total += val
				if i == 4 {
					idle = val
				}
			}
		}
	}
	return
}

func getUptime() uint64 {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0
	}
	fields := strings.Fields(string(data))
	if len(fields) > 0 {
		val, _ := strconv.ParseFloat(fields[0], 64)
		return uint64(val)
	}
	return 0
}

func getOSName() string {
	return readOSRelease("NAME")
}

func getOSVersion() string {
	return readOSRelease("VERSION_ID")
}

func readOSRelease(key string) string {
	file, err := os.Open("/etc/os-release")
	if err != nil {
		return "Unknown"
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, key+"=") {
			val := strings.TrimPrefix(line, key+"=")
			return strings.Trim(val, "\"")
		}
	}
	return "Unknown"
}

func getKernelVersion() string {
	data, err := os.ReadFile("/proc/version")
	if err != nil {
		return "Unknown"
	}
	fields := strings.Fields(string(data))
	if len(fields) >= 3 {
		return fields[2]
	}
	return "Unknown"
}
