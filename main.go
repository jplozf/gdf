package main

// ****************************************************************************
// IMPORTS
// ****************************************************************************
import (
	"fmt"
	"log"
	runtime "runtime"
	"strings"
	"syscall"
	"time"

	// "time"
	"flag"

	"github.com/moby/sys/mountinfo"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
)

// ****************************************************************************
// GLOBALS
// ****************************************************************************
var (
	flagMonochrome  bool
	flagFilesystems bool
	flagRAM         bool
	flagCPU         bool
	flagAll         bool
	flagWatch       int
)

var ignoredFSTypes = map[string]bool{
	"tmpfs":           true,
	"devtmpfs":        true,
	"proc":            true,
	"sysfs":           true,
	"cgroup2":         true,
	"securityfs":      true,
	"pstore":          true,
	"efivarfs":        true,
	"bpf":             true,
	"configfs":        true,
	"autofs":          true,
	"debugfs":         true,
	"hugetlbfs":       true,
	"tracefs":         true,
	"mqueue":          true,
	"fusectl":         true,
	"binfmt_misc":     true,
	"rpc_pipefs":      true,
	"overlay":         true, // Docker overlays
	"squashfs":        true, // Snap packages
	"nsfs":            true, // Docker namespaces
	"fuse.gvfsd-fuse": true, // Gnome virtual file system
	"fuse.portal":     true, // Gnome portal
	"devpts":          true, // Pseudo-filesystem for terminals
	"selinuxfs":       true, // SELinux filesystem
}

const colorReset = "\033[0m"

// ****************************************************************************
// displayMetrics()
// ****************************************************************************
func displayMetrics(showDisks, showRAM, showCPU bool) {
	// --- CPU Load Calculation ---
	numCPU := runtime.NumCPU()
	cpuLoadAvg, loadErr := load.Avg() // Corrected to use load.Avg()
	var cpu1Percent, cpu5Percent, cpu15Percent float64
	if loadErr != nil {
		log.Printf("Failed to get CPU load average: %v", loadErr)
	} else {
		// Ensure we don't divide by zero if numCPU is 0 (though unlikely)
		if numCPU > 0 {
			cpu1Percent = (cpuLoadAvg.Load1 / float64(numCPU)) * 100
			cpu5Percent = (cpuLoadAvg.Load5 / float64(numCPU)) * 100
			cpu15Percent = (cpuLoadAvg.Load15 / float64(numCPU)) * 100
		}
	}
	// --- End CPU Load Calculation ---

	// Get mount information only if we need to display disks
	var mounts []*mountinfo.Info
	var err error
	if showDisks {
		mounts, err = mountinfo.GetMounts(nil)
		if err != nil {
			log.Printf("Failed to get mount information: %v", err) // Changed from Fatalf to Logf to allow other metrics to be displayed
		}
	}

	// Print Disk Metrics
	if showDisks {
		for _, m := range mounts {
			if !ignoredFSTypes[m.FSType] && !(strings.HasPrefix(m.FSType, "fuse.") && strings.Contains(m.FSType, "AppImage")) {
				var stat syscall.Statfs_t
				if err := syscall.Statfs(m.Mountpoint, &stat); err != nil {
					log.Printf("Failed to get statfs for %s: %v", m.Mountpoint, err)
					continue
				}

				blockSize := uint64(stat.Bsize)
				totalSpace := stat.Blocks * blockSize
				freeSpace := stat.Bfree * blockSize
				usedSpace := totalSpace - freeSpace

				usagePercent := 0.0
				if totalSpace > 0 {
					usagePercent = (float64(usedSpace) / float64(totalSpace)) * 100
				}

				gauge, color := generateGauge(usagePercent, 30, flagMonochrome)
				fmt.Printf("% -25s %10s %s %s%5.2f%%%s\n", m.Mountpoint, byteCountToHumanReadable(totalSpace), gauge, color, usagePercent, colorReset)
			}
		}
	}

	// Print RAM Metrics
	if showRAM {
		v, err := mem.VirtualMemory()
		if err != nil {
			log.Printf("Failed to get virtual memory information: %v", err)
		} else {
			gauge, color := generateGauge(v.UsedPercent, 30, flagMonochrome)
			fmt.Printf("% -25s %10s %s %s%5.2f%%%s\n", "RAM", byteCountToHumanReadable(v.Total), gauge, color, v.UsedPercent, colorReset)
		}
	}

	// Print CPU Metrics
	if showCPU {
		// Only print CPU gauges if load.Avg() was successful and numCPU > 0
		if loadErr == nil && numCPU > 0 {
			gauge1, color1 := generateGauge(cpu1Percent, 30, flagMonochrome)
			fmt.Printf("% -25s %10s %s %s%5.2f%%%s\n", "CPU", "1 mn", gauge1, color1, cpu1Percent, colorReset) // No size for CPU

			gauge5, color5 := generateGauge(cpu5Percent, 30, flagMonochrome)
			fmt.Printf("% -25s %10s %s %s%5.2f%%%s\n", "CPU", "5 mn", gauge5, color5, cpu5Percent, colorReset)

			gauge15, color15 := generateGauge(cpu15Percent, 30, flagMonochrome)
			fmt.Printf("% -25s %10s %s %s%5.2f%%%s\n", "CPU", "15 mn", gauge15, color15, cpu15Percent, colorReset)
		}
	}
}

// ****************************************************************************
// main()
// ****************************************************************************
func main() {
	flag.BoolVar(&flagMonochrome, "m", false, "Display output in monochrome without colors")
	flag.BoolVar(&flagFilesystems, "d", false, "Display file systems metrics")
	flag.BoolVar(&flagRAM, "r", false, "Display RAM metrics")
	flag.BoolVar(&flagCPU, "c", false, "Display CPU metrics")
	flag.BoolVar(&flagAll, "a", false, "Display all metrics")
	flag.IntVar(&flagWatch, "w", 0, "Watch every n seconds")
	// Removed watchInterval flag as the application will now exit after displaying metrics once.
	flag.Parse()

	// Determine which metrics to display based on flags
	var showDisks, showRAM, showCPU bool

	// If -a flag is present, show all metrics
	if flagAll {
		showDisks = true
		showRAM = true
		showCPU = true
	} else {
		// If -a is not present, use the specific flags.
		showDisks = flagFilesystems
		showRAM = flagRAM
		showCPU = flagCPU

		// If no specific flags were set (-d, -r, -c), default to showing all metrics.
		if !flagFilesystems && !flagRAM && !flagCPU {
			showDisks = true
			showRAM = true
			showCPU = true
		}
	}

	if flagWatch == 0 {
		displayMetrics(showDisks, showRAM, showCPU)
	} else {
		watchMetrics(showDisks, showRAM, showCPU)
		for _ = range time.Tick(time.Duration(flagWatch) * time.Second) {
			watchMetrics(showDisks, showRAM, showCPU)
		}
	}
}

// ****************************************************************************
// watchMetrics()
// ****************************************************************************
func watchMetrics(showDisks, showRAM, showCPU bool) {
	fmt.Print("\033[H\033[2J") // Clear screen before
	fmt.Printf("Refreshing every %d second(s). Press Crl+C to exit.\n", flagWatch)
	displayMetrics(showDisks, showRAM, showCPU)
}

// ****************************************************************************
// generateGauge()
// ****************************************************************************
// generateGauge creates a textual gauge representing disk usage with color coding.
func generateGauge(usage float64, width int, monochrome bool) (string, string) {
	if monochrome {
		var gaugeBuilder strings.Builder
		gaugeBuilder.WriteString("[")
		numFilled := int((usage / 100.0) * float64(width))
		for i := 0; i < width; i++ {
			if i < numFilled {
				gaugeBuilder.WriteString("#")
			} else {
				gaugeBuilder.WriteString("-")
			}
		}
		gaugeBuilder.WriteString("]")
		return gaugeBuilder.String(), ""
	}

	const ( // ANSI escape codes for colors
		colorGreen  = "\033[32m"
		colorYellow = "\033[33m"
		colorRed    = "\033[31m"
		// colorReset  = "\033[0m" // Already defined globally
	)

	var gaugeBuilder strings.Builder
	gaugeBuilder.WriteString("[")
	numFilled := int((usage / 100.0) * float64(width))
	for i := 0; i < width; i++ {
		// Determine the percentage for the current segment
		segmentEndPercentage := (float64(i+1) / float64(width)) * 100
		var segmentColor string
		if segmentEndPercentage <= 50 {
			segmentColor = colorGreen
		} else if segmentEndPercentage <= 80 {
			segmentColor = colorYellow
		} else {
			segmentColor = colorRed
		}
		if i < numFilled {
			gaugeBuilder.WriteString(segmentColor + "#" + colorReset)
		} else {
			gaugeBuilder.WriteString("-")
		}
	}

	gaugeBuilder.WriteString("]" + colorReset)
	var overallColor string
	if usage < 50 {
		overallColor = colorGreen
	} else if usage < 80 {
		overallColor = colorYellow
	} else {
		overallColor = colorRed
	}
	return gaugeBuilder.String(), overallColor
}

// ****************************************************************************
// byteCountToHumanReadable()
// ****************************************************************************
// byteCountToHumanReadable converts a byte count to a human-readable string (e.g., 10GB, 2.5MB)
func byteCountToHumanReadable(b uint64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}

	div, exp := uint64(unit), 0
	for b >= div*unit && exp < len(numericUnits)-1 {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %s", float64(b)/float64(div), numericUnits[exp])
}

var numericUnits = []string{"kB", "MB", "GB", "TB", "PB", "EB"}
