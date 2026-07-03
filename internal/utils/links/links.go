package links

import (
	"net/url"
	"regexp"
	"strings"
)

//goland:noinspection RegExpUnnecessaryNonCapturingGroup
var linkPattern = regexp.MustCompile(`(?i)(?:^|[\s<(\["'])((?:https?://[^\s<>()\[\]{}"']+)|(?:(?:www\.)?(?:x\.com|twitter\.com|youtube\.com|youtu\.be|instagram\.com|tiktok\.com)|(?:(?:vm|vt|m)\.tiktok\.com))[^\s<>()\[\]{}"']*)`)

func IsSupportedLink(link string) bool {
	link = strings.TrimPrefix(link, "https://")

	//noinspection HttpUrlsUsage
	link = strings.TrimPrefix(link, "http://")

	link = strings.TrimPrefix(link, "www.")

	hostAndPath := link
	if i := strings.IndexAny(hostAndPath, "?#"); i >= 0 {
		hostAndPath = hostAndPath[:i]
	}
	hostAndPath = strings.TrimRight(hostAndPath, "/")

	parts := strings.SplitN(hostAndPath, "/", 2)
	host := parts[0]
	path := ""
	if len(parts) == 2 {
		path = "/" + parts[1]
	}

	switch host {
	case "youtube.com":
		return path == "/watch" || strings.HasPrefix(path, "/shorts/")
	case "youtu.be":
		return path != ""
	case "instagram.com", "ddinstagram.com", "kkinstagram.com":
		return strings.HasPrefix(path, "/reel/")
	case "x.com", "twitter.com", "fxtwitter.com":
		return strings.Contains(path, "/status/")
	case "tiktok.com", "vxtiktok.com":
		return strings.Contains(path, "/video/") || strings.HasPrefix(path, "/t/")
	case "vm.tiktok.com", "vt.tiktok.com", "vm.vxtiktok.com", "vt.vxtiktok.com":
		return path != ""
	default:
		return false
	}
}

// Platform returns the canonical platform name for a supported link ("x", "youtube",
// "instagram", "tiktok"), or "" if unknown. Used to pick the provider chain from config.
func Platform(link string) string {
	clean := strings.ToLower(DetrackLink(link))
	switch {
	case strings.HasPrefix(clean, "x.com/"), strings.HasPrefix(clean, "twitter.com/"), strings.HasPrefix(clean, "fxtwitter.com/"):
		return "x"
	case strings.HasPrefix(clean, "youtube.com/"), strings.HasPrefix(clean, "youtu.be/"):
		return "youtube"
	case strings.HasPrefix(clean, "instagram.com/"), strings.HasPrefix(clean, "ddinstagram.com/"), strings.HasPrefix(clean, "kkinstagram.com/"):
		return "instagram"
	case strings.HasPrefix(clean, "tiktok.com/"), strings.HasPrefix(clean, "vm.tiktok.com/"), strings.HasPrefix(clean, "vt.tiktok.com/"):
		return "tiktok"
	default:
		return ""
	}
}

func HasLink(s string) (string, bool) {
	if IsOnlyLink(s) {
		return normalizeLink(s), true
	}

	if ContainsLink(s) {
		s = ExtractLink(s)
		if s != "" {
			return s, true
		}
	}

	return "", false
}

func IsOnlyLink(message string) bool {
	message = strings.TrimSpace(message)
	if message == "" {
		return false
	}

	return ExtractLink(message) == message
}

func ContainsLink(message string) bool {
	return ExtractLink(message) != ""
}

func ExtractLink(message string) string {
	matches := linkPattern.FindStringSubmatch(message)
	if len(matches) < 2 {
		return ""
	}

	return normalizeLink(trimLinkCandidate(matches[1]))
}

func normalizeLink(link string) string {
	link = strings.TrimSpace(link)
	if link == "" {
		return ""
	}

	if !strings.HasPrefix(link, "http://") && !strings.HasPrefix(link, "https://") {
		link = "https://" + link
	}

	u, err := url.Parse(link)
	if err != nil {
		return link
	}

	host := normalizeHost(u.Hostname())
	path := trimTrailingSlash(u.EscapedPath())

	switch host {

	case "youtu.be":
		videoID := strings.TrimPrefix(path, "/")
		if videoID == "" {
			return "https://youtube.com"
		}
		return "https://youtube.com/watch?v=" + videoID

	case "youtube.com":
		if path == "/watch" {
			videoID := u.Query().Get("v")
			if videoID == "" {
				return "https://youtube.com/watch"
			}
			return "https://youtube.com/watch?v=" + videoID
		}
		if path == "" {
			return "https://youtube.com"
		}
		return "https://youtube.com" + path

	case "instagram.com":
		if path == "" {
			return "https://instagram.com"
		}
		return "https://instagram.com" + path

	case "x.com":
		if path == "" {
			return "https://x.com"
		}
		return "https://x.com" + path

	case "tiktok.com":
		if path == "" {
			return "https://tiktok.com"
		}
		return "https://tiktok.com" + path

	case "vm.tiktok.com", "vt.tiktok.com":
		if path == "" {
			return "https://" + host
		}
		return "https://" + host + path

	default:
		if path == "" {
			return "https://" + host
		}
		return "https://" + host + path

	}
}

func normalizeHost(host string) string {
	host = strings.ToLower(strings.TrimSpace(host))
	host = strings.TrimPrefix(host, "www.")

	switch host {
	case "twitter.com", "fxtwitter.com":
		return "x.com"
	case "m.youtube.com":
		return "youtube.com"
	case "ddinstagram.com", "kkinstagram.com":
		return "instagram.com"
	case "m.tiktok.com":
		return "tiktok.com"
	default:
		return host
	}
}

func trimTrailingSlash(path string) string {
	if path == "" || path == "/" {
		return ""
	}

	return strings.TrimRight(path, "/")
}

func trimLinkCandidate(link string) string {
	link = strings.TrimSpace(link)
	link = strings.TrimLeft(link, `"'([<{`)
	link = strings.TrimRight(link, `"'.,!;:)]}>`)

	return link
}
