package main

import (
	"github.com/labstack/echo/v4"
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
	if err != nil {
		panicError(err)
	}

	return c.JSON(http.StatusOK, out)
}
