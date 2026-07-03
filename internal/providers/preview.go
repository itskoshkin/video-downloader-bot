package providers

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/viper"

	"video-downloader-bot/internal/config"
	"video-downloader-bot/internal/utils/links"
)

// ---- URL-rewrite preview ----

// previewProvider rewrites the link's host to a third-party embed-fix service so Telegram renders an
// inline video preview. No API key, no download — returns a URL the caller sends as-is. This is the
// last-resort fallback when the file providers can't fetch the media.
//
// The configured domains are a list of INDEPENDENT embed-fix services (vxinstagram, EmbedEZ's
// tiktokez, fxtiktok, ...), not a single provider. The first is used; the "preview not working?"
// button cycles the rest.
type previewProvider struct{}

func NewPreview() Downloader { return &previewProvider{} }

func (*previewProvider) Name() string { return "preview" }

func (*previewProvider) Applicable(link string) bool {
	switch links.Platform(link) {
	case "instagram", "tiktok":
		return true
	default:
		return false
	}
}

func (*previewProvider) Download(_ context.Context, link string) (*Result, error) {
	rewritten, ok := PreviewURL(link, 0)
	if !ok {
		return nil, ErrNotApplicable
	}
	return &Result{Kind: KindURL, URL: rewritten}, nil
}

// PreviewDomains returns the configured embed-fix domains for a platform (primary first).
func PreviewDomains(platform string) []string {
	switch platform {
	case "instagram":
		return viper.GetStringSlice(config.PreviewInstagramDomains)
	case "tiktok":
		return viper.GetStringSlice(config.PreviewTiktokDomains)
	default:
		return nil
	}
}

// PreviewURL rewrites link to the embed-fix domain at index idx (wrapping around the list).
// Returns ok=false when the platform has no configured domains.
func PreviewURL(link string, idx int) (string, bool) {
	domains := PreviewDomains(links.Platform(link))
	if len(domains) == 0 {
		return "", false
	}
	rewritten, err := swapHost(link, domains[wrap(idx, len(domains))])
	if err != nil {
		return "", false
	}
	return rewritten, true
}

// PreviewCycleData builds the callback data for the "preview not working?" button, which cycles to
// the next domain. Returns "" when there is nothing to cycle to or it wouldn't fit Telegram's limit.
func PreviewCycleData(link string, nextIdx int) string {
	domains := PreviewDomains(links.Platform(link))
	if len(domains) <= 1 {
		return ""
	}
	data := fmt.Sprintf("preview:%d:%s", wrap(nextIdx, len(domains)), link)
	if len(data) > 64 { // Telegram callback_data hard limit
		return ""
	}
	return data
}

func wrap(i, n int) int { return ((i % n) + n) % n }

// swapHost replaces the host of link with newHost, keeping path and query intact.
func swapHost(link, newHost string) (string, error) {
	u, err := url.Parse(ensureScheme(link))
	if err != nil {
		return "", err
	}
	u.Host = newHost
	return u.String(), nil
}

func ensureScheme(link string) string {
	if !strings.HasPrefix(link, "http://") && !strings.HasPrefix(link, "https://") {
		return "https://" + link
	}
	return link
}

// Why there is no paid "fetch" preview provider: some of these services (notably EmbedEZ) also sell
// an API that resolves a real downloadable media URL — but EmbedEZ's is priced SEVERAL TIMES HIGHER
// than Hiker's actual media download, absurd for what is essentially fixing a preview link, and it
// has since dropped Instagram entirely. 🤡🤡🤡 So we only do free host-rewrite previews here.
