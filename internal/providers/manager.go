package providers

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/viper"

	"video-downloader-bot/internal/config"
	"video-downloader-bot/internal/logger"
	"video-downloader-bot/internal/utils/links"
)

// Manager holds every available provider and runs the per-platform chain from config.
type Manager struct {
	byName map[string]Downloader
}

func NewManager(providers ...Downloader) *Manager {
	m := &Manager{byName: make(map[string]Downloader, len(providers))}
	for _, p := range providers {
		m.byName[p.Name()] = p
	}
	return m
}

// Download walks the configured chain for the link's platform and returns the first
// success. Inapplicable and not-yet-implemented providers are skipped; any other error
// is remembered and returned only if nothing downstream succeeds.
func (m *Manager) Download(ctx context.Context, link string) (*Result, error) {
	chain := m.chainFor(link)
	if len(chain) == 0 {
		return nil, fmt.Errorf("providers: no chain configured for %q", link)
	}

	var lastErr error
	for _, name := range chain {
		p, ok := m.byName[name]
		if !ok {
			logger.WarnWithID(ctx, "providers: unknown provider %q in chain, skipping", name)
			continue
		}
		if !p.Applicable(link) {
			continue
		}

		res, err := p.Download(ctx, link)
		if err == nil {
			logger.DebugWithID(ctx, "providers: %q handled the link", name)
			return res, nil
		}
		if errors.Is(err, ErrNotApplicable) || errors.Is(err, ErrNotImplemented) {
			logger.DebugWithID(ctx, "providers: %q skipped (%v)", name, err)
			continue
		}

		logger.WarnWithID(ctx, "providers: %q failed, falling back: %v", name, err)
		lastErr = err
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("providers: chain exhausted for %q", link)
}

// chainFor resolves the ordered provider names for the link's platform, falling back to the default chain.
func (m *Manager) chainFor(link string) []string {
	if platform := links.Platform(link); platform != "" {
		if chain := viper.GetStringSlice(config.ProvidersChainKey(platform)); len(chain) > 0 {
			return chain
		}
	}
	return viper.GetStringSlice(config.ProvidersChainDefault)
}
