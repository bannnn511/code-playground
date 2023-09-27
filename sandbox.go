package main

import (
	"context"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"os"
)

type compileAndRunRequest struct {
	Code string `json:"code"`
}

func compileAndRun(c echo.Context) error {
	tmpDir, err := os.MkdirTemp("", "sandbox")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpDir)

	req := new(compileAndRunRequest)
	if err := c.Bind(req); err != nil {
		return err
	}

	out, err := buildCode(tmpDir, []byte(req.Code))
	log.Printf("buildCode at %v %v", tmpDir, out)
	if err != nil {
		panicError(err)
	}

	err = runCode(context.TODO(), out)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, out)
}
