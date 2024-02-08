package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

func split(output, prefix, suffix string) string {
	sa := strings.SplitN(output, prefix, 2)
	if len(sa) == 1 {
		return ""
	}
	sa = strings.SplitN(sa[1], suffix, 2)
	if len(sa) == 1 {
		return ""
	}
	return sa[0]
}

func itemExists(name string) (bool, bool, error) {
	info, err := os.Stat(name)
	if err == nil {
		return true, info.IsDir(), nil
	}

	if errors.Is(err, os.ErrNotExist) {
		return false, false, nil
	}

	return false, false, err
}

func copyItem(src string, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	err = os.WriteFile(dst, data, 0755)
	if err != nil {
		return err
	}

	return nil
}

func manifestCheck(manifestPath string) error {
	manifestExists, isDir, err := itemExists(manifestPath)
	if err != nil {
		return err
	}
	if manifestExists == false {
		return fmt.Errorf("manifest file does not exist, run init first")
	}
	if isDir {
		return fmt.Errorf("manifest file is a directory")
	}

	return nil
}

func readFirstLine(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			// Log error
		}
	}(file)

	reader := bufio.NewReader(file)
	result, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return result, nil
}

func addFirstLine(path, str string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Printf("File closing error: %s", err)
		}
	}(file)

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	fileContent := str + "\n"
	for _, line := range lines {
		fileContent += line + "\n"
	}

	return os.WriteFile(path, []byte(fileContent), 0755)
}
