package internal

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

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

// CheckAsarIntegrityForApp checks if ASAR integrity is enabled for a specific app
func CheckAsarIntegrityForApp(appPath string, verbose bool) AppResult {
	result := AppResult{
		Path: appPath,
	}

	// Check if it's an Electron app
	isElectron, version, err := IsElectronApp(appPath, verbose)
	if err != nil {
		result.IntegrityError = err.Error()
		return result
	}
	result.IsElectron = isElectron
	result.Version = version

	if !isElectron {
		if verbose {
			fmt.Printf("%s is not an Electron app\n", appPath)
		}
		return result
	}

	// Check if it has app.asar file
	result.HasAsarFile = HasAsarFile(appPath)
	if !result.HasAsarFile {
		if verbose {
			fmt.Printf("%s has no app.asar file\n", appPath)
		}
		return result
	}

	// Check for ASAR integrity and OnlyLoadFromAsar
	switch runtime.GOOS {
	case "darwin":
		hasIntegrity, onlyLoadFromAsar, err := checkAsarIntegrityMacos(appPath, verbose)
		result.AsarIntegrity = hasIntegrity
		result.OnlyLoadFromAsar = onlyLoadFromAsar
		if err != nil {
			result.IntegrityError = err.Error()
		}
	case "windows":
		hasIntegrity, onlyLoadFromAsar, err := checkAsarIntegrityWindows(appPath, verbose)
		result.AsarIntegrity = hasIntegrity
		result.OnlyLoadFromAsar = onlyLoadFromAsar
		if err != nil {
			result.IntegrityError = err.Error()
		}
	default:
		result.IntegrityError = "unsupported operating system"
	}

	return result
}

// checkAsarIntegrityMacos checks if ASAR integrity is enabled on macOS
func checkAsarIntegrityMacos(appPath string, verbose bool) (bool, bool, error) {
	// Check for 'ElectronAsarIntegrity' key in Info.plist
	plistPath := filepath.Join(appPath, "Contents", "Info.plist")

	if verbose {
		fmt.Printf("Checking Info.plist for ElectronAsarIntegrity: %s\n", plistPath)
	}

	// Read the Info.plist file
	plistContent, err := os.ReadFile(plistPath)
	if err != nil {
		return false, false, fmt.Errorf("error reading Info.plist: %v", err)
	}

	// Initialize result flags
	hasAsarIntegrity := false
	hasOnlyLoadFromAsar := false

	// Check for ElectronAsarIntegrity key in the contents
	if bytes.Contains(plistContent, []byte("<key>ElectronAsarIntegrity</key>")) {
		if verbose {
			fmt.Println("  Found ElectronAsarIntegrity key in Info.plist")
		}

		// Check if there's a hash value in the integrity dictionary
		if bytes.Contains(plistContent, []byte("<key>hash</key>")) &&
			bytes.Contains(plistContent, []byte("<key>algorithm</key>")) {
			if verbose {
				fmt.Println("  Found hash and algorithm keys - ASAR integrity appears properly configured")
			}
			hasAsarIntegrity = true
		} else {
			if verbose {
				fmt.Println("  ElectronAsarIntegrity key exists but hash/algorithm missing - may be misconfigured")
			}
			// Still return true since the integrity key exists
			hasAsarIntegrity = true
		}
	}

	// Check for OnlyLoadAppFromAsar fuse
	// There are multiple places this could be indicated in the app

	// Binary signature for OnlyLoadAppFromAsar
	executablePath := filepath.Join(appPath, "Contents", "MacOS", filepath.Base(strings.TrimSuffix(appPath, ".app")))
	if _, err := os.Stat(executablePath); err == nil {
		if verbose {
			fmt.Printf("  Checking executable for OnlyLoadAppFromAsar fuse: %s\n", executablePath)
		}

		execContent, err := os.ReadFile(executablePath)
		if err == nil {
			// Look for OnlyLoadAppFromAsar signature
			if bytes.Contains(execContent, []byte("OnlyLoadAppFromAsar")) {
				if verbose {
					fmt.Println("  Found OnlyLoadAppFromAsar fuse signature in executable")
				}
				hasOnlyLoadFromAsar = true
			}
		}
	}

	// Also check Info.plist for any indicators
	if !hasOnlyLoadFromAsar {
		signatures := []string{
			"OnlyLoadAppFromAsar",
			"OnlyLoadFromAsar",
			"FuseV1Options.OnlyLoadAppFromAsar",
		}

		for _, sig := range signatures {
			if bytes.Contains(plistContent, []byte(sig)) {
				if verbose {
					fmt.Printf("  Found %s reference in Info.plist\n", sig)
				}
				hasOnlyLoadFromAsar = true
				break
			}
		}
	}

	if verbose {
		if !hasAsarIntegrity {
			fmt.Println("  No ElectronAsarIntegrity key found in Info.plist")
		}
		if !hasOnlyLoadFromAsar {
			fmt.Println("  No OnlyLoadAppFromAsar fuse detected")
		}
	}

	return hasAsarIntegrity, hasOnlyLoadFromAsar, nil
}

// checkAsarIntegrityWindows checks if ASAR integrity is enabled on Windows
func checkAsarIntegrityWindows(appPath string, verbose bool) (bool, bool, error) {
	// On Windows, we need to check resource entries for ElectronAsar
	exePath := appPath
	if !strings.HasSuffix(exePath, ".exe") {
		exePath = filepath.Join(appPath, filepath.Base(appPath)+".exe")
	}

	if verbose {
		fmt.Printf("Checking for ASAR integrity in Windows executable: %s\n", exePath)
	}

	// Since we can't directly read resource entries in Go without C bindings or external tools,
	// we use basic binary content checking.
	// For a production tool, using a proper Windows resource parser would be better.
	exeContent, err := os.ReadFile(exePath)
	if err != nil {
		return false, false, fmt.Errorf("error reading executable: %v", err)
	}

	// Initialize result flags
	hasAsarIntegrity := false
	hasOnlyLoadFromAsar := false

	// Look for more specific signatures of ASAR integrity
	asarIntegritySignatures := [][]byte{
		[]byte("ElectronAsar"),
		[]byte("Integrity"),
		[]byte("sha256"), // Common hash algorithm used
	}

	// Count how many signatures we find - more matches increases confidence
	matchCount := 0
	for _, sig := range asarIntegritySignatures {
		if bytes.Contains(exeContent, sig) {
			matchCount++
			if verbose {
				fmt.Printf("  Found integrity signature: %s\n", string(sig))
			}
		}
	}

	// Look for EnableEmbeddedAsarIntegrityValidation which is specific to ASAR integrity
	if bytes.Contains(exeContent, []byte("EnableEmbeddedAsarIntegrityValidation")) {
		matchCount += 2 // This is a very strong indicator
		if verbose {
			fmt.Println("  Found EnableEmbeddedAsarIntegrityValidation signature")
		}
	}

	// Check for OnlyLoadAppFromAsar fuse
	if bytes.Contains(exeContent, []byte("OnlyLoadAppFromAsar")) {
		if verbose {
			fmt.Println("  Found OnlyLoadAppFromAsar fuse signature")
		}
		hasOnlyLoadFromAsar = true
	} else {
		// Check for alternative spellings or implementations
		onlyLoadSignatures := [][]byte{
			[]byte("OnlyLoadFromAsar"),
			[]byte("FuseV1Options.OnlyLoadAppFromAsar"),
		}

		for _, sig := range onlyLoadSignatures {
			if bytes.Contains(exeContent, sig) {
				if verbose {
					fmt.Printf("  Found alternative OnlyLoadFromAsar signature: %s\n", string(sig))
				}
				hasOnlyLoadFromAsar = true
				break
			}
		}
	}

	// If we found at least 2 signatures, consider it likely to have ASAR integrity
	if matchCount >= 2 {
		if verbose {
			fmt.Println("  Multiple ASAR integrity indicators found - likely enabled")
		}
		hasAsarIntegrity = true
	}

	if verbose {
		if !hasAsarIntegrity {
			fmt.Println("  No strong ASAR integrity indicators found")
		}
		if !hasOnlyLoadFromAsar {
			fmt.Println("  No OnlyLoadAppFromAsar fuse detected")
		}
	}

	return hasAsarIntegrity, hasOnlyLoadFromAsar, nil
}

// checkForFusesEnabled checks if the Electron fuses for ASAR integrity are enabled
func checkForFusesEnabled(appPath string, verbose bool) (bool, error) {
	// This would require binary analysis which is complex
	// For a complete solution, you might need to use specific tools or libraries
	// For now, we'll return a placeholder
	if verbose {
		fmt.Println("Checking for Electron fuses is not yet implemented")
	}
	return false, nil
}
