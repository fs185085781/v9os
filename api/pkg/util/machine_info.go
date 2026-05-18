package util

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

func CpuInfo() string {
	cpuInfos, err := cpu.Info()
	if err != nil {
		return "Unknown/0/0"
	}
	var cpuName string
	if len(cpuInfos) > 0 {
		cpuName = strings.TrimSpace(cpuInfos[0].ModelName)
	} else {
		cpuName = "Unknown"
	}
	physCores, err1 := cpu.Counts(false)
	logicCores, err2 := cpu.Counts(true)
	if err1 != nil || err2 != nil {
		physCores = runtime.NumCPU()
		logicCores = physCores
	}
	return fmt.Sprintf("%s/物理%d核/逻辑%d核", cpuName, physCores, logicCores)
}
func MemoryInfo() string {
	vmem, err := mem.VirtualMemory()
	if err != nil {
		return "0/0GB"
	}
	usedGB := float64(vmem.Used) / 1024 / 1024 / 1024
	totalGB := float64(vmem.Total) / 1024 / 1024 / 1024
	return fmt.Sprintf("%.1f/%.1fGB", usedGB, totalGB)
}
func DiskInfo() []string {
	if runtime.GOOS == "windows" {
		parts, _ := disk.Partitions(false)
		var disks []string
		for _, part := range parts {
			if len(part.Mountpoint) == 2 && part.Mountpoint[1] == ':' {
				usage, err := disk.Usage(part.Mountpoint)
				if err == nil && usage.Total > 100*1024*1024 {
					usedGB := float64(usage.Used) / 1024 / 1024 / 1024
					totalGB := float64(usage.Total) / 1024 / 1024 / 1024

					if totalGB >= 1024 {
						disks = append(disks,
							fmt.Sprintf("%s %.1fG/%.1fT",
								part.Mountpoint, usedGB, totalGB/1024))
					} else {
						disks = append(disks,
							fmt.Sprintf("%s %.1fG/%.1fG",
								part.Mountpoint, usedGB, totalGB))
					}
				}
			}
		}
		if len(disks) > 0 {
			return disks
		}
		return []string{"C: 0G/0G"}
	}
	usage, _ := disk.Usage("/")
	if usage != nil {
		usedGB := float64(usage.Used) / 1024 / 1024 / 1024
		totalGB := float64(usage.Total) / 1024 / 1024 / 1024

		if totalGB >= 1024 {
			return []string{fmt.Sprintf("/ %.1fG/%.1fT", usedGB, totalGB/1024)}
		}
		return []string{fmt.Sprintf("/ %.1fG/%.1fG", usedGB, totalGB)}
	}
	return []string{"/ 0G/0G"}
}

var threshold = 0.9 // 默认90%为满载
func LoadInfo() int {
	cpuPercent := getCPUPercent()
	memPercent := getMemoryPercent()
	cpuRatio := cpuPercent / 100 / threshold
	memRatio := memPercent / 100 / threshold
	if cpuRatio > 1.0 {
		cpuRatio = 1.0
	}
	if memRatio > 1.0 {
		memRatio = 1.0
	}
	load := cpuRatio * memRatio * 100
	return int(load)
}
func getCPUPercent() float64 {
	percent, err := cpu.Percent(0, false)
	if err != nil || len(percent) == 0 {
		return 0
	}
	return percent[0]
}
func getMemoryPercent() float64 {
	vmem, err := mem.VirtualMemory()
	if err != nil {
		return 0
	}
	return vmem.UsedPercent
}
