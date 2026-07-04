package providers

import (
	"context"
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
	"video-downloader-bot/internal/utils/links"
)

// aiograpi is a Go client for the external aiograpi-rest sidecar (the async rewrite of instagrapi-rest;
// a logged-in burner IG account, run outside this repo). It resolves an Instagram reel to a downloaded
// file via the sidecar's `GET /clip/download/by/url` endpoint, authenticated with an X-Session-ID header.
//
// NOTE: written against the aiograpi-rest contract but NOT yet verified end-to-end — the burner
// login / IP situation on the sidecar side is still being sorted. Re-check against a live session.
type aiograpi struct {
	client *http.Client
}

func NewAiograpi() Downloader {
	return &aiograpi{client: &http.Client{Timeout: 2 * time.Minute}}
}

func (*aiograpi) Name() string { return "aiograpi" }

func (*aiograpi) Applicable(link string) bool { return links.Platform(link) == "instagram" }

func (a *aiograpi) Download(ctx context.Context, link string) (*Result, error) {
	base := strings.TrimRight(viper.GetString(config.AiograpiBaseURL), "/")
	session := viper.GetString(config.AiograpiSessionID)
	if base == "" || session == "" {
		return nil, ErrNotImplemented // sidecar not configured — skip this provider
	}

	endpoint := fmt.Sprintf("%s/clip/download/by/url?url=%s&returnFile=true", base, url.QueryEscape(link))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("aiograpi: build request: %w", err)
	}
	req.Header.Set("X-Session-ID", session)
	req.Header.Set("Accept", "application/json")

	logger.DebugWithID(ctx, "aiograpi: requesting clip for \"%s\" from sidecar...", link)
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("aiograpi: request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("aiograpi: sidecar %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	// returnFile=true -> the response body is the video file itself; stream it into the downloads folder.
	out := filepath.Join(viper.GetString(config.TelegramBotVideoDownloadFolder), aiograpiFilename(link))
	f, err := os.Create(out)
	if err != nil {
		return nil, fmt.Errorf("aiograpi: create file: %w", err)
	}
	defer func() { _ = f.Close() }()

	if _, err = io.Copy(f, resp.Body); err != nil {
		_ = os.Remove(out)
		return nil, fmt.Errorf("aiograpi: save file: %w", err)
	}

	logger.DebugWithID(ctx, "aiograpi: downloaded clip to \"%s\".", out)
	return &Result{Kind: KindFile, FilePath: out}, nil
}

// aiograpiFilename derives a unique .mp4 name from the reel shortcode, falling back to a uuid.
func aiograpiFilename(link string) string {
	name := uuid.NewString()
	if u, err := url.Parse(ensureScheme(link)); err == nil {
		if parts := strings.Split(strings.Trim(u.Path, "/"), "/"); len(parts) >= 2 && parts[len(parts)-1] != "" {
			name = parts[len(parts)-1] // e.g. "/reel/ABC123" -> "ABC123"
		}
	}
	return name + ".mp4"
}
