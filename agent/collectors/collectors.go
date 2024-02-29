package collectors

import (
	"log"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
)

type Hardware struct {
	CPU       string `json:"cpu"`
	Memory    string `json:"memory"`
	Storage   string `json:"storage"`
	OS        string `json:"os"`
	OSVersion string `json:"os_version"`
	IPAddress string `json:"ip_address"`
}

type AgentData struct {
	Hostname      string   `json:"hostname"`
	IPAddress     string   `json:"ip_address"`
	OS            string   `json:"os"`
	OSVersion     string   `json:"os_version"`
	HardwareSpecs Hardware `json:"hardware_specs"`
	AgentVersion  string   `json:"agent_version"`
	LastSeen      string   `json:"last_seen"`
}

func getHardwareSpecs() (Hardware, error) {
	hardware := Hardware{}

	// Get OS info
	hardware.OS = runtime.GOOS
	log.Printf("OS: %s\n", hardware.OS)

	// Get OS version
	hostInfo, err := host.Info()
	if err != nil {
		return hardware, err
	}
	hardware.OSVersion = hostInfo.PlatformVersion
	log.Printf("OS Version: %s\n", hardware.OSVersion)

	// Get IP address
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return hardware, err
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				hardware.IPAddress = ipnet.IP.String()
				break
			}
		}
	}
	log.Printf("IP Address: %s\n", hardware.IPAddress)

	// Get CPU info
	cpuInfo, err := cpu.Info()
	if err != nil {
		return hardware, err
	}
	if len(cpuInfo) > 0 {
		hardware.CPU = strings.TrimSpace(cpuInfo[0].ModelName)
		log.Printf("CPU: %s", hardware.CPU)
	}

	// Get memory info
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return hardware, err
	}
	hardware.Memory = strconv.FormatUint(memInfo.Total/1024/1024, 10) + "MB"
	log.Printf("Memory: %s", hardware.Memory)

	// Get disk info (total storage)
	switch runtime.GOOS {
	case "windows":
		partitions, err := disk.Partitions(false)
		if err != nil {
			return hardware, err
		}
		for _, partition := range partitions {
			diskInfo, err := disk.Usage(partition.Mountpoint)
			if err != nil {
				log.Printf("could not get disk usage for %s: %v", partition.Device, err)
				continue
			}
			hardware.Storage += partition.Device + ": " + strconv.FormatUint(diskInfo.Total/1024/1024/1024, 10) + "GB; "
		}
		log.Printf("Storage: %s\n", hardware.Storage)

		return hardware, nil
	default:
		diskInfo, err := disk.Usage("/")
		if err != nil {
			return hardware, err
		}
		hardware.Storage = strconv.FormatUint(diskInfo.Total/1024/1024/1024, 10) + " GB"
		log.Printf("Storage: %s\n", hardware.Storage)

		return hardware, nil
	}
}

func CollectData() (AgentData, error) {
	hostname, _ := os.Hostname()

	// Get hardware specs
	hardware, err := getHardwareSpecs()
	if err != nil {
		return AgentData{}, err
	}

	return AgentData{
		Hostname:      hostname,
		IPAddress:     hardware.IPAddress,
		OS:            hardware.OS,
		OSVersion:     hardware.OSVersion,
		HardwareSpecs: hardware,
		AgentVersion:  "1.0.0",
		LastSeen:      time.Now().Format(time.RFC3339),
	}, nil
}
