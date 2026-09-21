package providers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/viper"

	"video-downloader-bot/internal/config"
	"video-downloader-bot/internal/proxy"
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

func (*previewProvider) Download(ctx context.Context, link string) (*Result, error) {
	domains := PreviewDomains(links.Platform(link))
	if len(domains) == 0 {
		return nil, ErrNotApplicable
	}
	// Probe the domains for one that actually yields a video embed; the fetch also warms lazy
	// services (InstaFix answers "Post not found" until it has fetched the reel). Falls back to
	// index 0 when nothing validates — Telegram may still render it, and the "preview not
	// working?" button lets the user cycle the rest by hand.
	idx, kind := probeWorkingDomain(ctx, link, domains)
	rewritten, err := swapHost(link, domains[idx])
	if err != nil {
		return nil, ErrNotApplicable
	}
	return &Result{Kind: KindURL, URL: rewritten, Index: idx, PreviewPhoto: kind == embedImage}, nil
}

// telegramPreviewUA mimics Telegram's link-preview fetcher, so embed services return the same
// og: markup Telegram will see (several of them serve og-tags only to known preview bots).
const telegramPreviewUA = "TelegramBot (like TwitterBot)"

// embedKind classifies what a candidate embed page can render.
type embedKind int

const (
	embedNone  embedKind = iota // no usable og media (error page, "Post not found", empty)
	embedImage                  // og:image only — a photo/carousel post, or a video with no playable embed
	embedVideo                  // og:video — a real inline video preview
)

// probeWorkingDomain fetches each candidate embed URL and returns the index of the best one: the
// first domain exposing og:video, or failing that the first with og:image (photo posts, or when
// the video isn't embeddable). The fetch doubles as a warm-up for lazy services, so a second pass
// (after a short delay) catches domains still fetching on the first. Returns 0 when probing is
// disabled or nothing usable is found within the budget.
func probeWorkingDomain(ctx context.Context, link string, domains []string) (int, embedKind) {
	if !viper.GetBool(config.PreviewProbeEnabled) {
		return 0, embedNone
	}
	perFetch := viper.GetDuration(config.PreviewProbeTimeout)
	if perFetch <= 0 {
		perFetch = 5 * time.Second
	}
	budget := viper.GetDuration(config.PreviewProbeBudget)
	if budget <= 0 {
		budget = 20 * time.Second
	}
	deadline := time.Now().Add(budget)
	client := proxy.Client(perFetch)

	const (
		rounds    = 2                       // first pass warms lazy services, second catches them warmed
		warmDelay = 1500 * time.Millisecond // grace period between passes for the warm-up to finish
	)
	imageIdx := -1 // first domain that offered at least an og:image, used when no og:video shows up
outer:
	for round := 0; round < rounds; round++ {
		for i, d := range domains {
			if ctx.Err() != nil || time.Now().After(deadline) {
				break outer
			}
			u, err := swapHost(link, d)
			if err != nil {
				continue
			}
			switch probeEmbed(ctx, client, u) {
			case embedVideo:
				return i, embedVideo // best possible — a real video preview
			case embedImage:
				if imageIdx == -1 {
					imageIdx = i
				}
			}
		}
		if round+1 < rounds {
			select {
			case <-time.After(warmDelay):
			case <-ctx.Done():
				break outer
			}
		}
	}
	if imageIdx >= 0 {
		return imageIdx, embedImage
	}
	return 0, embedNone
}

// probeEmbed fetches rawURL as Telegram would and reports the best og media it exposes.
func probeEmbed(ctx context.Context, client *http.Client, rawURL string) embedKind {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return embedNone
	}
	req.Header.Set("User-Agent", telegramPreviewUA)
	req.Header.Set("Accept", "text/html")
	resp, err := client.Do(req)
	if err != nil {
		return embedNone
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return embedNone
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024)) // og: tags live in <head>
	switch {
	case bytes.Contains(body, []byte("og:video")):
		return embedVideo
	case bytes.Contains(body, []byte("og:image")):
		return embedImage
	default:
		return embedNone
	}
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
