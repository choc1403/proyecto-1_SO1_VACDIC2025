package utils

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

// baseDir contendrá la ruta absoluta del directorio que contiene 'utils.go'
var baseDir string

func init() {
	// Usamos runtime.Caller(0) para obtener información sobre el archivo donde se llama init() (utils.go)
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		// En caso de fallo, volver a la lógica original (para escenarios extremos, pero preferimos el Caller)
		log.Println("Warning: runtime.Caller failed, falling back to os.Args[0] logic.")
		baseDir = filepath.Dir(os.Args[0])
	} else {
		// path.Dir(file) nos da la ruta absoluta de /path/to/so1-daemon/utils
		baseDir = filepath.Dir(file)
	}
}

func ABSPATH(rel string) string {
	// baseDir: /path/to/so1-daemon/utils
	// rel: ../bash/prueba.sh
	// filepath.Join maneja el '..' correctamente.
	absPath, err := filepath.Abs(filepath.Join(baseDir, rel))
	if err != nil {
		// Manejar el error de forma más elegante si es necesario, por ahora fatal es suficiente
		panic(fmt.Sprintf("Error al generar la ruta absoluta para %s: %v", rel, err))
	}
	return absPath
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
