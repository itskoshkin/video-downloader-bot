package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/viper"

	"video-downloader-bot/internal/config"
	"video-downloader-bot/internal/logger"
	"video-downloader-bot/internal/proxy"
	"video-downloader-bot/internal/utils/links"
)

// hikerAPI resolves an Instagram reel via the hosted HikerAPI private API (covers auth-gated /
// private reels yt-dlp can't fetch). It calls GET /v1/media/by/url (auth: x-access-key header),
// reads the media's direct video_url, and downloads it to a file.
//
// NOTE: written against API-HIKER.md but not yet verified end-to-end (no key on hand).
type hikerAPI struct {
	client *http.Client
}

// hikerBaseURL is the hosted API. Alternative without Cloudflare: https://api.instagrapi.com
const hikerBaseURL = "https://api.hikerapi.com"

func NewHikerAPI() Downloader {
	return &hikerAPI{client: proxy.Client(2 * time.Minute)}
}

func (*hikerAPI) Name() string { return "hikerapi" }

func (*hikerAPI) Applicable(link string) bool { return links.Platform(link) == "instagram" }

func (h *hikerAPI) Download(ctx context.Context, link string) (*Result, error) {
	key := viper.GetString(config.HikerApiKey)
	if key == "" {
		return nil, ErrNotImplemented // no API key configured — skip this provider
	}

	// 1. resolve the media by URL -> read its direct video_url
	infoURL := fmt.Sprintf("%s/v1/media/by/url?url=%s", hikerBaseURL, url.QueryEscape(link))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, infoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("hikerapi: build request: %w", err)
	}
	req.Header.Set("x-access-key", key)
	req.Header.Set("Accept", "application/json")

	logger.DebugWithID(ctx, "hikerapi: resolving media for \"%s\"...", link)
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("hikerapi: request: %w", err)
	}
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("hikerapi: media/by/url %s: %s", resp.Status, strings.TrimSpace(string(body[:min(len(body), 2048)])))
	}

	var media struct {
		VideoURL string `json:"video_url"`
	}
	if err = json.Unmarshal(body, &media); err != nil {
		return nil, fmt.Errorf("hikerapi: decode media: %w", err)
	}
	if media.VideoURL == "" {
		return nil, fmt.Errorf("hikerapi: media has no video_url")
	}

	// 2. download the direct video URL into the downloads folder
	out := filepath.Join(viper.GetString(config.TelegramBotVideoDownloadFolder), hikerFilename(link))
	if err = h.fetchToFile(ctx, media.VideoURL, out); err != nil {
		return nil, fmt.Errorf("hikerapi: download video: %w", err)
	}

	logger.DebugWithID(ctx, "hikerapi: downloaded reel to \"%s\".", out)
	return &Result{Kind: KindFile, FilePath: out}, nil
}

func (h *hikerAPI) fetchToFile(ctx context.Context, src, dst string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, src, nil)
	if err != nil {
		return err
	}
	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("video url returned %s", resp.Status)
	}

	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	if _, err = io.Copy(f, resp.Body); err != nil {
		_ = os.Remove(dst)
		return err
	}
	return nil
}

// hikerFilename derives a unique .mp4 name from the reel shortcode, falling back to a uuid.
func hikerFilename(link string) string {
	name := uuid.NewString()
	if u, err := url.Parse(ensureScheme(link)); err == nil {
		if parts := strings.Split(strings.Trim(u.Path, "/"), "/"); len(parts) >= 2 && parts[len(parts)-1] != "" {
			name = parts[len(parts)-1] // e.g. "/reel/ABC123" -> "ABC123"
		}
	}
	return name + ".mp4"
}
