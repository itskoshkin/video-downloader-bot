package providers

import (
	"context"
	"errors"
)

// Kind tells the caller how to deliver a successful Result.
type Kind int

const (
	// KindFile: FilePath points to a downloaded local file that still needs
	// converting and uploading (yt-dlp, hikerapi, instagrapi).
	KindFile Kind = iota
	// KindURL: URL is a ready link that Telegram renders as an inline preview;
	// nothing is downloaded or uploaded on our side (URL-rewrite preview).
	KindURL
)

// Result is what a successful Downloader returns.
type Result struct {
	Kind     Kind
	FilePath string // set when Kind == KindFile
	URL      string // set when Kind == KindURL
}

// Sentinel errors let the chain decide whether to fall through to the next provider.
var (
	// ErrNotApplicable: provider doesn't handle this link — skipped silently.
	ErrNotApplicable = errors.New("provider: not applicable")
	// ErrNotImplemented: provider is wired into the chain but not built yet (stub).
	ErrNotImplemented = errors.New("provider: not implemented")
)

// Downloader resolves a link into a media file or a preview URL.
type Downloader interface {
	// Name is the identifier referenced in the YAML chain (e.g. "yt-dlp", "preview").
	Name() string
	// Applicable reports whether this provider can handle the given link.
	Applicable(link string) bool
	// Download returns a Result, or a (preferably sentinel) error so the chain can fall through.
	Download(ctx context.Context, link string) (*Result, error)
}
