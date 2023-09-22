package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
)

var memoryLimitBytes = 100 << 20

func main() {
	startWorker(context.Background())
}

func startWorker(ctx context.Context) error {
	//select {
	//case <-ctx.Done():
	//	return nil
	//}

	name := "container" + randomHex(8)

	cmd := exec.Command("docker", "run",
		"gcr.io/golang-org/playground-sandbox-gvisor:latest",
		"--name=", name,
		"--rm",
		"-i",
		"--memory="+fmt.Sprint(memoryLimitBytes),
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		log.Println(err)
		os.Stderr.Write(stderr.Bytes())
	}

	return nil
}

type Container struct {
	name string

	stdin  io.WriteCloser
	stdout *io.Writer
	stderr *io.Writer

	cmd *exec.Cmd
}

func randomHex(n int) string {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)
}
