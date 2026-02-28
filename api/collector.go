package api

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Stats struct {
	Timestamp  int64      `json:"timestamp"`
	Hostname   string     `json:"hostname"`
	Uptime     string     `json:"uptime"`
	CPUPercent []float64  `json:"cpu_percent"`
	CPUTotal   float64    `json:"cpu_total"`
	CPUMHz     []float64  `json:"cpu_mhz"`
	RAMTotal   uint64     `json:"ram_total"`
	RAMUsed    uint64     `json:"ram_used"`
	RAMPercent float64    `json:"ram_percent"`
	SwapTotal  uint64     `json:"swap_total"`
	SwapUsed   uint64     `json:"swap_used"`
	Disks      []DiskStat `json:"disks"`
	Network    NetStat    `json:"network"`
	Temps      []TempStat `json:"temps"`
}

type DiskStat struct {
	Mount   string  `json:"mount"`
	Device  string  `json:"device"`
	Total   uint64  `json:"total"`
	Used    uint64  `json:"used"`
	Free    uint64  `json:"free"`
	Percent float64 `json:"percent"`
}

type NetStat struct {
	RxBytes uint64  `json:"rx_bytes"`
	TxBytes uint64  `json:"tx_bytes"`
	RxRate  float64 `json:"rx_rate"`
	TxRate  float64 `json:"tx_rate"`
}

type TempStat struct {
	Name string  `json:"name"`
	Temp float64 `json:"temp"`
}

type rawCPUStat struct {
	user, nice, system, idle, iowait, irq, softirq, steal uint64
}

type rawNetStat struct {
	rx, tx uint64
}

type Collector struct {
	prevCPU     []rawCPUStat
	prevNet     rawNetStat
	prevNetTime time.Time
}

func NewCollector() *Collector {
	return &Collector{}
}

func (c *Collector) Collect() *Stats {
	s := &Stats{Timestamp: time.Now().Unix()}

	if h, err := os.Hostname(); err == nil {
		s.Hostname = h
	}

	s.Uptime = readUptime()

	cpuStats := readCPUStat()
	if len(c.prevCPU) > 0 && len(cpuStats) == len(c.prevCPU) {
		for i := range cpuStats {
			pct := calcCPUPercent(c.prevCPU[i], cpuStats[i])
			if i == 0 {
				s.CPUTotal = pct
			} else {
				s.CPUPercent = append(s.CPUPercent, pct)
			}
		}
	} else {
		if len(cpuStats) > 1 {
			s.CPUTotal = 0
			for range cpuStats[1:] {
				s.CPUPercent = append(s.CPUPercent, 0)
			}
		}
	}
	c.prevCPU = cpuStats

	s.CPUMHz = readCPUMHz()

	for len(s.CPUMHz) < len(s.CPUPercent) {
		s.CPUMHz = append(s.CPUMHz, 0)
	}
	for len(s.CPUPercent) < len(s.CPUMHz) {
		s.CPUPercent = append(s.CPUPercent, 0)
	}

	s.RAMTotal, s.RAMUsed, s.SwapTotal, s.SwapUsed = readMemInfo()
	if s.RAMTotal > 0 {
		s.RAMPercent = float64(s.RAMUsed) / float64(s.RAMTotal) * 100
	}

	s.Disks = readDisks()

	rxBytes, txBytes := readNetDev()
	now := time.Now()
	if !c.prevNetTime.IsZero() {
		elapsed := now.Sub(c.prevNetTime).Seconds()
		if elapsed > 0 && rxBytes >= c.prevNet.rx && txBytes >= c.prevNet.tx {
			s.Network.RxRate = float64(rxBytes-c.prevNet.rx) / elapsed
			s.Network.TxRate = float64(txBytes-c.prevNet.tx) / elapsed
		}
	}
	s.Network.RxBytes = rxBytes
	s.Network.TxBytes = txBytes
	c.prevNet = rawNetStat{rx: rxBytes, tx: txBytes}
	c.prevNetTime = now

	s.Temps = readTemperatures()

	return s
}

func readCPUStat() []rawCPUStat {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return nil
	}
	defer f.Close()

	var stats []rawCPUStat
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "cpu") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 8 {
			continue
		}
		label := fields[0]
		if label != "cpu" && (len(label) < 4) {
			continue
		}

		var st rawCPUStat
		vals := make([]uint64, 8)
		for i := 0; i < 8 && i+1 < len(fields); i++ {
			vals[i], _ = strconv.ParseUint(fields[i+1], 10, 64)
		}
		st.user, st.nice, st.system, st.idle, st.iowait, st.irq, st.softirq, st.steal =
			vals[0], vals[1], vals[2], vals[3], vals[4], vals[5], vals[6], vals[7]
		stats = append(stats, st)
	}
	return stats
}

func calcCPUPercent(prev, curr rawCPUStat) float64 {
	prevTotal := prev.user + prev.nice + prev.system + prev.idle +
		prev.iowait + prev.irq + prev.softirq + prev.steal
	currTotal := curr.user + curr.nice + curr.system + curr.idle +
		curr.iowait + curr.irq + curr.softirq + curr.steal

	prevIdle := prev.idle + prev.iowait
	currIdle := curr.idle + curr.iowait

	totalDiff := float64(currTotal - prevTotal)
	idleDiff := float64(currIdle - prevIdle)

	if totalDiff == 0 {
		return 0
	}
	return (totalDiff - idleDiff) / totalDiff * 100
}

func readCPUMHz() []float64 {
	cpufreqBase := "/sys/devices/system/cpu"
	var result []float64

	for i := 0; ; i++ {
		freqPath := fmt.Sprintf("%s/cpu%d/cpufreq/scaling_cur_freq", cpufreqBase, i)
		data, err := os.ReadFile(freqPath)
		if err != nil {
			break
		}
		khz, err := strconv.ParseFloat(strings.TrimSpace(string(data)), 64)
		if err != nil {
			result = append(result, 0)
			continue
		}
		result = append(result, khz/1000.0)
	}

	if len(result) > 0 {
		return result
	}

	f, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return nil
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "cpu MHz") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				mhz, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
				if err == nil {
					result = append(result, mhz)
				}
			}
		}
	}
	return result
}

func readMemInfo() (ramTotal, ramUsed, swapTotal, swapUsed uint64) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return
	}
	defer f.Close()

	vals := make(map[string]uint64)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		key := strings.TrimSuffix(parts[0], ":")
		val, _ := strconv.ParseUint(parts[1], 10, 64)
		vals[key] = val * 1024
	}

	ramTotal = vals["MemTotal"]
	available := vals["MemAvailable"]
	if available > 0 {
		ramUsed = ramTotal - available
	} else {
		ramUsed = ramTotal - vals["MemFree"] - vals["Buffers"] - vals["Cached"]
	}
	swapTotal = vals["SwapTotal"]
	swapUsed = swapTotal - vals["SwapFree"]
	return
}

var skipFS = map[string]bool{
	"proc": true, "sysfs": true, "devtmpfs": true, "devpts": true,
	"tmpfs": true, "cgroup": true, "cgroup2": true, "pstore": true,
	"debugfs": true, "securityfs": true, "configfs": true, "fusectl": true,
	"hugetlbfs": true, "mqueue": true, "overlay": true, "squashfs": true,
	"ramfs": true, "autofs": true, "efivarfs": true, "bpf": true,
}

func readDisks() []DiskStat {
	f, err := os.Open("/proc/mounts")
	if err != nil {
		return nil
	}
	defer f.Close()

	seen := make(map[string]bool)
	var disks []DiskStat

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 3 {
			continue
		}
		device, mount, fstype := fields[0], fields[1], fields[2]

		if skipFS[fstype] {
			continue
		}
		if strings.HasPrefix(mount, "/proc") ||
			strings.HasPrefix(mount, "/sys") ||
			strings.HasPrefix(mount, "/dev") ||
			strings.HasPrefix(mount, "/run") {
			continue
		}

		if seen[device] {
			continue
		}
		seen[device] = true

		var st syscall.Statfs_t
		if err := syscall.Statfs(mount, &st); err != nil {
			continue
		}
		total := st.Blocks * uint64(st.Bsize)
		free := st.Bfree * uint64(st.Bsize)
		used := total - free
		if total == 0 {
			continue
		}

		disks = append(disks, DiskStat{
			Mount:   mount,
			Device:  filepath.Base(device),
			Total:   total,
			Used:    used,
			Free:    free,
			Percent: float64(used) / float64(total) * 100,
		})
	}
	return disks
}

func readNetDev() (rxTotal, txTotal uint64) {
	f, err := os.Open("/proc/net/dev")
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		if lineNum <= 2 {
			continue
		}
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		iface := strings.TrimSpace(parts[0])
		if iface == "lo" {
			continue
		}
		fields := strings.Fields(parts[1])
		if len(fields) < 9 {
			continue
		}
		rx, _ := strconv.ParseUint(fields[0], 10, 64)
		tx, _ := strconv.ParseUint(fields[8], 10, 64)
		rxTotal += rx
		txTotal += tx
	}
	return
}

func readTemperatures() []TempStat {
	var temps []TempStat
	seen := make(map[string]bool)

	zones, _ := filepath.Glob("/sys/class/thermal/thermal_zone*/temp")
	for _, path := range zones {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		milliC, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
		if err != nil || milliC <= 0 || milliC > 150000 {
			continue
		}
		tempC := float64(milliC) / 1000.0

		dir := filepath.Dir(path)
		name := filepath.Base(dir)
		if typeData, err := os.ReadFile(filepath.Join(dir, "type")); err == nil {
			name = strings.TrimSpace(string(typeData))
		}

		key := fmt.Sprintf("%s_%.1f", name, tempC)
		if seen[key] {
			continue
		}
		seen[key] = true
		temps = append(temps, TempStat{Name: name, Temp: tempC})
	}

	hwmonInputs, _ := filepath.Glob("/sys/class/hwmon/hwmon*/temp*_input")
	for _, path := range hwmonInputs {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		milliC, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
		if err != nil || milliC <= 0 || milliC > 150000 {
			continue
		}
		tempC := float64(milliC) / 1000.0

		base := strings.TrimSuffix(path, "_input")
		name := filepath.Base(base)
		if labelData, err := os.ReadFile(base + "_label"); err == nil {
			name = strings.TrimSpace(string(labelData))
		} else {
			hwmonDir := filepath.Dir(path)
			if nameData, err := os.ReadFile(filepath.Join(hwmonDir, "name")); err == nil {
				name = strings.TrimSpace(string(nameData)) + "/" + filepath.Base(base)
			}
		}

		key := fmt.Sprintf("%s_%.1f", name, tempC)
		if seen[key] {
			continue
		}
		seen[key] = true
		temps = append(temps, TempStat{Name: name, Temp: tempC})
	}

	return temps
}

func readUptime() string {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return "N/A"
	}
	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return "N/A"
	}
	secs, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return "N/A"
	}
	total := int(secs)
	days := total / 86400
	hours := (total % 86400) / 3600
	mins := (total % 3600) / 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, mins)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	return fmt.Sprintf("%dm", mins)
}
