package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

var packageTemplate = `// swift-tools-version: 5.9
// The swift-tools-version declares the minimum version of Swift required to build this package.

import PackageDescription

let package = Package(
    name: "Bitrise",
    targets: [
        .target(
            name: "Bitrise", 
            dependencies: ["BitriseDescription"]
        ),
        .binaryTarget(
            name: "BitriseDescription",
            path: "Framework/BitriseDescription.xcframework"
        )
    ]
)
`

func edit(manifestPath, currentDir, pluginDataDir string) error {
	err := manifestCheck(manifestPath)
	if err != nil {
		return err
	}

	err = setupRequiredFramework(manifestPath, pluginDataDir)
	if err != nil {
		return err
	}

	workDir := filepath.Join(currentDir, ".bitrise")
	exists, isDir, err := itemExists(workDir)
	if err != nil {
		return err
	}

	if exists && isDir {
		err := os.RemoveAll(workDir)
		if err != nil {
			return err
		}
	}

	err = os.MkdirAll(workDir, 0755)
	if err != nil {
		return err
	}

	packagePath := filepath.Join(workDir, "Package.swift")
	err = os.WriteFile(packagePath, []byte(packageTemplate), 0755)
	if err != nil {
		return err
	}

	frameworkDirPath := filepath.Join(workDir, "Framework")
	err = os.MkdirAll(frameworkDirPath, 0755)
	if err != nil {
		return err
	}

	sourceDirPath := filepath.Join(workDir, "Sources", "Bitrise")
	err = os.MkdirAll(sourceDirPath, 0755)
	if err != nil {
		return err
	}

	manifestName := filepath.Base(manifestPath)
	manifestPackagePath := filepath.Join(sourceDirPath, manifestName)
	err = copyItem(manifestPath, manifestPackagePath)
	if err != nil {
		return err
	}

	originalFrameworkPath := filepath.Join(pluginDataDir, "Framework", "0.0.1", "BitriseDescription.xcframework")
	frameworkSymlinkPath := filepath.Join(frameworkDirPath, "BitriseDescription.xcframework")
	err = os.Symlink(originalFrameworkPath, frameworkSymlinkPath)
	if err != nil {
		return err
	}

	fmt.Println("Opening your Bitrise.swift file in Xcode")

	cmd := exec.Command("/usr/bin/open", packagePath)
	_, err = cmd.Output()
	if err != nil {
		return fmt.Errorf("could not open %s: %w", packagePath, err)
	}

	fmt.Println()
	fmt.Println("Keep this command running to automatically save your manifest (you can quit it by pressing CTRL + C)")
	fmt.Println()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("file watcher failed to init: %w", err)
	}
	defer func(watcher *fsnotify.Watcher) {
		err := watcher.Close()
		if err != nil {
			fmt.Printf("Failed to close file watcher: %s", err)
		}
	}(watcher)

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				// Apparently, by looking at the fs events, Xcode removes then recreates the file instead writing to it.
				if event.Op == fsnotify.Write || event.Op == fsnotify.Create {
					fmt.Printf("Saving at %s\n", time.Now().Format(time.RFC850))

					err = copyItem(manifestPackagePath, manifestPath)
					if err != nil {
						fmt.Printf("Could not save manifest: %s\n", err)
					}
				}

			case err := <-watcher.Errors:
				fmt.Printf("File watcher error: %s\n", err)
			}
		}
	}()

	if err := watcher.Add(manifestPackagePath); err != nil {
		return fmt.Errorf("file (%s) watching failure: %w", manifestPath, err)
	}

	// This is needed to block the execution until the cli quits.
	done := make(chan bool)
	<-done

	return nil
}
