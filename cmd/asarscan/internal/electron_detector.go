package internal

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// IsElectronApp checks if the given path is an Electron application
func IsElectronApp(appPath string, verbose bool) (bool, string, error) {
	if verbose {
		fmt.Printf("Checking if %s is an Electron app...\n", appPath)
	}

	switch runtime.GOOS {
	case "darwin":
		return isElectronAppMacos(appPath, verbose)
	case "windows":
		return isElectronAppWindows(appPath, verbose)
	default:
		return false, "", errors.New("unsupported operating system")
	}
}

// isElectronAppMacos checks if the given path is an Electron application on macOS
func isElectronAppMacos(appPath string, verbose bool) (bool, string, error) {
	// Check for app bundle structure
	if !strings.HasSuffix(appPath, ".app") {
		if verbose {
			fmt.Printf("  Not an app bundle: %s\n", appPath)
		}
		return false, "", nil
	}

	// Look for the Info.plist
	plistPath := filepath.Join(appPath, "Contents", "Info.plist")
	if _, err := os.Stat(plistPath); os.IsNotExist(err) {
		if verbose {
			fmt.Printf("  No Info.plist found: %s\n", plistPath)
		}
		return false, "", nil
	}

	// Check for the Electron framework
	frameworkPath := filepath.Join(appPath, "Contents", "Frameworks", "Electron Framework.framework")
	if _, err := os.Stat(frameworkPath); err == nil {
		if verbose {
			fmt.Printf("  Found Electron Framework: %s\n", frameworkPath)
		}

		// Try to extract Electron version from Info.plist
		version := "unknown"
		plistContent, err := os.ReadFile(plistPath)
		if err == nil {
			plistStr := string(plistContent)

			// Try to find version-like strings
			versionRegexes := []string{
				`<key>ElectronVersion</key>\s*<string>([0-9.]+)`,
				`<key>CFBundleVersion</key>\s*<string>([0-9.]+)`,
				`Electron/([0-9.]+)`,
				`electron@([0-9.]+)`,
				`electron": "([^"]+)"`,
				`"electronVersion": "([^"]+)"`,
			}

			for _, regex := range versionRegexes {
				re := regexp.MustCompile(regex)
				matches := re.FindStringSubmatch(plistStr)
				if len(matches) > 1 {
					if verbose {
						fmt.Printf("  Found Electron version: %s\n", matches[1])
					}
					version = matches[1]
					break
				}
			}

			// Also check the framework's Info.plist
			frameworkPlistPath := filepath.Join(frameworkPath, "Resources", "Info.plist")
			if _, err := os.Stat(frameworkPlistPath); err == nil {
				if verbose {
					fmt.Printf("  Checking Electron framework Info.plist for version\n")
				}
				frameworkPlist, err := os.ReadFile(frameworkPlistPath)
				if err == nil {
					for _, regex := range versionRegexes {
						re := regexp.MustCompile(regex)
						matches := re.FindStringSubmatch(string(frameworkPlist))
						if len(matches) > 1 {
							if verbose {
								fmt.Printf("  Found Electron version in framework: %s\n", matches[1])
							}
							version = matches[1]
							break
						}
					}
				}
			}
		}

		return true, version, nil
	}

	// Check for app.asar file
	asarPath := filepath.Join(appPath, "Contents", "Resources", "app.asar")
	if _, err := os.Stat(asarPath); err == nil {
		if verbose {
			fmt.Printf("  Found app.asar: %s\n", asarPath)
		}

		// Try to extract version from package.json if it exists
		version := "unknown"
		packageJsonPath := filepath.Join(appPath, "Contents", "Resources", "app", "package.json")
		if _, err := os.Stat(packageJsonPath); err == nil {
			if verbose {
				fmt.Printf("  Found package.json, checking for Electron version\n")
			}

			packageContent, err := os.ReadFile(packageJsonPath)
			if err == nil {
				// Simple regex to find electron version in package.json
				re := regexp.MustCompile(`"electron":\s*"([^"]+)"`)
				matches := re.FindStringSubmatch(string(packageContent))
				if len(matches) > 1 {
					if verbose {
						fmt.Printf("  Found Electron version in package.json: %s\n", matches[1])
					}
					version = matches[1]
				}
			}
		}

		return true, version, nil
	}

	return false, "", nil
}

// isElectronAppWindows checks if the given path is an Electron application on Windows
func isElectronAppWindows(appPath string, verbose bool) (bool, string, error) {
	// Check for common Electron files
	exePath := appPath
	if !strings.HasSuffix(exePath, ".exe") {
		exePath = filepath.Join(appPath, filepath.Base(appPath)+".exe")
	}

	if _, err := os.Stat(exePath); os.IsNotExist(err) {
		if verbose {
			fmt.Printf("  No executable found: %s\n", exePath)
		}
		return false, "", nil
	}

	// Check for resources directory
	resourcesDir := filepath.Join(filepath.Dir(exePath), "resources")
	if _, err := os.Stat(resourcesDir); os.IsNotExist(err) {
		if verbose {
			fmt.Printf("  No resources directory found: %s\n", resourcesDir)
		}
		return false, "", nil
	}

	// Check for app.asar file
	asarPath := filepath.Join(resourcesDir, "app.asar")
	if _, err := os.Stat(asarPath); err == nil {
		if verbose {
			fmt.Printf("  Found app.asar: %s\n", asarPath)
		}

		// Try to extract version from package.json if it exists
		version := "unknown"
		packageJsonPath := filepath.Join(resourcesDir, "app", "package.json")
		if _, err := os.Stat(packageJsonPath); err == nil {
			if verbose {
				fmt.Printf("  Found package.json, checking for Electron version\n")
			}

			packageContent, err := os.ReadFile(packageJsonPath)
			if err == nil {
				// Look for electron in dependencies or devDependencies
				re := regexp.MustCompile(`"electron":\s*"([^"]+)"`)
				matches := re.FindStringSubmatch(string(packageContent))
				if len(matches) > 1 {
					if verbose {
						fmt.Printf("  Found Electron version in package.json: %s\n", matches[1])
					}
					version = matches[1]
				} else {
					// Try to find electronVersion
					re = regexp.MustCompile(`"electronVersion":\s*"([^"]+)"`)
					matches = re.FindStringSubmatch(string(packageContent))
					if len(matches) > 1 {
						if verbose {
							fmt.Printf("  Found electronVersion in package.json: %s\n", matches[1])
						}
						version = matches[1]
					}
				}
			}
		} else {
			// Check if there's version info in the executable
			exeContent, err := os.ReadFile(exePath)
			if err == nil {
				// Look for patterns like Electron/X.Y.Z
				re := regexp.MustCompile(`Electron/([0-9.]+)`)
				matches := re.FindStringSubmatch(string(exeContent))
				if len(matches) > 1 {
					if verbose {
						fmt.Printf("  Found Electron version in executable: %s\n", matches[1])
					}
					version = matches[1]
				} else {
					// Look for other common patterns
					versionPatterns := []string{
						`electron@([0-9.]+)`,
						`"electron": "([^"]+)"`,
						`"electronVersion": "([^"]+)"`,
					}

					for _, pattern := range versionPatterns {
						re := regexp.MustCompile(pattern)
						matches := re.FindStringSubmatch(string(exeContent))
						if len(matches) > 1 {
							if verbose {
								fmt.Printf("  Found Electron version pattern in executable: %s\n", matches[1])
							}
							version = matches[1]
							break
						}
					}
				}
			}
		}

		return true, version, nil
	}

	// Look for electron.asar which is common in Electron apps
	electronAsarPath := filepath.Join(resourcesDir, "electron.asar")
	if _, err := os.Stat(electronAsarPath); err == nil {
		if verbose {
			fmt.Printf("  Found electron.asar: %s\n", electronAsarPath)
		}

		// Try to find version in the electron.asar metadata
		version := "unknown"
		// Check executable for version info
		exeContent, err := os.ReadFile(exePath)
		if err == nil {
			// Look for patterns like Electron/X.Y.Z
			re := regexp.MustCompile(`Electron/([0-9.]+)`)
			matches := re.FindStringSubmatch(string(exeContent))
			if len(matches) > 1 {
				if verbose {
					fmt.Printf("  Found Electron version in executable: %s\n", matches[1])
				}
				version = matches[1]
			}
		}

		return true, version, nil
	}

	return false, "", nil
}

// GetAsarPath returns the path to the app.asar file for an Electron application
func GetAsarPath(appPath string) string {
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(appPath, "Contents", "Resources", "app.asar")
	case "windows":
		exePath := appPath
		if !strings.HasSuffix(exePath, ".exe") {
			exePath = filepath.Join(appPath, filepath.Base(appPath)+".exe")
		}
		return filepath.Join(filepath.Dir(exePath), "resources", "app.asar")
	default:
		return ""
	}
}

// HasAsarFile checks if the app has an app.asar file
func HasAsarFile(appPath string) bool {
	asarPath := GetAsarPath(appPath)
	_, err := os.Stat(asarPath)
	return err == nil
}

// FindNodeFiles finds .node files in an Electron application
// maxFiles specifies the maximum number of files to return (0 for unlimited)
func FindNodeFiles(appPath string, maxFiles int, verbose bool) []string {
	var nodeFiles []string

	// Define the search roots based on the OS
	var searchRoots []string
	switch runtime.GOOS {
	case "darwin":
		// For macOS, search in the main app resources
		searchRoots = []string{
			filepath.Join(appPath, "Contents", "Resources"),
			filepath.Join(appPath, "Contents", "Frameworks"),
		}
	case "windows":
		// For Windows, search in the app directory and resources
		exePath := appPath
		if !strings.HasSuffix(exePath, ".exe") {
			exePath = filepath.Join(appPath, filepath.Base(appPath)+".exe")
		}
		dirPath := filepath.Dir(exePath)
		searchRoots = []string{
			dirPath,
			filepath.Join(dirPath, "resources"),
		}
	default:
		if verbose {
			fmt.Printf("Unsupported OS for .node file search: %s\n", runtime.GOOS)
		}
		return nodeFiles
	}

	// Search each root directory
	for _, root := range searchRoots {
		if verbose {
			fmt.Printf("Searching for .node files in: %s\n", root)
		}

		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				if verbose {
					fmt.Printf("Error accessing path %s: %v\n", path, err)
				}
				return filepath.SkipDir
			}

			// Check if it's a .node file
			if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".node") {
				if verbose {
					fmt.Printf("Found .node file: %s\n", path)
				}
				nodeFiles = append(nodeFiles, path)

				// Check if we've reached the maximum number of files
				if maxFiles > 0 && len(nodeFiles) >= maxFiles {
					return errors.New("max files reached")
				}
			}

			return nil
		})

		if err != nil && verbose {
			fmt.Printf("Error walking the path %s: %v\n", root, err)
		}

		// Stop if we've reached the maximum number of files
		if maxFiles > 0 && len(nodeFiles) >= maxFiles {
			break
		}
	}

	return nodeFiles
}
