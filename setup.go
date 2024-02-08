package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func setupRequiredFramework(manifestPath, pluginDataDir string) error {
	line, err := readFirstLine(manifestPath)
	if err != nil {
		return err
	}

	components := strings.Split(line, ":")
	if len(components) != 2 {
		return fmt.Errorf("missing version specification at beginning of the manifest")
	}

	version := strings.TrimSpace(components[1])
	installed, err := isInstalled(version, pluginDataDir)
	if err != nil {
		return err
	}
	if installed {
		return nil
	}

	fmt.Printf("Installing BitriseDescription version %s\n", version)

	return install(version, pluginDataDir)
}

func isInstalled(version, pluginDataDir string) (bool, error) {
	frameworkPath := filepath.Join(pluginDataDir, "Framework", version, "BitriseDescription.xcframework")

	exists, isDir, err := itemExists(frameworkPath)
	if err != nil {
		return false, err
	}

	return exists && isDir, nil
}

func install(version, pluginDataDir string) error {
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		log.Fatal(err)
	}
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			fmt.Printf("Failed to remove temp dir: %s", err)
		}
	}(dir)

	path := filepath.Join(dir, "framework.zip")
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			fmt.Printf("Failed to close file: %s", err)
		}
	}(out)

	fmt.Println("Downloading framework from Github")

	url := fmt.Sprintf("https://github.com/tothszabi/BitriseDescription/releases/download/%s/BitriseDescription.xcframework.zip", version)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("Failed to close response body: %s", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad request status: %s", resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	fmt.Println("Extracting framework")

	frameworkPath := filepath.Join(pluginDataDir, "Framework", version)
	err = os.MkdirAll(frameworkPath, os.ModePerm)
	if err != nil {
		return err
	}

	cmd := exec.Command("/usr/bin/ditto", "-xk", path, frameworkPath)
	_, err = cmd.Output()

	if err != nil {
		return fmt.Errorf("command failed: %w", err)
	}

	fmt.Printf("BitriseDescription version %s is installed\n", version)

	return nil
}
