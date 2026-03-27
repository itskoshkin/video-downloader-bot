package errs

import (
	"errors"
	"fmt"
	"strings"

	"video-downloader-bot/internal/utils/exec"
	"video-downloader-bot/internal/utils/str"
)

func ShortStderr(err error) string {
	var execErr *exec.ExecutionError
	if !errors.As(err, &execErr) || execErr.Stderr == "" {
		return ""
	}

	stderr := strings.TrimSpace(execErr.Stderr)

	// yt-dlp usually puts the useful message on the last ERROR: line
	for _, line := range str.ReverseLines(stderr) {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "ERROR:") {
			return strings.TrimPrefix(line, "ERROR: ")
		}
	}

	lines := strings.Split(stderr, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if l := strings.TrimSpace(lines[i]); l != "" {
			return l
		}
	}

	return ""
}

func FullStderr(err error) string {
	var execErr *exec.ExecutionError
	if !errors.As(err, &execErr) || execErr.Stderr == "" {
		return ""
	}
	return strings.TrimSpace(execErr.Stderr)
}

func ExecBinary(err error) string {
	var execErr *exec.ExecutionError
	if !errors.As(err, &execErr) {
		return ""
	}
	return execErr.Binary
}

// ShortError returns the error message with full stderr replaced by its last non-empty line.
// For non-exec errors it returns err.Error() unchanged.
func ShortError(err error) string {
	var execErr *exec.ExecutionError
	if !errors.As(err, &execErr) {
		return err.Error()
	}
	short := ShortStderr(err)
	command := strings.TrimSpace(strings.Join(append([]string{execErr.Binary}, execErr.Args...), " "))
	if short != "" {
		return fmt.Sprintf("%s: %v: %s", command, execErr.Err, short)
	}
	return fmt.Sprintf("%s: %v", command, execErr.Err)
}
