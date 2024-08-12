package collectors

import (
	"log"
	"net"
	"os"
	"os/user"
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
	ID            int32    `json:"id"`
	Hostname      string   `json:"hostname"`
	IPAddress     string   `json:"ip_address"`
	OS            string   `json:"os"`
	OSVersion     string   `json:"os_version"`
	HardwareSpecs Hardware `json:"hardware_specs"`
	AgentVersion  string   `json:"agent_version"`
	LastSeen      string   `json:"last_seen"`
	LastUser      string   `json:"last_user"`
	Token         string   `json:"token"`
}

func getHardwareSpecs() (Hardware, error) {
	hardware := Hardware{}

	// Get OS info
	hardware.OS = runtime.GOOS

	// Get OS version
	hostInfo, err := host.Info()
	if err != nil {
		return hardware, err
	}
	hardware.OS = strings.Replace(hostInfo.Platform, "Microsoft", "", 1)
	hardware.OSVersion = hostInfo.PlatformVersion

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

	// Get CPU info
	cpuInfo, err := cpu.Info()
	if err != nil {
		return hardware, err
	}
	if len(cpuInfo) > 0 {
		hardware.CPU = strings.TrimSpace(cpuInfo[0].ModelName)
	}

	// Get memory info
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return hardware, err
	}
	hardware.Memory = strconv.FormatUint(memInfo.Total/1024/1024, 10) + "MB"

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

		return hardware, nil
	default:
		diskInfo, err := disk.Usage("/")
		if err != nil {
			return hardware, err
		}
		hardware.Storage = strconv.FormatUint(diskInfo.Total/1024/1024/1024, 10) + " GB"

		return hardware, nil
	}
}

func getCurrentUser() (string, error) {
	user, err := user.Current()
	if err != nil {
		return "", err
	}

	return user.Username, nil
}

func CollectData() (AgentData, error) {
	hostname, _ := os.Hostname()

	// Get hardware specs
	hardware, err := getHardwareSpecs()
	if err != nil {
		return AgentData{}, err
	}

	// Get current user
	user, err := getCurrentUser()
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
		LastUser:      user,
		LastSeen:      time.Now().Format(time.RFC3339),
	}, nil
}
