package proxy

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/viper"

	"video-downloader-bot/internal/config"
)

// Address returns the configured SOCKS5 proxy URL (byedpi/tunnel) from config/env, or "" for direct.
func Address() string { return strings.TrimSpace(viper.GetString(config.ProxySocks5)) }

// Client returns an *http.Client with the given timeout, routed through the SOCKS5 proxy when one is
// configured, or a direct client otherwise. net/http dials socks5:// (and socks5h://) proxies natively.
func Client(timeout time.Duration) *http.Client {
	c := &http.Client{Timeout: timeout}
	if addr := Address(); addr != "" {
		if u, err := url.Parse(addr); err == nil {
			c.Transport = &http.Transport{Proxy: http.ProxyURL(u)}
		}
	}
	return c
}
