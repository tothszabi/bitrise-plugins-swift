package main

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"slices"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}

func run() error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("only macOS is supported")
	}

	command := os.Args[1]
	acceptedCommands := []string{
		"init",
		"generate",
		"edit",
	}
	if slices.Contains(acceptedCommands, command) == false {
		return fmt.Errorf("unsupported command (%s)", command)
	}

	currentDir := os.Getenv("PWD")
	dataDir := os.Getenv("BITRISE_PLUGIN_INPUT_DATA_DIR")
	manifestPath := path.Join(currentDir, "Bitrise.swift")

	if command == "init" {
		return initialise(manifestPath)
	} else if command == "generate" {
		return generate(manifestPath, currentDir, dataDir)
	} else if command == "edit" {
		return edit(manifestPath, currentDir, dataDir)
	}

	return fmt.Errorf("unexpected execution flow for command (%s)", command)
}
