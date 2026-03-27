package links

import (
	"net/url"
	"strings"
)

func DetrackLink(link string) string {
	link = strings.TrimSpace(link)
	if link == "" {
		return ""
	}

	parsed, err := url.Parse(link)
	if err != nil {
		return link
	}

	host := normalizeHost(parsed.Hostname())
	path := trimTrailingSlash(parsed.EscapedPath())

	switch host {
	case "youtu.be":
		videoID := strings.TrimPrefix(path, "/")
		if videoID == "" {
			return "youtube.com"
		}
		return "youtube.com/watch?v=" + videoID
	case "youtube.com":
		if path == "/watch" {
			videoID := parsed.Query().Get("v")
			if videoID == "" {
				return "youtube.com/watch"
			}
			return "youtube.com/watch?v=" + videoID
		}
		if path == "" {
			return "youtube.com"
		}
		return "youtube.com" + path
	default:
		if path == "" {
			return host
		}
		return host + path
	}
}
