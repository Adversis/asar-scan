package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// ScanForElectronApps searches the system for Electron applications
func ScanForElectronApps(verbose bool) ([]string, error) {
	if runtime.GOOS == "darwin" {
		return scanForElectronAppsMacos(verbose)
	} else {
		return scanForElectronAppsWindows(verbose)
	}
}

// scanForElectronAppsMacos searches macOS for Electron applications
func scanForElectronAppsMacos(verbose bool) ([]string, error) {
	var appPaths []string

	// Common locations for applications on macOS
	searchDirs := []string{
		"/Applications",
		filepath.Join(os.Getenv("HOME"), "Applications"),
	}

	for _, dir := range searchDirs {
		if verbose {
			fmt.Printf("Scanning directory: %s\n", dir)
		}

		// Check if directory exists
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if verbose {
				fmt.Printf("Directory does not exist: %s\n", dir)
			}
			continue
		}

		// Walk the directory looking for .app bundles
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				if verbose {
					fmt.Printf("Error accessing path %s: %v\n", path, err)
				}
				return nil // Continue despite error
			}

			// Check for .app directories (bundles)
			if info.IsDir() && strings.HasSuffix(path, ".app") {
				if verbose {
					fmt.Printf("Found app bundle: %s\n", path)
				}
				appPaths = append(appPaths, path)
			}

			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("error scanning directory %s: %v", dir, err)
		}
	}

	return appPaths, nil
}

// scanForElectronAppsWindows searches Windows for Electron applications
func scanForElectronAppsWindows(verbose bool) ([]string, error) {
	var appPaths []string

	// Common locations for applications on Windows
	searchDirs := []string{
		filepath.Join(os.Getenv("ProgramFiles")),
		filepath.Join(os.Getenv("ProgramFiles(x86)")),
		filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs"),
	}

	for _, dir := range searchDirs {
		if verbose {
			fmt.Printf("Scanning directory: %s\n", dir)
		}

		// Check if directory exists
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if verbose {
				fmt.Printf("Directory does not exist: %s\n", dir)
			}
			continue
		}

		// Walk the directory looking for potential Electron apps
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				if verbose {
					fmt.Printf("Error accessing path %s: %v\n", path, err)
				}
				return nil // Continue despite error
			}

			// Look for .exe files or directories containing them
			if !info.IsDir() && strings.HasSuffix(path, ".exe") {
				resourcesDir := filepath.Join(filepath.Dir(path), "resources")
				if _, err := os.Stat(resourcesDir); err == nil {
					if verbose {
						fmt.Printf("Found potential Electron app: %s\n", path)
					}
					appPaths = append(appPaths, path)
				}
			}

			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("error scanning directory %s: %v", dir, err)
		}
	}

	return appPaths, nil
}
