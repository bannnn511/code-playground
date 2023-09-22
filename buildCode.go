package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// buildCode compiles code from src and return execute path
func buildCode(tmpDir string, src []byte) (string, error) {
	fs, err := splitFiles(src)
	if err != nil {
		return "", err
	}

	if !fs.Contains("go.mod") {
		fs.AddFile("go.mod", []byte("module play\n"))
	}
	for f, src := range fs.m {
		in := filepath.Join(tmpDir, f)
		if err := os.WriteFile(in, src, 0644); err != nil {
			return "", fmt.Errorf("error creating temp file %q: %v", f, err)
		}
	}

	exePath := filepath.Join(tmpDir, "a.out")

	var goArgs []string
	goArgs = append(goArgs, "build")
	goArgs = append(goArgs, "-o", exePath, ".")
	cmd := exec.Command("go", goArgs...)
	cmd.Dir = tmpDir
	cmd.Env = append(cmd.Env, "CGO_ENABLED=0")

	out := &bytes.Buffer{}
	cmd.Stderr, cmd.Stdout = out, out

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("error starting go build: %v", err)
	}

	return exePath, nil
}
