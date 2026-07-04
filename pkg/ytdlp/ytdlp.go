package ytdlp

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"

	"video-downloader-bot/internal/config"
	"video-downloader-bot/internal/logger"
	"video-downloader-bot/internal/utils/exec"
)

func binary() string { return viper.GetString(config.YtDlpBinary) }

type DownloadResult struct {
	Filepath string
	Meta     Metadata
	Stdout   string
	Stderr   string
}

type Metadata struct {
	ID             string  `json:"id"`
	Title          string  `json:"title"`
	Author         string  `json:"uploader"`
	AuthorID       string  `json:"uploader_id"`
	Description    string  `json:"description"`
	Thumbnail      string  `json:"thumbnail"`
	Extension      string  `json:"ext"`
	Duration       float64 `json:"duration"`
	Filesize       int64   `json:"filesize"`
	FilesizeApprox int64   `json:"filesize_approx"`
	WebpageURL     string  `json:"webpage_url"`
	OriginalURL    string  `json:"original_url"`
	Extractor      string  `json:"extractor"`
}

func CheckIfInstalled(ctx context.Context) error {
	_, _, err := exec.Run(ctx, binary(), "--version")
	return err
}

func FetchMetadata(ctx context.Context, link string) (*Metadata, error) {
	if strings.TrimSpace(link) == "" {
		return nil, fmt.Errorf("yt-dlp: fetch metadata: empty url")
	}

	args := []string{
		"-J",            // Print full metadata instead of downloading video
		"--no-playlist", // Treat link as single media item
		"--no-warnings", // Suppress non-fatal errors and warnings
		"--no-update",   // Never check for updates (silences the periodic "version is out of date" warning); updates are handled via pip only
	}
	if viper.GetBool(config.YtDlpUseCookies) {
		args = append(args, "--cookies", viper.GetString(config.YtDlpCookiesFile)) // Load cookies from a file and use them for authenticated requests, needed when anonymous access is rate-limited or blocked because of age restrictions
	}
	if p := strings.TrimSpace(viper.GetString(config.ProxySocks5)); p != "" {
		args = append(args, "--proxy", p) // Route through the SOCKS5 proxy (byedpi/tunnel) for DPI bypass
	}
	args = append(args, link)

	logger.DebugWithID(ctx, "Fetching video metadata from \"%s\"...", link)
	stdout, _, err := exec.Run(ctx, binary(), args...)
	if err != nil {
		return nil, err
	}
	logger.DebugWithID(ctx, "Fetched metadata for \"%s\".", link)

	var metadata Metadata
	if err = json.Unmarshal([]byte(stdout), &metadata); err != nil {
		return nil, fmt.Errorf("yt-dlp: fetch metadata: error decoding json: %w", err)
	}

	return &metadata, nil
}

func DownloadVideo(ctx context.Context, link string) (string, error) {
	if strings.TrimSpace(link) == "" {
		return "", fmt.Errorf("yt-dlp: download videos: empty url")
	}

	args := []string{
		"--print", "after_move:filepath", // Print the final file path after the file has been fully downloaded and moved into place
		"--newline",      // Print progress and log output line by line instead of updating one terminal line in place
		"-f", "bv*+ba/b", // Choose format: best available video and available audio or fallback to best single file if separate video/audio is not available
		"--merge-output-format", "mp4", // f video and audio are downloaded separately, merge them into an MP4 container
		"-o", viper.GetString(config.TelegramBotVideoDownloadFolder) + "%(id)s.%(ext)s", // Output file name template ("%(id)s" is the media ID and "%(ext)s" is the resulting file extension)
		"--no-playlist",                                                                     // Download only the single media item, not the whole playlist/thread/collection
		"--no-update",                                                                       // Never check for updates (silences the periodic "version is out of date" warning); updates are handled via pip only
		"--max-filesize", fmt.Sprintf("%dM", viper.GetInt(config.TelegramBotMaxFileSizeMB)), // Skip download if filesize exceeds limit
		"--match-filter", fmt.Sprintf("duration<=?%d", viper.GetInt(config.TelegramBotMaxVideoDuration)), // Skip download if duration exceeds limit (? = skip check if duration is unknown)
	}
	if viper.GetBool(config.YtDlpUseCookies) {
		args = append(args, "--cookies", viper.GetString(config.YtDlpCookiesFile)) // Load cookies from a file and use them for authenticated requests, needed when anonymous access is rate-limited or blocked because of age restrictions
	}
	if p := strings.TrimSpace(viper.GetString(config.ProxySocks5)); p != "" {
		args = append(args, "--proxy", p) // Route through the SOCKS5 proxy (byedpi/tunnel) for DPI bypass
	}
	args = append(args, link)

	logger.DebugWithID(ctx, "Downloading video from \"%s\"...", link)
	stdout, stderr, err := exec.Run(ctx, binary(), args...)
	if err != nil {
		return "", err
	}
	logger.DebugWithID(ctx, "Finished downloading video from \"%s\".", link)

	if strings.TrimSpace(stdout) == "" { // No file was downloaded — video was skipped by --max-filesize or --match-filter
		if strings.TrimSpace(stderr) != "" {
			return "", fmt.Errorf("yt-dlp: video skipped: %s", strings.TrimSpace(stderr))
		}
		return "", fmt.Errorf("yt-dlp: no file downloaded (video may exceed size or duration limit)")
	}

	switch {
	case stderr == "":
		return filepath.Base(strings.TrimSpace(stdout)), nil
	case strings.HasPrefix(strings.TrimSpace(stderr), "WARNING"):
		logger.WarnWithID(ctx, "yt-dlp: %s", strings.TrimSpace(stderr))
		return filepath.Base(strings.TrimSpace(stdout)), nil
	default:
		return "", fmt.Errorf("%s", strings.TrimSpace(stderr))
	}
}
