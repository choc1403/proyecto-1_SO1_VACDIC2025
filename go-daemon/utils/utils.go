package utils

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func absPath(rel string) string {
	base, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	return filepath.Join(base, rel)
}

func runCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()

	if err != nil {
		return out.String(), fmt.Errorf("cmd %s %v failed: %v - %s", name, args, err, stderr.String())
	}

	return out.String(), nil
}
