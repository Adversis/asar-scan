package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/adversis/electron-integrity/cmd/asarscan/internal"
)

// Version is set during build via ldflags
var version = "dev"

// AppResult contains the result of checking an application
type AppResult struct {
	Path             string   `json:"path"`
	IsElectron       bool     `json:"is_electron"`
	Version          string   `json:"electron_version,omitempty"`
	HasAsarFile      bool     `json:"has_asar_file"`
	AsarIntegrity    bool     `json:"asar_integrity_enabled"`
	OnlyLoadFromAsar bool     `json:"only_load_from_asar"`
	NodeFiles        []string `json:"node_files,omitempty"`
	IntegrityError   string   `json:"integrity_error,omitempty"`
}

func main() {
	// Parse command-line flags
	verbose := flag.Bool("verbose", false, "Enable verbose output")
	outputJson := flag.Bool("json", false, "Output results in JSON format")
	listNodeFiles := flag.Bool("node-files", true, "List .node files in Electron applications")
	maxNodeFiles := flag.Int("max-node-files", 5, "Maximum number of .node files to list per application (0 for unlimited)")
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	// Show version and exit if requested
	if *showVersion {
		fmt.Printf("Electron ASAR Integrity Scanner v%s\n", version)
		os.Exit(0)
	}

	fmt.Println("Electron ASAR Integrity Scanner v" + version)
	fmt.Println("-------------------------------")

	// Check if we're running on a supported OS
	if runtime.GOOS != "darwin" && runtime.GOOS != "windows" {
		fmt.Fprintf(os.Stderr, "Error: Unsupported operating system: %s. This tool only works on macOS and Windows.\n", runtime.GOOS)
		os.Exit(1)
	}

	fmt.Printf("Scanning %s system for Electron applications...\n", runtime.GOOS)

	// Scan for Electron applications
	apps, err := internal.ScanForElectronApps(*verbose)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning for applications: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d potential Electron applications\n", len(apps))

	// Check ASAR integrity for each application
	var results []internal.AppResult

	for _, app := range apps {
		if *verbose {
			fmt.Printf("Checking ASAR integrity for: %s\n", app)
		}

		result := internal.CheckAsarIntegrityForApp(app, *verbose)

		// Find .node files if requested
		if *listNodeFiles && result.IsElectron {
			if *verbose {
				fmt.Printf("Searching for .node files in: %s\n", app)
			}
			result.NodeFiles = internal.FindNodeFiles(app, *maxNodeFiles, *verbose)
		}

		results = append(results, result)
	}

	// Output results
	if *outputJson {
		outputResultsJson(results)
	} else {
		outputResultsText(results, *listNodeFiles)
	}
}

// outputResultsJson outputs the results in JSON format
func outputResultsJson(results []internal.AppResult) {
	jsonData, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
		return
	}

	fmt.Println(string(jsonData))
}

// outputResultsText outputs the results in human-readable text format
func outputResultsText(results []internal.AppResult, showNodeFiles bool) {
	fmt.Println("\nResults:")
	fmt.Println("========")

	// First output detailed results
	index := 1
	for _, result := range results {
		// Skip non-Electron apps
		if !result.IsElectron {
			continue
		}

		fmt.Printf("\n[%d] %s\n", index, result.Path)
		fmt.Printf("  Is Electron App: %t\n", result.IsElectron)
		fmt.Printf("  Electron Version: %s\n", result.Version)
		fmt.Printf("  Has ASAR File: %t\n", result.HasAsarFile)

		if result.HasAsarFile {
			fmt.Printf("  ASAR Integrity Enabled: %t\n", result.AsarIntegrity)
			fmt.Printf("  OnlyLoadFromAsar Enabled: %t\n", result.OnlyLoadFromAsar)

			if result.IntegrityError != "" {
				fmt.Printf("  Error: %s\n", result.IntegrityError)
			}
		}

		// Show .node files if available
		if showNodeFiles && len(result.NodeFiles) > 0 {
			fmt.Printf("  .node Files (%d found):\n", len(result.NodeFiles))
			for i, nodeFile := range result.NodeFiles {
				// Print the full path as requested by the user
				fmt.Printf("    %d. %s\n", i+1, nodeFile)
			}
		}

		index++
	}

	// Summary statistics
	electronCount := 0
	asarCount := 0
	integrityCount := 0
	onlyLoadCount := 0

	for _, result := range results {
		if result.IsElectron {
			electronCount++
			if result.HasAsarFile {
				asarCount++
				if result.AsarIntegrity {
					integrityCount++
				}
				if result.OnlyLoadFromAsar {
					onlyLoadCount++
				}
			}
		}
	}

	fmt.Printf("\nSummary:\n")
	fmt.Printf("  Total apps scanned: %d\n", len(results))
	fmt.Printf("  Electron apps: %d\n", electronCount)
	fmt.Printf("  Apps with ASAR files: %d\n", asarCount)
	fmt.Printf("  Apps with ASAR integrity enabled: %d\n", integrityCount)
	fmt.Printf("  Apps with OnlyLoadAppFromAsar enabled: %d\n", onlyLoadCount)

	// Add a table summary of Electron apps
	fmt.Printf("\nSummary Table:\n")
	fmt.Printf("===================================================================================\n")
	fmt.Printf("%-30s | %-10s | %-10s | %-10s | %-15s\n", "Application", "Version", "ASAR File", "Integrity", "OnlyLoadAppFromAsar")
	fmt.Printf("===================================================================================\n")

	// Only include electron apps in the table
	for _, result := range results {
		if result.IsElectron {
			// Format the version string better
			version := result.Version
			if version == "" || version == "unknown" {
				version = "Unknown"
			} else if version == "detected" {
				version = "✓" // Checkmark indicates version detected but not parsed
			}

			// Format has ASAR and integrity as yes/no
			hasAsar := "No"
			if result.HasAsarFile {
				hasAsar = "Yes"
			}

			integrity := "N/A"
			if result.HasAsarFile {
				if result.AsarIntegrity {
					integrity = "Yes"
				} else {
					integrity = "No"
				}
			}

			onlyLoad := "N/A"
			if result.HasAsarFile {
				if result.OnlyLoadFromAsar {
					onlyLoad = "Yes"
				} else {
					onlyLoad = "No"
				}
			}

			// Get the app name from the full path
			appName := filepath.Base(result.Path)
			if len(appName) > 28 {
				appName = appName[:25] + "..."
			}

			fmt.Printf("%-30s | %-10s | %-10s | %-10s | %-15s\n", appName, version, hasAsar, integrity, onlyLoad)
		}
	}
	fmt.Printf("===================================================================================\n")
}
