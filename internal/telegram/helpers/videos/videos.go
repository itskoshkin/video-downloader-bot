package videos

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"

	"video-downloader-bot/internal/config"
	"video-downloader-bot/internal/telegram/constants"
	"video-downloader-bot/pkg/ffmpeg"
	"video-downloader-bot/pkg/ytdlp"
)

func FetchMetadata(ctx context.Context, link string) (*ytdlp.Metadata, error) {
	return ytdlp.FetchMetadata(ctx, link)
}

func Convert(ctx context.Context, inputPath string) (string, error) {
	outputPath := filepath.Join(viper.GetString(config.TelegramBotVideoConvertedFolder), strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))+".mp4")
	_, err := ffmpeg.Convert(ctx, inputPath, outputPath)
	if err != nil {
		return "", err
	}

	limit := min(viper.GetInt(config.TelegramBotMaxFileSizeMB), constants.TelegramMaxFileSizeMB)
	if info, statErr := os.Stat(outputPath); statErr == nil && info.Size() > int64(limit)*1024*1024 {
		return "", fmt.Errorf("converted video is too large (%.1f MB), max is %d MB", float64(info.Size())/1024/1024, limit)
	}

	return outputPath, nil
}
