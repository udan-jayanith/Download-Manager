package main

import (
	"os"
	"strings"
)

// This only support ${homeDir} because of that this replaces '${homeDir}' with home dir.
func Dir(dir string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(dir, "${homeDir}", homeDir), nil
}
