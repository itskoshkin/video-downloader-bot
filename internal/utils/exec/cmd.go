package exec

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type ExecutionError struct {
	Binary string
	Args   []string
	Stdout string
	Stderr string
	Err    error
}

func (e *ExecutionError) Error() string {
	command := strings.TrimSpace(strings.Join(append([]string{e.Binary}, e.Args...), " "))
	if e.Stderr != "" {
		return fmt.Sprintf("%s: %v: %s", command, e.Err, strings.TrimSpace(e.Stderr))
	}

	return fmt.Sprintf("%s: %v", command, e.Err)
}

func Run(ctx context.Context, binary string, args ...string) (string, string, error) {
	cmd := exec.CommandContext(ctx, binary, args...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return stdout.String(), stderr.String(), &ExecutionError{
			Binary: binary,
			Args:   args,
			Stdout: stdout.String(),
			Stderr: stderr.String(),
			Err:    err,
		}
	}

	return stdout.String(), stderr.String(), nil
}
