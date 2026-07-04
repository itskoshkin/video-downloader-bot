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

// authGatedMarkers are yt-dlp/extractor stderr fragments meaning the media sits behind an
// age / login / private gate — unreachable without an authenticated session.
var authGatedMarkers = []string{
	"empty media response",
	"login_required",
	"requested content is not available",
	"restricted video",
	"age-restricted",
	"sign in to confirm your age",
	"log in for access",                // TikTok 18+
	"this post may not be comfortable", // TikTok 18+
}

// IsAuthGated reports whether err comes from age/login/private-gated media that can't be
// fetched without a logged-in session. Callers use it to show a human hint instead of raw stderr.
func IsAuthGated(err error) bool {
	if err == nil {
		return false
	}
	text := strings.ToLower(err.Error()) // ExecutionError.Error() already embeds the stderr
	for _, m := range authGatedMarkers {
		if strings.Contains(text, m) {
			return true
		}
	}
	return false
}
