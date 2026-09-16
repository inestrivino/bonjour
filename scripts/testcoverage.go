package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// Global command runner to allow mocking exec.Command in unit tests
var execCommand = exec.Command

func main() {
	if err := run(os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

// main run logic for the script
func run(stdout io.Writer) error {
	// Names for the files produced during execution
	rawProfile := "coverage_raw.out"
	filteredProfile := "coverage_filtered.out"
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
	fmt.Fprintln(stdout, string(summaryOutput))

	// Information to the user about the HTML coverage report being produced
	fmt.Fprintln(stdout, "generating HTML coverage report...")
	// We execute the go command to transform the filtered coverage results into an html and print a success or error message
	htmlCmd := execCommand("go", "tool", "cover", "-html="+filteredProfile, "-o", htmlReport)
	if err := htmlCmd.Run(); err != nil {
		return fmt.Errorf("error generating HTML coverage report: %w", err)
	}
	fmt.Fprintf(stdout, "success! Filtered HTML report saved to: %s\n", htmlReport)

	return nil
}

// Helper function to filter out files that shouldn't be counted into the overall coverage percentage as they can't be tested
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
