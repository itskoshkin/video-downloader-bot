package links

import (
	"fmt"
	"strings"

	"video-downloader-bot/internal/telegram/constants"
	s "video-downloader-bot/internal/telegram/strings"
	"video-downloader-bot/pkg/ytdlp"
)

const (
	unknown   = "🎬"  // 0
	twitter   = "𝕏"  // 1
	youTube   = "🔴"  // 2
	instagram = "💝"  // 3
	tikTok    = "⬛️" // 4
)

// noinspection SpellCheckingInspection
func getServiceIcon(meta *ytdlp.Metadata, link string) (string, int) {
	switch strings.ToLower(meta.Extractor) {
	case "twitter", "x":
		return twitter, 1
	case "youtube":
		return youTube, 2
	case "instagram":
		return instagram, 3
	case "tiktok":
		return tikTok, 4
	}

	clean := strings.ToLower(DetrackLink(link))

	switch {
	case strings.HasPrefix(clean, "x.com/"), strings.HasPrefix(clean, "twitter.com/"), strings.HasPrefix(clean, "fxtwitter.com/"):
		return twitter, 1
	case strings.HasPrefix(clean, "youtube.com/"), strings.HasPrefix(clean, "youtu.be/"):
		return youTube, 2
	case strings.HasPrefix(clean, "instagram.com/"), strings.HasPrefix(clean, "ddinstagram.com/"), strings.HasPrefix(clean, "kkinstagram.com/"):
		return instagram, 3
	case strings.HasPrefix(clean, "tiktok.com/"), strings.HasPrefix(clean, "vm.tiktok.com/"), strings.HasPrefix(clean, "vt.tiktok.com/"):
		return tikTok, 4
	default:
		return unknown, 0
	}
}

func GetVideoCaption(metadata *ytdlp.Metadata, link string) string {
	link = DetrackLink(link)

	var icon, service = getServiceIcon(metadata, link)
	var caption = fmt.Sprintf("%s ", icon)

repeat:
	caption = fmt.Sprintf("%s ", icon)
	desc := quotedDescription(metadata.Description)
	switch service {
	case 1:
		caption += "Tweet from " + s.Bold(metadata.Author) + " (" + getTwitterUsername(link) + ")" + quotedDescription(trimTwitterLinkInTweet(metadata.Description)) + "\n\n" + link
	case 2:
		author := s.Bold(metadata.Author)
		if metadata.AuthorID != "" {
			author += " (" + metadata.AuthorID + ")"
		}
		if strings.Contains(link, ".com/shorts/") {
			caption += "Short «" + metadata.Title + "» from " + author + desc + "\n\n" + link
		} else {
			caption += "Video «" + metadata.Title + "» from " + author + desc + "\n\n" + link
		}
	case 3:
		caption += "Reel from " + s.Bold(metadata.Author) + " (" + getInstagramUsername(metadata.Title) + ")" + desc + "\n\n" + link
	case 4:
		caption += "TikTok from " + s.Bold(metadata.Author) + desc + "\n\n" + link
	default:
		caption += metadata.Title + " — " + metadata.Author + desc + "\n\n" + link
	}

	if len([]rune(caption)) > constants.MaxMediaCaptionLength {
		overflow := len([]rune(caption)) - constants.MaxMediaCaptionLength
		description := []rune(metadata.Description)
		if overflow >= len(description) {
			metadata.Description = ""
		} else {
			metadata.Description = strings.TrimSpace(string(description[:len(description)-overflow]))
		}
		goto repeat // May my teacher forgive me
	}

	if len([]rune(caption)) > constants.MaxMediaCaptionLength {
		captionRunes := []rune(caption)
		caption = string(captionRunes[:constants.MaxMediaCaptionLength-1]) + "…"
	}

	return caption
}

func getTwitterUsername(s string) string {
	return "@" + strings.TrimPrefix(strings.Split(s, "/status/")[0], "x.com/")
}

func trimTwitterLinkInTweet(s string) string {
	return strings.Split(s, "https://t.co/")[0]
}

func getInstagramUsername(s string) string {
	return "@" + strings.TrimPrefix(s, "Video by ")
}

func quotedDescription(desc string) string {
	desc = strings.TrimSpace(desc)
	if desc == "" {
		return ""
	}

	return "\n\n" + s.Quote(desc)
}
