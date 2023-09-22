package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	tmpDir, err := os.MkdirTemp("", "sandbox")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpDir)
	test := "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Hello, world\")\n}\n"
	fs, err := splitFiles([]byte(test))
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	if !fs.Contains("go.mod") {
		fs.AddFile("go.mod", []byte("module play\n"))
	}
	for f, src := range fs.m {
		in := filepath.Join(tmpDir, f)
		if err := os.WriteFile(in, src, 0644); err != nil {
			panicError(err)
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

	if err := cmd.Run(); err != nil {
		fmt.Println("error here", out)
		panicError(err)
	}

	fmt.Println("output", out)
}

func panicError(err error) {
	fmt.Println("error here", err)
	panic(err)
}
