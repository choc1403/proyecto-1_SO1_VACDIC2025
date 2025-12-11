package utils

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func ABSPATH(rel string) string {
	base, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	return filepath.Join(base, rel)
}

func RunCommand(name string, args ...string) (string, error) {
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

var TrailingCommaRe = regexp.MustCompile(`,\s*([\]\}])`)

func SanitizeJSON(b []byte) []byte {
	return TrailingCommaRe.ReplaceAll(b, []byte("$1"))
}

func ReadProcFile(path string) ([]byte, error) {
	f, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	defer f.Close()

	data, err := ioutil.ReadAll(io.LimitReader(f, 10<<20)) //10MB cap

	if err != nil {
		return nil, err
	}

	return data, nil
}

func ParseMemPct(s string) (float64, error) {
	s = strings.TrimSpace(s)

	if s == "" {
		return 0, nil
	}

	return strconv.ParseFloat(s, 64)
}
