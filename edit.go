package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	gotty "github.com/mattn/go-tty"
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
	fmt.Println("Press any key to save your manifest file and press CTRL + C once you are finished editing")
	fmt.Println()

	tty, err := gotty.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer func(tty *gotty.TTY) {
		err := tty.Close()
		if err != nil {
			// print error
		}
	}(tty)

	for {
		_, err := tty.ReadRune()
		if err != nil {
			log.Fatal(err)
		}

		err = copyItem(manifestPackagePath, manifestPath)
		if err != nil {
			return err
		}

		fmt.Printf("Saved at %s\n", time.Now().Format(time.RFC850))
	}

	return nil
}
