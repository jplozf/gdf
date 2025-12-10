package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"syscall"

	"github.com/moby/sys/mountinfo"
	"github.com/shirou/gopsutil/v3/mem"
)

var monochromeMode bool
var onlyFilesystems bool

func main() {
	flag.BoolVar(&monochromeMode, "m", false, "Display output in monochrome without colors")
	flag.BoolVar(&onlyFilesystems, "d", false, "Display only file systems and hide RAM usage") // New flag definition
	flag.Parse()

	mounts, err := mountinfo.GetMounts(nil)
	if err != nil {
		log.Fatalf("Failed to get mount information: %v", err)
	}

	ignoredFSTypes := map[string]bool{
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

		"devpts":    true, // Pseudo-filesystem for terminals
		"selinuxfs": true, // SELinux filesystem
	}

	for _, m := range mounts {
		if !ignoredFSTypes[m.FSType] && !(strings.HasPrefix(m.FSType, "fuse.") && strings.Contains(m.FSType, "AppImage")) {
			var stat syscall.Statfs_t
			if err := syscall.Statfs(m.Mountpoint, &stat); err != nil {
				log.Printf("Failed to get statfs for %s: %v", m.Mountpoint, err)
				continue
			}

			// Blocks, Bsize, Bfree, Bavail are in 512-byte units, need to convert to bytes
			blockSize := uint64(stat.Bsize)
			totalSpace := stat.Blocks * blockSize
			freeSpace := stat.Bfree * blockSize
			// availableSpace := stat.Bavail * blockSize // Blocks available to non-root user
			usedSpace := totalSpace - freeSpace

			usagePercent := 0.0
			if totalSpace > 0 {
				usagePercent = (float64(usedSpace) / float64(totalSpace)) * 100
			}

			gauge, color := generateGauge(usagePercent, 30, monochromeMode)

			fmt.Printf("%-25s %10s %s %s%5.2f%%%s\n", m.Mountpoint, byteCountToHumanReadable(totalSpace), gauge, color, usagePercent, "\033[0m")

		}

	}

	// Add RAM gauge (conditionally)
	if !onlyFilesystems { // Check if the -d flag is NOT set
		v, err := mem.VirtualMemory()

		if err != nil {

			log.Printf("Failed to get virtual memory information: %v", err)

		} else {

			gauge, color := generateGauge(v.UsedPercent, 30, monochromeMode)

			fmt.Printf("%-25s %10s %s %s%5.2f%%%s\n", "RAM", byteCountToHumanReadable(v.Total), gauge, color, v.UsedPercent, "\033[0m")

		}
	}
}

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

		colorGreen = "\033[32m"

		colorYellow = "\033[33m"

		colorRed = "\033[31m"

		colorReset = "\033[0m"
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

	gaugeBuilder.WriteString("]")

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
