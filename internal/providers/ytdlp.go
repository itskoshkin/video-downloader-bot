package providers

import (
	"context"
	"path/filepath"

	"github.com/spf13/viper"

	"video-downloader-bot/internal/config"
	"video-downloader-bot/internal/utils/links"
	"video-downloader-bot/pkg/ytdlp"
)

// ytDlp wraps the existing yt-dlp download logic as the primary, general-purpose provider.
type ytDlp struct{}

func NewYtDlp() Downloader { return &ytDlp{} }

func (*ytDlp) Name() string { return "yt-dlp" }

// Applicable: yt-dlp is the general-purpose provider — it handles every supported link.
func (*ytDlp) Applicable(link string) bool { return links.IsSupportedLink(link) }

func (*ytDlp) Download(ctx context.Context, link string) (*Result, error) {
	name, err := ytdlp.DownloadVideo(ctx, link)
	if err != nil {
		return nil, err
	}
	full := filepath.Join(viper.GetString(config.TelegramBotVideoDownloadFolder), name)
	return &Result{Kind: KindFile, FilePath: full}, nil
}
