package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

const sandboxBackEndUrl = "localhost:3000/run"

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

func runCode(ctx context.Context, exePath string) error {
	exeBytes, err := os.ReadFile(exePath)
	if err != nil {
		return err
	}

	sreq, err := http.NewRequestWithContext(ctx, "POST", sandboxBackEndUrl, bytes.NewReader(exeBytes))
	if err != nil {
		return fmt.Errorf("NewRequestWithContext %q:%v", sandboxBackEndUrl, err)
	}
	res, err := http.DefaultClient.Do(sreq)
	if err != nil {
		return fmt.Errorf("POST %q: %w", sandboxBackEndUrl, err)
	}

	fmt.Println(res)

	return nil
}
