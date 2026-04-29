package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var procRoot = "/proc"
var sysRoot = "/sys"

func setProcRoot(root string) {
	if strings.TrimSpace(root) != "" {
		procRoot = root
	}
}

func setSysRoot(root string) {
	if strings.TrimSpace(root) != "" {
		sysRoot = root
	}
}

func readProcFile(name string) ([]byte, error) {
	return os.ReadFile(filepath.Join(procRoot, name))
}

func readFirstExisting(paths ...string) ([]byte, error) {
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err == nil {
			return b, nil
		}
	}
	return nil, os.ErrNotExist
}

func parseTempValue(raw string) (float64, error) {
	val, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil {
		return 0, err
	}
	if val > 1000 {
		val = val / 1000
	}
	return val, nil
}

func getLocalIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "unknown"
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok || ipNet.IP == nil || ipNet.IP.IsLoopback() {
				continue
			}
			if ip4 := ipNet.IP.To4(); ip4 != nil {
				return ip4.String()
			}
		}
	}
	return "unknown"
}

func getExternalIP() string {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://api.ipify.org")
	if err != nil {
		return "no internet"
	}
	defer resp.Body.Close()
	ip, err := io.ReadAll(resp.Body)
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(ip))
}

func getSSID() string {
	out, err := exec.Command("iwgetid", "-r").Output()
	if err != nil {
		return "not connected"
	}
	ssid := strings.TrimSpace(string(out))
	if ssid == "" {
		return "not connected"
	}
	return ssid
}

func getUptime() string {
	b, readErr := readProcFile("uptime")
	if readErr != nil {
		return "unknown"
	}
	fields := strings.Fields(string(b))
	if len(fields) == 0 {
		return "unknown"
	}
	seconds, parseErr := strconv.ParseFloat(fields[0], 64)
	if parseErr != nil {
		return "unknown"
	}
	d := time.Duration(seconds) * time.Second
	days := d / (24 * time.Hour)
	d -= days * 24 * time.Hour
	hours := d / time.Hour
	d -= hours * time.Hour
	mins := d / time.Minute
	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, mins)
	}
	return fmt.Sprintf("%dh %dm", hours, mins)
}

func getMemory() string {
	b, readErr := readProcFile("meminfo")
	if readErr != nil {
		return "unknown"
	}
	var totalKB, availKB int64
	for _, line := range strings.Split(string(b), "\n") {
		if strings.HasPrefix(line, "MemTotal:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				totalKB, _ = strconv.ParseInt(parts[1], 10, 64)
			}
		}
		if strings.HasPrefix(line, "MemAvailable:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				availKB, _ = strconv.ParseInt(parts[1], 10, 64)
			}
		}
	}
	if totalKB == 0 {
		return "unknown"
	}
	usedKB := totalKB - availKB
	totalMB := totalKB / 1024
	usedMB := usedKB / 1024
	freeMB := availKB / 1024
	usedPercent := float64(usedKB) / float64(totalKB) * 100

	return fmt.Sprintf(
		"Total: %d MB\nUsed: %d MB (%.1f%%)\nFree: %d MB",
		totalMB,
		usedMB,
		usedPercent,
		freeMB,
	)
}

func cpuCores() int {
	b, err := readProcFile("cpuinfo")
	if err != nil {
		return 0
	}
	count := 0
	for _, line := range strings.Split(string(b), "\n") {
		if strings.HasPrefix(line, "processor") {
			count++
		}
	}
	if count == 0 {
		return runtime.NumCPU()
	}
	return count
}

type cpuTimes struct {
	idle  int64
	total int64
}

func readCPUTimes() (cpuTimes, error) {
	b, err := readProcFile("stat")
	if err != nil {
		return cpuTimes{}, err
	}
	lines := strings.Split(string(b), "\n")
	for _, line := range lines {
		if !strings.HasPrefix(line, "cpu ") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 8 {
			return cpuTimes{}, errors.New("invalid /proc/stat format")
		}
		var vals []int64
		for _, f := range fields[1:] {
			v, parseErr := strconv.ParseInt(f, 10, 64)
			if parseErr != nil {
				return cpuTimes{}, parseErr
			}
			vals = append(vals, v)
		}
		var total int64
		for _, v := range vals {
			total += v
		}
		idle := vals[3]
		if len(vals) > 4 {
			idle += vals[4] // iowait is treated as idle time
		}
		return cpuTimes{idle: idle, total: total}, nil
	}
	return cpuTimes{}, errors.New("cpu line not found")
}

func cpuUsagePercent() string {
	first, err := readCPUTimes()
	if err != nil {
		return "unknown"
	}
	time.Sleep(250 * time.Millisecond)
	second, err := readCPUTimes()
	if err != nil {
		return "unknown"
	}
	totalDelta := second.total - first.total
	idleDelta := second.idle - first.idle
	if totalDelta <= 0 {
		return "unknown"
	}
	usage := float64(totalDelta-idleDelta) / float64(totalDelta) * 100
	if usage < 0 {
		usage = 0
	}
	if usage > 100 {
		usage = 100
	}
	return fmt.Sprintf("%.1f%%", usage)
}

func cpuTemperature() string {
	zonePaths, _ := filepath.Glob(filepath.Join(sysRoot, "class/thermal/thermal_zone*/temp"))
	if len(zonePaths) > 0 {
		for _, p := range zonePaths {
			b, err := os.ReadFile(p)
			if err != nil {
				continue
			}
			if val, parseErr := parseTempValue(string(b)); parseErr == nil {
				return fmt.Sprintf("%.1f°C", val)
			}
		}
	}

	hwmonPaths, _ := filepath.Glob(filepath.Join(sysRoot, "class/hwmon/hwmon*/temp*_input"))
	if len(hwmonPaths) > 0 {
		for _, p := range hwmonPaths {
			b, err := os.ReadFile(p)
			if err != nil {
				continue
			}
			if val, parseErr := parseTempValue(string(b)); parseErr == nil {
				return fmt.Sprintf("%.1f°C", val)
			}
		}
	}

	b, err := readFirstExisting(
		filepath.Join(sysRoot, "class/thermal/thermal_zone0/temp"),
		filepath.Join(sysRoot, "devices/virtual/thermal/thermal_zone0/temp"),
		filepath.Join(sysRoot, "class/hwmon/hwmon0/temp1_input"),
	)
	if err != nil {
		return "unknown"
	}
	val, parseErr := parseTempValue(string(b))
	if parseErr != nil {
		return "unknown"
	}
	return fmt.Sprintf("%.1f°C", val)
}

func runReboot(cmdline string) error {
	parts := strings.Fields(cmdline)
	if len(parts) == 0 {
		return errors.New("empty reboot command")
	}
	cmd := exec.Command(parts[0], parts[1:]...)
	return cmd.Run()
}
