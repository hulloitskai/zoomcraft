// Command exec is a toolchain wrapper used to build and run Joya Travel
// programs by controlling the linker flags passed to 'go build' and 'go run'.
//
// For example, to build the program 'api', you can run (from the repo root):
//   ./tool/exec.sh run ./cmd/api
package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/gobuffalo/here"
)

func main() {
	if err := func() error {
		// Derive "build package" name.
		pkg, err := buildPackage()
		if err != nil {
			return errors.Wrap(err, "locate build package")
		}
		varflag := func(name string, value string) string {
			return fmt.Sprintf("-X '%s.%s=%s'", pkg, name, value)
		}

		// Set linker flags:
		flags := []string{
			varflag("version", gitVersion()),
			varflag("timestamp", time.Now().Format(time.RFC3339)),
		}

		// Configure toolchain arguments.
		args := os.Args[1:]
		if len(args) > 0 {
			args = []string{os.Args[1], "-ldflags", strings.Join(flags, " ")}
			args = append(args, os.Args[2:]...)
		}

		// Configure and run command.
		cmd := exec.Command("go", args...)
		cmd.Env = os.Environ()
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if os.Getenv("GOENV") == "production" {
			cmd.Env = append(cmd.Env, "GOOS=linux", "GOARCH=amd64")
		}

		return cmd.Run()
	}(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %+v\n", err)
		os.Exit(1)
	}
}

const buildPath = "internal/build"

func buildPackage() (string, error) {
	info, err := here.Dir(".")
	if err != nil {
		return "", err
	}
	return path.Join(info.Module.Path, buildPath), nil
}

func gitVersion() string {
	out, err := exec.
		Command("git", "describe", "--always", "--tags", "--dirty").
		Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read version: %+v\n", err)
		return "(dev)"
	}
	return string(bytes.TrimSpace(out))
}

func isTruthy(s string) bool {
	s = strings.TrimSpace(strings.ToLower(s))
	switch s {
	case "1", "true", "t", "yes":
		return true
	default:
		return false
	}
}
