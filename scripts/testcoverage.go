package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// Global command runner to allow mocking exec.Command in unit tests
var execCommand = exec.Command

func main() {
	if err := run(os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "COVERAGE CHECK FAILED: %v\n", err)
		os.Exit(1)
	}
}

// main run logic for the script
func run(stdout io.Writer) error {
	// Names for the files produced during execution
	rawProfile := "coverage_raw.out"
	filteredProfile := "coverage.out"
	htmlReport := "coverage.html"

	// Remove the raw coverage file
	defer os.Remove(rawProfile)

	// Information for the user
	fmt.Fprintln(stdout, "running tests and generating raw coverage profile...")
	// In the terminal, the code executes the go command to test the project and obtain the cover profile
	cmd := execCommand("go", "test", "-coverprofile="+rawProfile, "./...")
	// We print the results and possible errors
	cmd.Stdout = stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running tests: %w", err)
	}

	// Information to the user about files not considered in the final count
	fmt.Fprintln(stdout, "filtering out wizard files from coverage data...")
	if err := filterFiles(rawProfile, filteredProfile); err != nil {
		return fmt.Errorf("error filtering coverage profile: %w", err)
	}

	// Information to the user about the coverage summary for each package after filtering out the wizards
	fmt.Fprintln(stdout, "\nadjusted coverage summary per package:")
	// We execute the go coverage command over the filtered results and print them (or errors incurred)
	summaryCmd := execCommand("go", "tool", "cover", "-func="+filteredProfile)
	summaryOutput, err := summaryCmd.Output()
	if err != nil {
		return fmt.Errorf("error running cover tool: %w", err)
	}

	summaryStr := string(summaryOutput)
	fmt.Fprintln(stdout, summaryStr)

	// Extract and check overall coverage percentage
	totalCoverage, err := parseTotalCoverage(summaryStr)
	if err != nil {
		return fmt.Errorf("failed to parse total coverage: %w", err)
	}

	fmt.Fprintln(stdout, "generating HTML coverage report...")
	// We execute the go command to transform the filtered coverage results into an html and print a success or error message
	htmlCmd := execCommand("go", "tool", "cover", "-html="+filteredProfile, "-o", htmlReport)
	if err := htmlCmd.Run(); err != nil {
		return fmt.Errorf("error generating HTML coverage report: %w", err)
	}
	fmt.Fprintf(stdout, "success! Filtered HTML report saved to: %s\n", htmlReport)

	// Validate coverage threshold
	const minCoverage = 90.0
	if totalCoverage < minCoverage {
		return fmt.Errorf("total coverage %.1f%% is below the required threshold of %.1f%%", totalCoverage, minCoverage)
	}

	fmt.Fprintf(stdout, "PASSED: Total coverage is %.1f%% (Threshold: %.1f%%)\n", totalCoverage, minCoverage)
	return nil
}

// parseTotalCoverage extracts the final percentage value from `go tool cover -func` output.
func parseTotalCoverage(output string) (float64, error) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "total:") {
			fields := strings.Fields(line)
			if len(fields) < 3 {
				return 0, fmt.Errorf("malformed total line: %s", line)
			}
			// fields[2] contains the percentage
			rawPercent := strings.TrimSuffix(fields[len(fields)-1], "%")
			return strconv.ParseFloat(rawPercent, 64)
		}
	}
	return 0, fmt.Errorf("total coverage summary line not found")
}

func filterFiles(inputPath, outputPath string) error {
	inFile, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer inFile.Close()

	var outBuf bytes.Buffer
	scanner := bufio.NewScanner(inFile)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "mode:") {
			outBuf.WriteString(line)
			outBuf.WriteString("\n")
			continue
		}

		// Skip any file in /scripts/ or ending with _wizard.go / _wizards.go
		// This is due to the difficulties around testing user interactive forms
		if strings.Contains(line, "/scripts/") ||
			strings.Contains(line, "_wizard.go:") ||
			strings.Contains(line, "_wizards.go:") {
			continue
		}

		outBuf.WriteString(line)
		outBuf.WriteString("\n")
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return os.WriteFile(outputPath, outBuf.Bytes(), 0644)
}
