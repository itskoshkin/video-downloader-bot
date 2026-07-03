package providers

import (
	"context"

	"video-downloader-bot/internal/utils/links"
)

// hikerAPI is a hosted private Instagram API for auth-gated / 18+ / private reels that
// yt-dlp can't fetch without cookies. Key comes from config.
//
// STUB: blocked on the API contract (response format — direct media URL or file).
// Returns ErrNotImplemented so the chain falls through until it's built.
type hikerAPI struct{}

func NewHikerAPI() Downloader { return &hikerAPI{} }

func (*hikerAPI) Name() string { return "hikerapi" }

func (*hikerAPI) Applicable(link string) bool { return links.Platform(link) == "instagram" }

func (*hikerAPI) Download(_ context.Context, _ string) (*Result, error) {
	return nil, ErrNotImplemented
}
