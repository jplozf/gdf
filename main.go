package main

import (
	"fmt"
	"log"
	"strings"
	"syscall"

	"github.com/moby/sys/mountinfo"
)

func main() {

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

		"devpts":                                                  true, // Pseudo-filesystem for terminals
		"selinuxfs":                                               true, // SELinux filesystem
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

			// fmt.Printf("Mount Point: %-25s FSType: %-10s Source: %-20s Total: %-10s Used: %-10s Free: %-10s Usage: %.2f%%\n",
			fmt.Printf("%-25s %10s %s %5.2f%%\n", m.Mountpoint, byteCountToHumanReadable(totalSpace), generateGauge(usagePercent, 30), usagePercent)
		}
	}
}

// generateGauge creates a textual gauge representing disk usage.
func generateGauge(usage float64, width int) string {
	numFilled := int((usage / 100.0) * float64(width))
	numEmpty := width - numFilled

	filled := ""
	for i := 0; i < numFilled; i++ {
		filled += "#"
	}

	empty := ""
	for i := 0; i < numEmpty; i++ {
		empty += "-"
	}

	return fmt.Sprintf("[%s%s]", filled, empty)
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
