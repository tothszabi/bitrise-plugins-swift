package main

import (
	"fmt"
	"os"
)

var bitriseSwiftTemplate = `// BitriseDescription: 0.0.1

import BitriseDescription

let bitrise = Bitrise(
    formatVersion: .v13,
    projectType: .iOS
)
`

func initialise(manifestPath string) error {
	manifestExists, isDir, err := itemExists(manifestPath)
	if err != nil {
		return err
	}
	if manifestExists && isDir == false {
		return fmt.Errorf("file exists at %s", manifestPath)
	}

	err = os.WriteFile(manifestPath, []byte(bitriseSwiftTemplate), 0755)
	if err != nil {
		return err
	}

	return nil
}
