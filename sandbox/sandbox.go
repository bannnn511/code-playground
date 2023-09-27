package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
)

const memoryLimitBytes = 100 << 20
const maxBinarySize = 100 << 20

var (
	readyContainer chan *Container
)

func main() {
	go startWorker(context.TODO())

	mux := http.NewServeMux()
	mux.HandleFunc("/health", health)
	mux.HandleFunc("/run", runHandler)

	httpServer := http.Server{
		Addr:      ":3000",
		Handler:   mux,
		TLSConfig: nil,
	}
	log.Fatal(httpServer.ListenAndServe())
}

func health(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte("OK"))
}

func runHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "expected POST method", http.StatusBadRequest)
		return
	}

	bin, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBinarySize))
	if err != nil {
		log.Printf("failed to read request body: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	c, err := getContainer(context.Background())
	if err != nil {
		if cerr := r.Context().Err(); cerr != nil {
			log.Printf("getContainer, client side cancellation: %v", cerr)
			return
		}
		http.Error(w, "failed to get container", http.StatusInternalServerError)
		log.Printf("failed to get container: %v", err)
		return
	}

	if _, err := c.stdin.Write(bin); err != nil {
		log.Printf("failed write to container %v", err)
		http.Error(w, "unknown error during docker run", http.StatusInternalServerError)
		return
	}

	c.stdin.Close()
	log.Println("wrote+closed")
	err = c.cmd.Wait()

	res := &Response{}
	if err != nil {
		var ee *exec.ExitError
		if !errors.As(err, &ee) {
			http.Error(w, "unknown error during docker run", http.StatusBadRequest)
		}
		res.ExitCode = ee.ExitCode()
	}

	sendResponse(w, res)
}

func startWorker(ctx context.Context) (c *Container, err error) {
	//select {
	//case <-ctx.Done():
	//	return nil
	//}

	name := "container_" + randomHex(8)

	cmd := exec.Command("docker", "run",
		"gcr.io/golang-org/playground-sandbox-gvisor:latest",
		"--name=", name,
		"--rm",
		"-i",
		"--memory="+fmt.Sprint(memoryLimitBytes),
	)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	out := &bytes.Buffer{}
	cmd.Stderr, cmd.Stdout = out, out

	c = &Container{
		name:   name,
		stdin:  stdin,
		stdout: &cmd.Stdout,
		stderr: &cmd.Stderr,
		cmd:    cmd,
	}

	if err := cmd.Start(); err != nil {
		if ee := (*exec.ExitError)(nil); !errors.As(err, &ee) {
			log.Printf("failed to start container: %v %v", err, out.Bytes())
			return nil, fmt.Errorf("failed to start container: %v", err)
		}

		log.Printf("failed to start container: %v %v", err, out.Bytes())
	}

	log.Printf("Started container %q", name)
	readyContainer <- c

	return c, err
}

type Container struct {
	name string

	stdin  io.WriteCloser
	stdout *io.Writer
	stderr *io.Writer

	cmd *exec.Cmd

	errorMsg string
}

func getContainer(ctx context.Context) (*Container, error) {
	select {
	case c := <-readyContainer:
		return c, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func randomHex(n int) string {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)
}

type Response struct {
	Error    string `json:"error,omitempty"`
	ExitCode int    `json:"exitCode"`
	Stdout   []byte `json:"stdout"`
	Stderr   []byte `json:"stderr"`
}

func sendError(w http.ResponseWriter, errMsg string) {
	sendResponse(w, &Response{Error: errMsg})
}

func sendResponse(w http.ResponseWriter, r *Response) {
	jres, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		http.Error(w, "error encoding JSON", http.StatusInternalServerError)
		log.Printf("json marshal: %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", fmt.Sprint(len(jres)))
	w.Write(jres)
}
