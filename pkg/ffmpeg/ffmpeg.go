package ffmpeg

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
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

// convertSlot allows one conversion at a time — parallel x264 encodes on a small VPS starve the CPU and get OOM-killed
var convertSlot = make(chan struct{}, 1)

// maxLongSide caps the re-encoded resolution — Telegram doesn't need 1440p/4K, and x264 memory grows with the frame size
const maxLongSide = 1920

func Convert(ctx context.Context, inputPath, outputPath string) (string, error) {
	if strings.TrimSpace(inputPath) == "" {
		return "", fmt.Errorf("%s: convert: empty input path", binary)
	}
	if strings.TrimSpace(outputPath) == "" {
		return "", fmt.Errorf("%s: convert: empty output path", binary)
	}

	select {
	case convertSlot <- struct{}{}:
		defer func() { <-convertSlot }()
	case <-ctx.Done():
		return "", fmt.Errorf("%s: convert: waiting for a free slot: %w", binary, ctx.Err())
	}

	args := []string{
		"-y",                 // Overwrite the output file if already present
		"-hide_banner",       // Skip the build/config banner
		"-nostats",           // No progress lines in stderr, so an error message is the error, not a wall of "frame=..."
		"-loglevel", "error", // Print only real errors
		"-i", inputPath, // Input media file path
	}
	if remuxable(ctx, inputPath) {
		// Already H.264/AAC in yuv420p with square pixels — copy the streams as-is, no re-encode
		args = append(args,
			"-c", "copy", // Copy audio and video streams without re-encoding
			"-movflags", "+faststart", // Move MP4 metadata to beginning of file so playback can start faster when streaming/downloading
			outputPath,
		)
	} else {
		args = append(args,
			"-c:v", "libx264", // Encode video stream with the H.264 codec
			"-preset", "veryfast", // Several times faster and lighter on memory than the default "medium", at a slightly bigger file
			"-pix_fmt", "yuv420p", // Force the pixel format to yuv420p (should be more compatible format for Telegram mobile clients)
			"-movflags", "+faststart", // Move MP4 metadata to beginning of file so playback can start faster when streaming/downloading
			"-vf", "scale=trunc(iw*if(sar\\,sar\\,1)/2)*2:trunc(ih/2)*2,setsar=1,"+ // Compensate for non-square SAR (default to 1 if undefined), make dimensions even, reset SAR to 1:1
				fmt.Sprintf("scale=trunc(iw*min(1\\,%[1]d/max(iw\\,ih))/2)*2:trunc(ih*min(1\\,%[1]d/max(iw\\,ih))/2)*2", maxLongSide), // Downscale so the long side fits maxLongSide, keeping dimensions even
			"-c:a", "aac", // Encode the audio stream with AAC
			"-b:a", "128k", // Set audio bitrate to 128 kbps
			outputPath,
		)
	}

	logger.DebugWithID(ctx, "Converting video \"%s\"...", filepath.Base(inputPath))
	_, stderr, err := exec.Run(ctx, binary, args...)
	if err != nil {
		_ = os.Remove(outputPath) // Don't leave a truncated file behind
		return "", err
	}
	logger.DebugWithID(ctx, "Converted video \"%s\".", filepath.Base(outputPath))

	if strings.Contains(strings.TrimSpace(stderr), "Error") {
		_ = os.Remove(outputPath)
		return "", fmt.Errorf("%s", strings.TrimSpace(stderr))
	}

	return outputPath, nil
}

type streamInfo struct {
	CodecType         string `json:"codec_type"`
	CodecName         string `json:"codec_name"`
	PixFmt            string `json:"pix_fmt"`
	SampleAspectRatio string `json:"sample_aspect_ratio"`
}

// remuxable reports whether the input already fits Telegram (H.264 yuv420p video with square pixels, AAC or no audio), so a stream copy is enough
func remuxable(ctx context.Context, inputPath string) bool {
	stdout, _, err := exec.Run(ctx, "ffprobe",
		"-v", "quiet", // Suppress ffprobe logs and print only the requested output
		"-print_format", "json", // Format the probe result as JSON
		"-show_entries", "stream=codec_type,codec_name,pix_fmt,sample_aspect_ratio", // Only the fields needed to decide
		inputPath,
	)
	if err != nil {
		return false
	}
	var probe struct {
		Streams []streamInfo `json:"streams"`
	}
	if json.Unmarshal([]byte(stdout), &probe) != nil {
		return false
	}

	videos := 0
	for _, st := range probe.Streams {
		switch st.CodecType {
		case "video":
			videos++
			square := st.SampleAspectRatio == "" || st.SampleAspectRatio == "1:1" || st.SampleAspectRatio == "0:1" || st.SampleAspectRatio == "N/A"
			if st.CodecName != "h264" || st.PixFmt != "yuv420p" || !square {
				return false
			}
		case "audio":
			if st.CodecName != "aac" {
				return false
			}
		}
	}
	return videos == 1
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
