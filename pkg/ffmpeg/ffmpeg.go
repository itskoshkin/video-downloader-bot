package ffmpeg

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"video-downloader-bot/internal/logger"
	"video-downloader-bot/internal/utils/exec"
)

// noinspection SpellCheckingInspection
const binary = "ffmpeg"

func CheckIfInstalled(ctx context.Context) error {
	_, _, err := exec.Run(ctx, binary, "-version")
	return err
}

func Convert(ctx context.Context, inputPath, outputPath string) (string, error) {
	if strings.TrimSpace(inputPath) == "" {
		return "", fmt.Errorf("%s: convert: empty input path", binary)
	}
	if strings.TrimSpace(outputPath) == "" {
		return "", fmt.Errorf("%s: convert: empty output path", binary)
	}

	args := []string{
		"-y",            // Overwrite the output file if already present
		"-i", inputPath, // Input media file path
		"-c:v", "libx264", // Encode video stream with the H.264 codec
		"-pix_fmt", "yuv420p", // Force the pixel format to yuv420p (should be more compatible format for Telegram mobile clients)
		"-movflags", "+faststart", // Move MP4 metadata to beginning of file so playback can start faster when streaming/downloading
		"-vf", "scale=trunc(iw*if(sar\\,sar\\,1)/2)*2:trunc(ih/2)*2,setsar=1", // Apply video filters: compensate for non-square SAR (default to 1 if undefined), make dimensions even, reset SAR to 1:1
		"-c:a", "aac", // Encode the audio stream with AAC
		"-b:a", "128k", // Set audio bitrate to 128 kbps
		outputPath,
	}

	logger.DebugWithID(ctx, "Converting video \"%s\"...", filepath.Base(inputPath))
	_, stderr, err := exec.Run(ctx, binary, args...)
	if err != nil {
		return "", err
	}
	logger.DebugWithID(ctx, "Converted video \"%s\".", filepath.Base(outputPath))

	if strings.Contains(strings.TrimSpace(stderr), "Error") {
		return "", fmt.Errorf("%s", strings.TrimSpace(stderr))
	}

	return outputPath, nil
}

type ProbeResult struct {
	Streams []VideoDimensions `json:"streams"`
}

type VideoDimensions struct {
	Width  int64 `json:"width"`
	Height int64 `json:"height"`
}

func Probe(ctx context.Context, filePath string) (*VideoDimensions, error) {
	args := []string{
		"-v", "quiet", // Suppress ffprobe logs and print only the requested output
		"-print_format", "json", // Format the probe result as JSON
		"-show_streams",          // Include stream-level metadata in the output
		"-select_streams", "v:0", // Inspect only the first video stream
		filePath,
	}

	stdout, _, err := exec.Run(ctx, "ffprobe", args...)
	if err != nil {
		return nil, fmt.Errorf("ffprobe: %w", err)
	}

	var result ProbeResult
	if err = json.Unmarshal([]byte(stdout), &result); err != nil {
		return nil, fmt.Errorf("ffprobe: error decoding json: %w", err)
	}

	if len(result.Streams) == 0 {
		return nil, fmt.Errorf("ffprobe: no video streams found")
	}

	return &result.Streams[0], nil
}
