package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	// Define the min-coverage flag
	minCoverage := flag.Float64("min-coverage", 70.0, "Minimum test coverage percentage required")
	flag.Parse()

	fmt.Printf("\033[1mRunning tests and checking coverage (Minimum: %.1f%%)...\033[0m\n", *minCoverage)

	// Get the list of packages
	pkgsCmd := exec.Command("go", "list", "./app/...")
	pkgsOutput, err := pkgsCmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error listing packages:", err)
		os.Exit(1)
	}

	pkgs := strings.Split(strings.TrimSpace(string(pkgsOutput)), "\n")

	var totalCoverage float64
	var totalPkg int
	var exitCode int

	for _, pkg := range pkgs {
		// Run go test with coverage
		testCmd := exec.Command("go", "test", "-coverprofile=coverage.out", pkg)
		testOutput, err := testCmd.CombinedOutput()
		if err != nil {
			fmt.Printf("\033[1;31mfail\t%s\terror running tests: %s\033[0m\n", pkg, err)
			exitCode = 1
			continue
		}

		result := string(testOutput)
		re := regexp.MustCompile(`[0-9]*\.[0-9]*%`)
		coverageMatch := re.FindString(result)

		if coverageMatch == "" {
			fmt.Printf("\033[1;33m?\t%s\t[no test files]\033[0m\n", pkg)
			continue
		}

		coverage, err := strconv.ParseFloat(strings.TrimSuffix(coverageMatch, "%"), 64)
		if err != nil {
			fmt.Printf("Error parsing coverage for package %s: %v\n", pkg, err)
			exitCode = 1
			continue
		}

		// Check if coverage meets the threshold
		if coverage >= *minCoverage {
			fmt.Printf("\033[1;32mok\t%s\tcoverage: %.1f%% of statements. Passed.\033[0m\n", pkg, coverage)
		} else {
			fmt.Printf("\033[1;31mfail\t%s\tcoverage: %.1f%% of statements. Failed (below %.1f%%).\033[0m\n", pkg, coverage, *minCoverage)
			exitCode = 1
		}

		totalCoverage += coverage
		totalPkg++
	}

	// Calculate and display overall coverage
	if totalPkg > 0 {
		averageCoverage := totalCoverage / float64(totalPkg)
		fmt.Printf("\033[1mOverall coverage is %.1f%%. Push allowed if all individual packages are above %.1f%%.\033[0m\n", averageCoverage, *minCoverage)
	} else {
		fmt.Println("\033[1;31mNo tests found in any package.\033[0m")
	}

	os.Exit(exitCode)
}
