package main

import (
	"bytes"
	"flag"
	"io"
	"os"
	"strings"
	"testing"
)

// save original command line arguments
var originalArgs []string

func TestMain(m *testing.M) {
	// Save original command-line arguments
	originalArgs = os.Args

	// Run tests
	exitCode := m.Run()

	// Restore original command-line arguments
	os.Args = originalArgs

	os.Exit(exitCode)
}

// setup test environment
func setup() {
	// Reset flags for each test
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	monochromeMode = false // Reset global flag
	os.Args = originalArgs
}

// Helper to capture stdout
func captureStdout(f func()) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestRamGaugeOutput(t *testing.T) {
	setup()
	output := captureStdout(func() {
		main()
	})

	if !strings.Contains(output, "RAM") {
		t.Errorf("Expected output to contain 'RAM' but got: %s", output)
	}

	// Basic check for the gauge format.
	// This might need to be more robust if the gauge format changes,
	// but for now, we just check for the presence of the gauge characters.
	if !strings.Contains(output, "[") || !strings.Contains(output, "]") || (!strings.Contains(output, "#") && !strings.Contains(output, "-")) {
		t.Errorf("Expected RAM gauge in output, but format not found: %s", output)
	}

	// Check for a percentage and color reset code
	if !strings.Contains(output, "%") || !strings.Contains(output, "\033[0m") {
		t.Errorf("Expected a percentage and color reset code in RAM output: %s", output)
	}
}

func TestMonochromeMode(t *testing.T) {
	setup()
	os.Args = []string{originalArgs[0], "-m"}

	output := captureStdout(func() {
		main()
	})

	// Check that no ANSI color codes are present
	if strings.Contains(output, "\033[3") || strings.Contains(output, "\033[0m") {
		t.Errorf("Monochrome mode: Expected no ANSI color codes, but found them in output: %s", output)
	}

	// Ensure basic gauge format is still present
	if !strings.Contains(output, "[") || !strings.Contains(output, "]") || (!strings.Contains(output, "#") && !strings.Contains(output, "-")) {
		t.Errorf("Monochrome mode: Expected RAM gauge in output, but format not found: %s", output)
	}
}
