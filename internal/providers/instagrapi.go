package providers

import (
	"context"

	"video-downloader-bot/internal/utils/links"
)

// instagrapi is a Go client for the external instagrapi-rest sidecar (a logged-in,
// warmed-up burner session). The sidecar runs outside this repo; base URL comes from config.
//
// STUB: blocked on the sidecar HTTP client. Returns ErrNotImplemented so the chain
// falls through until it's built.
type instagrapi struct{}

func NewInstagrapi() Downloader { return &instagrapi{} }

func (*instagrapi) Name() string { return "instagrapi" }

func (*instagrapi) Applicable(link string) bool { return links.Platform(link) == "instagram" }

func (*instagrapi) Download(_ context.Context, _ string) (*Result, error) {
	return nil, ErrNotImplemented
}
