package main

import (
	"fmt"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/bitrise/bitrise"
)

func generate(manifestPath, outputDir, pluginDataDir string) error {
	err := manifestCheck(manifestPath)
	if err != nil {
		return err
	}

	err = setupRequiredFramework(manifestPath, pluginDataDir)
	if err != nil {
		return err
	}

	frameworkPath := filepath.Join(pluginDataDir, "Framework", "0.0.1", "BitriseDescription.xcframework", "macos-arm64_x86_64")
	frameworkName := "BitriseDesription"
	swiftArgs := []string{
		"swift",
		//"-v",
		//"-suppress-warnings",
		"-F", frameworkPath,
		fmt.Sprintf("-l%s", frameworkName),
		"-framework", frameworkName,
		manifestPath,
		"--dump",
	}

	fmt.Printf("Converting %s to Bitrise config\n", manifestPath)

	cmd := exec.Command("/usr/bin/xcrun", swiftArgs...)
	stdout, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("command failed: %w", err)
	}

	fmt.Println("Conversion successful")

	outputString := strings.TrimSpace(string(stdout))
	jsonString := split(outputString, "BITRISE_MANIFEST_BEGIN", "BITRISE_MANIFEST_END")
	jsonString = strings.TrimSpace(jsonString)
	if jsonString == "" {
		return fmt.Errorf("missing json manifest")
	}

	model, _, err := bitrise.ConfigModelFromJSONBytes([]byte(jsonString))
	if err != nil {
		return fmt.Errorf("cannot create Bitrise model: %w", err)
	}

	bitriseYamlPath := path.Join(outputDir, "bitrise.yml")
	err = bitrise.SaveConfigToFile(bitriseYamlPath, model)
	if err != nil {
		return fmt.Errorf("could not save bitrise.yml: %w", err)
	}

	comment := fmt.Sprintf("# Yaml generated by Bitrise swift plugin %s. DO NOT EDIT\n", version)
	err = addFirstLine(bitriseYamlPath, comment)
	if err != nil {
		return fmt.Errorf("could not edit bitrise.yml: %w", err)
	}

	fmt.Printf("Bitrise config saved at %s\n", bitriseYamlPath)

	return nil
}
